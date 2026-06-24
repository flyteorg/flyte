---
name: flyte-deploy
description: Use when deploying a Flyte v2 (flyte-binary / flyte2) cluster on AWS from scratch — provisions EKS + S3 + RDS PostgreSQL + AWS Load Balancer Controller, then helm-installs the flyte-binary chart behind an ALB, with optional TLS and Okta SSO. Trigger words: "deploy flyte", "flyte v2 on AWS", "flyte EKS".
---

# Deploying Flyte v2 on AWS (EKS + RDS + S3 + ALB)

Flyte v2 ships as a single unified binary (`flyte-binary-v2`) plus a separate console
image. One HTTP ingress serves the console (`/v2`), the `flyteidl2.*` Connect API, and
auth-discovery — there is no separate gRPC port. You scale it vertically.

The chart does NOT provision infrastructure. Stand up four things first:
**EKS cluster, S3 bucket, PostgreSQL (RDS), and (for external access) an ingress
controller.** This skill does all four with `eksctl` + `aws` + `helm`, then installs the
`flyte-binary` chart (the v2 chart; defaults to `flyte-binary-v2` + `flyteconsole-v2`).

Chart source: the `charts/flyte-binary` directory in the flyte repo, or `helm repo add
flyteorg https://flyteorg.github.io/flyte` once published. Official deployment docs:
https://flyte.org (Deployment → Flyte deployment). Validated end-to-end on EKS.

> Replace every placeholder in angle brackets and the example hostnames/IDs with your own.

## Prerequisites & decisions

- CLIs: `aws` v2, **`eksctl` ≥ 0.227** (older caps out at k8s 1.29 — see gotcha), `kubectl`, `helm`, `jq`.
- Admin (or EKS+RDS+IAM+S3+EC2) creds. STS/SSO works — export the 3 env vars + region.
- Decide: region, name prefix, **RDS vs in-cluster Postgres**, and **exposure**:
  ALB+TLS needs a Route53 zone + ACM cert; **ALB HTTP-only needs neither** (reached at
  the auto `*.elb.amazonaws.com` name) — the simplest default when you own no domain.

```bash
export AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... AWS_SESSION_TOKEN=...
export AWS_DEFAULT_REGION=us-west-2
PREFIX=flyte-v2; REGION=us-west-2; CLUSTER=$PREFIX
ACCT=$(aws sts get-caller-identity --query Account --output text)   # confirm the RIGHT account first
# Check for an existing domain/cert (empty => go ALB HTTP-only):
aws route53 list-hosted-zones --query 'HostedZones[].Name' --output text
aws acm list-certificates --region $REGION --query 'CertificateSummaryList[].DomainName' --output text
```

## Step 1 — EKS cluster (eksctl)

`cluster.yaml` — `iam.withOIDC: true` is what makes IRSA possible:

```yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata: { name: flyte-v2, region: us-west-2, version: "1.33" }
iam: { withOIDC: true }
managedNodeGroups:
  - name: ng-default
    instanceType: m5.large
    desiredCapacity: 2
    minSize: 2
    maxSize: 3
    volumeSize: 50
    iam: { withAddonPolicies: { ebs: true } }
addons: [{name: vpc-cni},{name: coredns},{name: kube-proxy},{name: aws-ebs-csi-driver}]
```

```bash
eksctl create cluster -f cluster.yaml     # ~15-20 min; writes kubeconfig + sets context
kubectl get nodes                          # expect Ready
```

The VPC + private subnets exist within ~2 min (before the control plane finishes), so you
can start RDS (step 3) in parallel.

## Step 2 — S3 bucket + IRSA role

```bash
BUCKET=$PREFIX-data-$ACCT       # account-id suffix => globally unique
aws s3api create-bucket --bucket $BUCKET --region $REGION \
  --create-bucket-configuration LocationConstraint=$REGION
aws s3api put-public-access-block --bucket $BUCKET --public-access-block-configuration \
  BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
aws s3api put-bucket-encryption --bucket $BUCKET --server-side-encryption-configuration \
  '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'

# Scoped S3 policy: ListBucket on the bucket, Get/Put/Delete on its objects.
cat > s3-policy.json <<EOF
{"Version":"2012-10-17","Statement":[
 {"Effect":"Allow","Action":["s3:ListBucket"],"Resource":"arn:aws:s3:::$BUCKET"},
 {"Effect":"Allow","Action":["s3:GetObject","s3:PutObject","s3:DeleteObject"],"Resource":"arn:aws:s3:::$BUCKET/*"}]}
EOF
POLICY_ARN=$(aws iam create-policy --policy-name $PREFIX-s3-access \
  --policy-document file://s3-policy.json --query Policy.Arn --output text)

# --role-only: create the IAM role with OIDC trust for system:serviceaccount:flyte:flyte,
# but NOT the k8s SA (the chart creates+annotates it). Works before the ns exists.
eksctl create iamserviceaccount --cluster $CLUSTER --region $REGION \
  --namespace flyte --name flyte --role-name $PREFIX-irsa \
  --attach-policy-arn "$POLICY_ARN" --role-only --approve
IRSA_ARN=$(aws iam get-role --role-name $PREFIX-irsa --query Role.Arn --output text)
```

## Step 3 — RDS PostgreSQL  (can run in parallel with step 1)

```bash
VPC=$(aws ec2 describe-vpcs --region $REGION \
  --filters "Name=tag:alpha.eksctl.io/cluster-name,Values=$CLUSTER" --query 'Vpcs[0].VpcId' --output text)
# Private subnets (internal-elb role tag):
SUBNETS=$(aws ec2 describe-subnets --region $REGION --filters "Name=vpc-id,Values=$VPC" \
  "Name=tag:kubernetes.io/role/internal-elb,Values=1" --query 'Subnets[].SubnetId' --output text)
# CRITICAL: source SG must be the EKS-managed cluster SG actually on the NODES,
# NOT ClusterSharedNodeSecurityGroup. Pod egress uses the node primary-ENI SG.
NODESG=$(aws ec2 describe-instances --region $REGION \
  --filters "Name=tag:eks:cluster-name,Values=$CLUSTER" \
  --query 'Reservations[0].Instances[0].SecurityGroups[?contains(GroupName,`eks-cluster-sg`)].GroupId' --output text)
# (If nodes aren't up yet, grab it later and add the rule then — that's all the DB needs.)

aws rds create-db-subnet-group --region $REGION --db-subnet-group-name $PREFIX-db-subnets \
  --db-subnet-group-description "Flyte v2 private DB subnets" --subnet-ids $SUBNETS
RDSSG=$(aws ec2 create-security-group --region $REGION --group-name $PREFIX-rds-sg \
  --description "Flyte v2 RDS 5432 from cluster nodes" --vpc-id $VPC --query GroupId --output text)
aws ec2 authorize-security-group-ingress --region $REGION --group-id $RDSSG \
  --protocol tcp --port 5432 --source-group $NODESG
DBPW=$(LC_ALL=C tr -dc 'A-Za-z0-9' </dev/urandom | head -c 28)   # store it somewhere safe
aws rds create-db-instance --region $REGION --db-instance-identifier $PREFIX-db \
  --engine postgres --db-instance-class db.t3.micro --allocated-storage 20 --storage-type gp3 \
  --master-username flyte --master-user-password "$DBPW" --db-name flyte \
  --vpc-security-group-ids $RDSSG --db-subnet-group-name $PREFIX-db-subnets \
  --no-publicly-accessible --backup-retention-period 1
# Endpoint (when status=available):
RDS_HOST=$(aws rds describe-db-instances --region $REGION --db-instance-identifier $PREFIX-db \
  --query 'DBInstances[0].Endpoint.Address' --output text)
```

## Step 4 — AWS Load Balancer Controller (for ALB ingress)

```bash
# Use the policy matching the controller version the chart installs — currently v3.x.
curl -sL https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v3.4.0/docs/install/iam_policy.json -o alb-iam-policy.json
ALB_POLICY_ARN=$(aws iam create-policy --policy-name AWSLoadBalancerControllerIAMPolicy \
  --policy-document file://alb-iam-policy.json --query Policy.Arn --output text)
eksctl create iamserviceaccount --cluster $CLUSTER --region $REGION \
  --namespace kube-system --name aws-load-balancer-controller \
  --role-name $PREFIX-alb-controller --attach-policy-arn "$ALB_POLICY_ARN" --approve
helm repo add eks https://aws.github.io/eks-charts && helm repo update eks
helm upgrade --install aws-load-balancer-controller eks/aws-load-balancer-controller -n kube-system \
  --set clusterName=$CLUSTER --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller --set region=$REGION --set vpcId=$VPC
kubectl -n kube-system rollout status deploy/aws-load-balancer-controller
```

If the controller image is newer than the policy you fetched, you'll see `AccessDenied`
on actions like `elasticloadbalancing:DescribeListenerAttributes`. Fix WITHOUT reinstalling:
```bash
curl -sL .../aws-load-balancer-controller/v<INSTALLED>/docs/install/iam_policy.json -o p.json
aws iam create-policy-version --policy-arn $ALB_POLICY_ARN --policy-document file://p.json --set-as-default
```
(Check version: `kubectl -n kube-system get deploy aws-load-balancer-controller -o jsonpath='{..image}'`.)

## Step 5 — helm install flyte-binary

`values-eks.yaml` (ALB HTTP-only variant). Note this chart uses `metadataContainer`
(no `userDataContainer`) and its run output prefix defaults to a nonexistent
`s3://flyte-data` — override `storagePrefix` to your bucket:

```yaml
fullnameOverride: flyte
flyte-core-components:
  runs: { storagePrefix: "s3://BUCKET" }   # under `runs`, NOT `runs.server` (else ignored)
configuration:
  database:
    postgres:
      host: RDS_HOST
      port: 5432
      dbname: flyte
      username: flyte
      password: "DBPW"
      options: "sslmode=require"
  storage:
    metadataContainer: BUCKET
    provider: s3
    providerConfig: { s3: { region: us-west-2, authType: iam } }
  inline: { executor: { defaultK8sServiceAccount: flyte } }   # task pods inherit S3 via IRSA
serviceAccount:
  create: true
  name: flyte
  annotations: { eks.amazonaws.com/role-arn: IRSA_ARN }
ingress:
  create: true
  host: ""                       # empty => rule matches any host => reach by ALB DNS name
  ingressClassName: alb
  httpAnnotations:
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTP": 80}]'
    alb.ingress.kubernetes.io/healthcheck-path: /healthz      # binary serves /healthz on :8090
    alb.ingress.kubernetes.io/healthcheck-port: "8090"
```

For **TLS**: add `certificate-arn`, `listen-ports: '[{"HTTP":80},{"HTTPS":443}]'`,
`ssl-redirect: "443"`, and set `ingress.host` to the cert hostname + a Route53 record.
See the TLS section below — works even when DNS lives in a different AWS account.

```bash
helm install flyte ./charts/flyte-binary -n flyte --create-namespace -f values-eks.yaml --dry-run  # check
helm install flyte ./charts/flyte-binary -n flyte --create-namespace -f values-eks.yaml
kubectl -n flyte get pods   # flyte stuck Init:0/1 => wait-for-db can't reach RDS (see gotchas)
```

## Step 6 — Verify

```bash
ALB=$(kubectl -n flyte get ingress flyte-http -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
curl -s -X POST "http://$ALB/flyteidl2.project.ProjectService/ListProjects" \
  -H 'Content-Type: application/json' -d '{}'        # => JSON with the seeded flytesnacks project
curl -s -o /dev/null -w "%{http_code}\n" "http://$ALB/v2"   # => 200 (console)
```
ALB takes ~2-3 min after the address appears to pass health checks and serve 200.

## TLS (ACM + ALB), including cross-account DNS

The **cert and ALB must be in the Flyte account + the ALB's region**; the **DNS zone can
live in another account** — you just add two records there. No domain transfer/delegation.
TLS is a prerequisite for browser SSO (ALB `authenticate-oidc` only runs on HTTPS listeners;
IdPs reject non-`https` redirect URIs for non-localhost hosts).

```bash
HOST=flyte.example.com; ZONE=<YOUR_ROUTE53_ZONE_ID>   # zone is in whichever account owns DNS
# 1. Flyte account: request a DNS-validated cert (same region as the ALB)
CERT=$(aws acm request-certificate --region $REGION --domain-name $HOST \
  --validation-method DNS --query CertificateArn --output text)
aws acm describe-certificate --region $REGION --certificate-arn $CERT \
  --query 'Certificate.DomainValidationOptions[0].ResourceRecord'    # -> {Name,Type,Value}
# 2. DNS account: UPSERT that validation CNAME into the zone (change-resource-record-sets)
# 3. Flyte account: wait until status == ISSUED (aws acm describe-certificate ...)
# 4. Flyte account: helm upgrade with ingress.host=$HOST and these httpAnnotations:
#      listen-ports: '[{"HTTP": 80}, {"HTTPS": 443}]'
#      ssl-redirect: "443"
#      certificate-arn: <CERT>
# 5. DNS account: UPSERT a CNAME  $HOST -> <ALB DNS name>  (subdomain => CNAME is fine; no
#    cross-account alias-target dance needed)
# 6. Verify: curl https://$HOST/v2 -> 200; http -> 301; openssl s_client shows CN=$HOST
```

Two AWS credential sets may be in play (Flyte acct + DNS acct) — keep them in separate env
files and `source` the right one per command. STS/SSO session tokens expire mid-deploy;
when a call returns `ExpiredTokenException`, refresh that account's creds (kubectl/helm to
the cluster also need the Flyte account's creds, via `aws eks get-token`).

## ALB edge SSO (Okta / OIDC)

Gates the console at the load balancer via ALB `authenticate-oidc` — the binary is
unchanged. **Requires HTTPS** (see TLS above). Tradeoff: the action applies to ALL paths on
the ingress, so the browser console works (same-origin API calls carry the ALB session
cookie) but **CLI/SDK clients get 302'd** — add a higher-precedence `ingress.apiJwtIngress`
(Bearer-match, no auth action) to let token clients bypass (see CLI Bearer-bypass below).

1. Add the redirect URI **`https://<host>/oauth2/idpresponse`** to the OIDC app (fixed ALB
   callback path) — login fails without it.
2. Create the OIDC Secret in the **ingress namespace** (keys exactly `clientID`/`clientSecret`):
   ```bash
   kubectl -n flyte create secret generic flyte-console-oidc \
     --from-literal=clientID=<oidc-client-id> --from-literal=clientSecret=<oidc-secret>
   ```
3. Add these `ingress.httpAnnotations` and `helm upgrade`:
   ```yaml
   alb.ingress.kubernetes.io/auth-type: oidc
   alb.ingress.kubernetes.io/auth-scope: openid email profile   # profile => given/family name in x-amzn-oidc-data
   alb.ingress.kubernetes.io/auth-on-unauthenticated-request: authenticate
   alb.ingress.kubernetes.io/auth-session-timeout: "604800"
   alb.ingress.kubernetes.io/auth-idp-oidc: '{"issuer":"https://<idp>/oauth2/default","authorizationEndpoint":".../v1/authorize","tokenEndpoint":".../v1/token","userInfoEndpoint":".../v1/userinfo","secretName":"flyte-console-oidc"}'
   ```
4. **Grant the controller RBAC to read that Secret** (REQUIRED — otherwise the rule silently
   stays plain `forward` and you keep getting HTTP 200 instead of 302). The controller SA
   `kube-system:aws-load-balancer-controller` has no secret access in app namespaces by default:
   ```bash
   kubectl -n flyte create role alb-oidc-secret-reader --verb=get,list,watch --resource=secrets
   kubectl -n flyte create rolebinding alb-oidc-secret-reader --role=alb-oidc-secret-reader \
     --serviceaccount=kube-system:aws-load-balancer-controller
   ```
   Symptom if missing: controller logs `secrets "flyte-console-oidc" is forbidden`.
   - **Okta issuer host:** use the **non-admin** org domain (`https://<org>.okta.com/oauth2/default`),
     NOT the `-admin` console host — tokens carry the non-admin host as `iss`, so jwt-validation
     fails if you use `-admin`. Confirm via `…/oauth2/default/.well-known/openid-configuration`.
     Switching IdPs is pure config (Secret + the issuer refs + `flyteClient.clientId`); no ALB/DNS churn.
5. Verify: `curl -s -o /dev/null -w '%{http_code} %{redirect_url}' https://<host>/v2`
   → `302 https://<idp>/oauth2/default/v1/authorize?client_id=...&redirect_uri=https://<host>/oauth2/idpresponse`.
6. **IdP-side (can't be fixed from the cluster):** the user must be **assigned to the app**,
   and (on Okta) the `default` auth server's **Access Policy** must permit the app + the
   requested scopes (`openid email profile`). Okta error *"Bad Request — Policy evaluation
   failed"* after the redirect = the access-policy rule is missing the app or restricts scopes
   → Security → API → Authorization Servers → default → Access Policies, allow the app with
   "Any scopes" (or add `email`/`profile`).

### CLI Bearer-bypass (dual-auth: keep CLI/SDK working alongside console SSO)

Edge SSO alone 302s CLI clients. To let token clients through, add two more ingresses in
the SAME ALB group plus `authMetadata` so the CLI knows to fetch a token. Three ingresses,
ordered by `group.order` (lower = higher precedence), all sharing one `group.name`:

| Ingress | order | matches | auth |
|---|---|---|---|
| `wellknownIngress` | -150 | `/.well-known/*`, `AuthMetadataService` | none (discovery before token) |
| `apiJwtIngress` | -140 | `Authorization: Bearer*` (via `conditions.<fullname>-http`) | ALB `jwt-validation` vs IdP JWKS |
| http (main) | -100 | everything else | `authenticate-oidc` (cookie) |

1. Tell the binary to advertise the IdP + the PKCE CLI client — under **`runs`, NOT `runs.server`**:
   ```yaml
   flyte-core-components:
     runs:
       storagePrefix: "s3://<bucket>"        # also belongs under runs, not runs.server
       authMetadata:
         externalAuthServerBaseUrl: "https://<idp>/oauth2/default"
         flyteClient: { clientId: <native-PKCE-app>, redirectUri: http://localhost:53593/callback, scopes: [openid, profile, offline_access] }
   ```
2. Add `group.name: <group>` + `group.order: "-100"` to the main `httpAnnotations`, and add
   `ingress.apiJwtIngress` (enabled, order -140, `conditions.<fullname>-http` matching
   `Bearer*`, `jwt-validation` pointing at the IdP JWKS endpoint) and `ingress.wellknownIngress`
   (enabled, order -150, no auth). All three need cert-arn + listen-ports + ssl-redirect.
3. `helm upgrade`. **Adding group.name recreates the ALB under a new name** (`k8s-<group>-*`)
   and deletes the old standalone one — **re-point the DNS CNAME to the new ALB DNS name.**
4. Verify (use `curl --connect-to host:443:<alb>:443` before DNS propagates):
   - `POST .../AuthMetadataService/GetOAuth2Metadata` (with `Content-Type: application/json`) → 200
   - API `+ Authorization: Bearer fake` → **401, not 302** (matched the JWT ingress, validated)
   - `/v2` and API without Bearer → 302 (cookie path)

Gotchas: (a) `authMetadata`/`storagePrefix` go under `runs`, not `runs.server` — misplaced,
they're silently ignored and `GetOAuth2Metadata` returns `unimplemented`. (b) config-only helm
changes may not roll the pod — `kubectl rollout restart deploy/flyte` to be sure. (c) the
`conditions.*` key must match the rendered backend service name (`<fullname>-http`).

## Teardown

```bash
helm uninstall flyte -n flyte          # deletes the ingress => controller removes the ALB
helm uninstall aws-load-balancer-controller -n kube-system
aws rds delete-db-instance --region $REGION --db-instance-identifier $PREFIX-db --skip-final-snapshot --delete-automated-backups
aws rds delete-db-subnet-group --region $REGION --db-subnet-group-name $PREFIX-db-subnets
aws s3 rb s3://$PREFIX-data-$ACCT --force
eksctl delete cluster -f cluster.yaml   # tears down VPC, nodegroup, OIDC, IRSA stacks
# Delete the standalone IAM policies (detach first if needed):
aws iam delete-policy --policy-arn arn:aws:iam::$ACCT:policy/$PREFIX-s3-access
aws iam delete-policy --policy-arn arn:aws:iam::$ACCT:policy/AWSLoadBalancerControllerIAMPolicy
```

## Gotchas (each one bit during a real run)

1. **eksctl too old → "unsupported Kubernetes version".** eksctl 0.175 only offers up to
   1.29, but EKS has dropped 1.29 from standard support → CFN `ControlPlane` fails ~30s in
   and rolls back. Use eksctl ≥ 0.227 (defaults to a current version); pin a supported one
   (1.33 worked). After a failed create, delete the `ROLLBACK_COMPLETE` stack before retrying.
2. **RDS unreachable: wrong source SG.** The pod stays `Init:0/1` (`wait-for-db ... no
   response`). EKS managed-nodegroup nodes run with the **EKS-managed cluster SG**
   (`eks-cluster-sg-<cluster>-*`), NOT `ClusterSharedNodeSecurityGroup`. Pod egress (VPC CNI
   secondary IPs on the primary ENI) uses the node-ENI SG. Authorize 5432 on the RDS SG from
   the actual node SG (`describe-instances ... SecurityGroups`), not the shared one. Init
   container retries on its own once the rule lands.
3. **ALB controller IAM lag.** The eks chart installs the latest controller (v3.x), which
   needs newer IAM actions (e.g. `DescribeListenerAttributes`) than older policy JSON.
   Match `iam_policy.json` to the installed controller version (create-policy-version
   --set-as-default; no reinstall needed).
4. **Default storagePrefix is fake.** `flyte-core-components.runs.storagePrefix` (under `runs`,
   NOT `runs.server`) defaults to `s3://flyte-data` — override to your real bucket or run I/O
   fails. Misplaced under `runs.server` it's silently ignored.
5. **ALB by DNS name:** leave `ingress.host: ""` so the rule matches any host; the binary
   serves `/healthz` on `:8090` for the ALB health check. Add ACM cert + Route53 for TLS.
6. Postgres default major from RDS is fine (chart needs ≥12).
7. **Task pods loop/recreate every ~75s — missing control-plane callback env vars.** A run's
   task pod (image `ghcr.io/flyteorg/flyte:py3.x-vX`) calls back to the backend to enqueue
   child actions / watch state. Without config it uses the **devbox default
   `host.docker.internal:8090`** → `dns error: Name or service not known` → retries exhaust →
   controller recreates the pod, forever. Recent `flyte-binary` chart versions inject these by
   default; on older charts add them via `configuration.inline.plugins.k8s.default-env-vars`:
   ```yaml
   configuration:
     inline:
       plugins:
         k8s:
           default-env-vars:
             - _U_EP_OVERRIDE: flyte-http.flyte:8090   # in-cluster HTTP svc = <fullname>-http.<ns>:8090
             - _U_INSECURE: "true"                     # svc is plain HTTP on :8090; without this the
                                                       # SDK uses https:// → "received corrupt message
                                                       # of type InvalidContentType"
             - _U_USE_ACTIONS: "1"                     # enable the QueueService/actions path
   ```
   Verify a task pod: `kubectl -n flyte get pod <run>-a0-0 -o jsonpath='{..env[*].name}'` shows
   `_U_EP_OVERRIDE`, and its logs no longer mention `host.docker.internal` or `InvalidContentType`.
