# SSO at the ALB for the Flyte 2 console (`/v2`)

This guide shows how to put **OIDC single sign-on in front of the Flyte 2 console**
(`/v2`) at the **AWS ALB / ingress layer**, so that hitting
`https://<your-host>/v2` challenges the user to log in (Okta, Google, …) *before*
the request ever reaches the console — with **no change to the Flyte binary**.

It works with any OIDC provider. The worked example is the `flyte-development`
cluster (`development.uniondemo.run`) reusing the v1 flyte console's Okta client.

---

## What you get

- ALB intercepts unauthenticated requests to `/v2`, redirects to your IdP, and only
  forwards to the console after a valid login. ALB manages the session cookie.
- **One ingress for console + APIs.** Clients use buf Connect over HTTP, so there is
  no separate gRPC ingress: the `-http` ingress serves `/v2`, `/v2/*`, and the Connect
  API paths. The auth annotations go on `httpAnnotations`, so they apply to every rule
  on that ingress. SDK/machine clients that send `Authorization: Bearer` are matched
  first by a higher-precedence, separately-managed JWT ingress and bypass the cookie flow.

```
Browser ──GET /v2──▶ ALB ──(no session)──▶ 302 ▶ IdP login
                       ▲                              │
                       └──── /oauth2/idpresponse ◀────┘  (ALB swaps code→token,
                              sets session cookie, forwards to flyte2-console)
```

## How it works

ALB's native `authenticate-oidc` action is configured entirely through
**annotations** on the ingress. The callback path is **fixed** at
`/oauth2/idpresponse` on whatever host the user browsed to — ALB derives the
`redirect_uri` from the request's `Host` header. You cannot point it at a different
host; the callback must come back to the same ALB that started the session.

ALB reads the OIDC **client ID and secret from a Kubernetes Secret** (keys must be
exactly `clientID` and `clientSecret`) in the **ingress's namespace**.

---

## Prerequisites

- **AWS Load Balancer Controller** managing the ingress (`ingressClassName: alb`).
- An **HTTPS listener** with an ACM cert covering your host
  (`alb.ingress.kubernetes.io/certificate-arn`, `listen-ports` includes `443`).
  OIDC auth only applies to **HTTPS** listener rules.
- An **OIDC application** at your IdP (confidential client, Authorization Code flow)
  with a client ID + secret.

---

## Setup

### 1. Configure the OIDC app at your IdP

On the OIDC application:

- **Sign-in / redirect URI** — add exactly (note the path, not just `/`):
  ```
  https://<your-host>/oauth2/idpresponse
  ```
- **Grant type**: Authorization Code; **scopes**: at least `openid email`.
- **Assign** the users/groups who should be allowed into the console.

You need the app's **client ID** and **client secret**.

> **Endpoint shapes** (you'll need issuer + authorize/token/userinfo URLs below):
>
> | | Okta (custom auth server) | Google |
> |---|---|---|
> | issuer | `https://<domain>/oauth2/<id>` | `https://accounts.google.com` |
> | authorize | `…/oauth2/<id>/v1/authorize` | `https://accounts.google.com/o/oauth2/v2/auth` |
> | token | `…/oauth2/<id>/v1/token` | `https://oauth2.googleapis.com/token` |
> | userinfo | `…/oauth2/<id>/v1/userinfo` | `https://openidconnect.googleapis.com/v1/userinfo` |
>
> For Okta, the easiest source of truth is the discovery doc:
> `https://<domain>/oauth2/<id>/.well-known/openid-configuration`.

### 2. Create the Kubernetes Secret

In the **same namespace as the ingress** (here `flyte`), with keys **exactly**
`clientID` / `clientSecret`:

```bash
kubectl create secret generic flyte2-console-oidc -n flyte \
  --from-literal=clientID='<client-id>' \
  --from-literal=clientSecret='<client-secret>'
```

> **Reusing the v1 client.** If v1 flyte-binary already has a working OIDC client on
> this cluster, you can reuse it instead of creating a new app. The v1 secret lives in
> `flyte-binary-client-secrets-external-secret` (key `oidc_client_secret`) and the
> client ID is in the v1 release values (`configuration.auth.oidc.clientId`):
> ```bash
> SEC=$(kubectl get secret flyte-binary-client-secrets-external-secret -n flyte \
>   -o jsonpath='{.data.oidc_client_secret}' | base64 -d)
> kubectl create secret generic flyte2-console-oidc -n flyte \
>   --from-literal=clientID='0oakkheteNjCMERst5d6' \
>   --from-literal=clientSecret="$SEC" --dry-run=client -o yaml | kubectl apply -f -
> ```
> You still must add `https://<host>/oauth2/idpresponse` to that app's redirect URIs.

### 3. Grant the LB controller permission to read the Secret

The AWS Load Balancer Controller's service account must be able to `get`/`list`/
`watch` Secrets in the ingress namespace. The upstream Helm chart grants this
cluster-wide, but **hardened installs may not** — if yours doesn't, add a namespaced
Role + RoleBinding (least privilege):

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: aws-lb-controller-oidc-secret
  namespace: flyte
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: aws-lb-controller-oidc-secret
  namespace: flyte
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: aws-lb-controller-oidc-secret
subjects:
- kind: ServiceAccount
  name: aws-load-balancer-controller   # adjust to your controller's SA
  namespace: kube-system
```

> Skipping this produces `FailedBuildModel … secrets "flyte2-console-oidc" is
> forbidden` on the ingress and **no auth rule is applied** (see Troubleshooting).

### 4. Add the auth annotations to the console ingress

**Chart-managed (recommended)** — add to `ingress.httpAnnotations` in your values
(e.g. `values-union.yaml`) and `helm upgrade`:

```yaml
ingress:
  httpAnnotations:
    # … existing alb annotations (cert-arn, group.name, listen-ports, etc.) …
    alb.ingress.kubernetes.io/auth-type: oidc
    alb.ingress.kubernetes.io/auth-scope: openid email
    alb.ingress.kubernetes.io/auth-on-unauthenticated-request: authenticate
    alb.ingress.kubernetes.io/auth-session-timeout: "604800"   # 7 days
    alb.ingress.kubernetes.io/auth-idp-oidc: '{"issuer":"<issuer>","authorizationEndpoint":"<authorize>","tokenEndpoint":"<token>","userInfoEndpoint":"<userinfo>","secretName":"flyte2-console-oidc"}'
```

**Surgical (no helm upgrade)** — if the live release has drift you don't want to
reconcile (e.g. an image/configmap patched by hand), apply the annotations directly
to the `-http` ingress; auth only affects that ingress's `/v2` rule:

```bash
kubectl annotate ingress flyte2-http -n flyte --overwrite \
  alb.ingress.kubernetes.io/auth-type=oidc \
  'alb.ingress.kubernetes.io/auth-scope=openid email' \
  alb.ingress.kubernetes.io/auth-on-unauthenticated-request=authenticate \
  alb.ingress.kubernetes.io/auth-session-timeout=604800 \
  'alb.ingress.kubernetes.io/auth-idp-oidc={"issuer":"<issuer>","authorizationEndpoint":"<authorize>","tokenEndpoint":"<token>","userInfoEndpoint":"<userinfo>","secretName":"flyte2-console-oidc"}'
```
(If you go surgical, still commit the same block to your values file so a future
`helm upgrade` doesn't drop it.)

### 5. Verify

```bash
# Console redirects to the IdP:
curl -s -o /dev/null -D - https://<host>/v2 | grep -i '^location'
# → 302 … https://<issuer>/v1/authorize?client_id=…&redirect_uri=https%3A%2F%2F<host>%2Foauth2%2Fidpresponse&scope=openid%20email…

# API path is NOT gated (should be a normal response, not a 302 to the IdP):
curl -s -o /dev/null -w '%{http_code}\n' -X POST \
  https://<host>/flyteidl2.project.ProjectService/ListProjects \
  -H 'Content-Type: application/json' -d '{}'
```

Then open `https://<host>/v2` in a browser — you should be bounced through the IdP
and back into the console.

---

## Run attribution (`executed_by`)

Once auth happens at the proxy, the **runs service** records who created each run
(`ActionMetadata.executed_by`) by reading the identity headers the proxy forwards —
it does **not** re-validate tokens itself. After ALB `authenticate-oidc`, those are:

- `X-Amzn-Oidc-Data` — signed JWT with the full claims (`sub`, `email`, `given_name`,
  `family_name`); used for the browser/cookie path.
- `X-Amzn-Oidc-Identity` — the subject only; used when the data header is absent.
- `Authorization: Bearer <jwt>` — the SDK/CLI path (proxy-agnostic, always honored).
  The token carries only the subject, so name/email are filled from the IdP's
  `userinfo` endpoint when `runs.authMetadata.externalAuthServerBaseUrl` is set.

The defaults match ALB, so nothing extra is needed here. **The decoded JWTs are not
signature-verified by the runs service** — that's safe only behind a trusted proxy
that validates tokens and strips client-supplied copies of these headers. If the
service can be reached directly, set `runs.trustForwardedIdentityHeaders: false`
and `executed_by` is left unset rather than risk a spoofed identity.

### Behind a non-ALB proxy (oauth2-proxy / Traefik forward-auth)

The header names are configurable, so this works behind any auth proxy. For
oauth2-proxy / Traefik forward-auth, which forward plain values rather than a JWT:

```yaml
runs:
  identityHeaders:
    claimsJwtHeader: ""                  # no JWT header on this path
    subjectHeader: X-Auth-Request-User   # ALB default: X-Amzn-Oidc-Identity
    emailHeader: X-Auth-Request-Email    # ALB default: unset (email is in the JWT)
```

| Config | Header read | ALB default | oauth2-proxy / Traefik |
|---|---|---|---|
| `claimsJwtHeader` | JWT with full claims | `X-Amzn-Oidc-Data` | *(empty)* |
| `subjectHeader` | subject, plain value | `X-Amzn-Oidc-Identity` | `X-Auth-Request-User` |
| `emailHeader` | email, plain value | *(unset)* | `X-Auth-Request-Email` |

The same trust boundary applies: the proxy must validate identity and strip any
client-supplied copies of these headers, with `trustForwardedIdentityHeaders` on.

---

## Troubleshooting

These are the real failures you'll hit, and how to tell them apart.

### `FailedBuildModel … secrets "…" is forbidden`
The LB controller can't read the Secret. Apply the RBAC in **step 3**. Check with:
```bash
kubectl describe ingress flyte2-http -n flyte | sed -n '/Events:/,$p'
```

### Browser shows `invalid_request … 'redirect_uri' parameter must be a Login redirect URI`
The exact callback isn't registered on the IdP app. Add
`https://<host>/oauth2/idpresponse` (with that path) to the app's sign-in redirect
URIs. You can probe it without a browser — this returns the IdP login page once the
URI is registered, and the error until then:
```bash
curl -s 'https://<issuer>/v1/authorize?client_id=<cid>&response_type=code&scope=openid%20email&redirect_uri=https://<host>/oauth2/idpresponse&state=t&nonce=n' \
  | grep -oiE 'invalid_request|must be a Login redirect URI|okta-sign-in|sign-in-widget'
```

### `401 Authorization Required` *after* a successful login
ALB got the code but the **token exchange failed** — almost always a wrong client
**secret** (or wrong client_id). Validate the creds directly against the token
endpoint with a bogus code; the error format tells you which:

```bash
curl -s -X POST 'https://<issuer>/v1/token' -u '<client-id>:<client-secret>' \
  --data-urlencode grant_type=authorization_code \
  --data-urlencode code=bogus --data-urlencode redirect_uri=https://<host>/oauth2/idpresponse
```

| Response | Meaning |
|---|---|
| `{"error":"invalid_grant", …}` | **creds are GOOD** (only the bogus code is rejected) ✅ |
| `{"error":"invalid_client","error_description":"The client secret … is invalid"}` | client_id is recognized but the **secret is wrong** |
| `{"errorCode":"invalid_client","errorSummary":"Invalid value for 'client_id' parameter."}` | the **client_id is wrong** (e.g. you used the app *label* or app *instance id* instead of the real OAuth client_id) |

> **Two common copy-paste traps:**
> - A **trailing `%`** on a secret copied from a terminal is zsh's
>   "no-newline" marker, not part of the secret. Strip it.
> - In Okta, the OAuth **client_id** is usually the opaque `0oa…` id (or a custom
>   string), **not** the human-readable app name. The `errorCode` vs `error` JSON
>   format above disambiguates which one the IdP accepted.

---

## Annotation reference

| Annotation | Purpose |
|---|---|
| `auth-type: oidc` | enable OIDC auth on this ingress's HTTPS rules |
| `auth-idp-oidc` | JSON: `issuer`, `authorizationEndpoint`, `tokenEndpoint`, `userInfoEndpoint`, `secretName` (Secret with `clientID`/`clientSecret`) |
| `auth-scope` | space-separated scopes, e.g. `openid email` |
| `auth-on-unauthenticated-request` | `authenticate` (challenge), `allow`, or `deny` |
| `auth-session-timeout` | session cookie lifetime in seconds |

ALB callback path (fixed): `/oauth2/idpresponse`. Auth applies only to the
annotated ingress, and only to its **HTTPS** listener rules.

---

## Teardown

```bash
kubectl annotate ingress flyte2-http -n flyte \
  alb.ingress.kubernetes.io/auth-type- \
  alb.ingress.kubernetes.io/auth-scope- \
  alb.ingress.kubernetes.io/auth-on-unauthenticated-request- \
  alb.ingress.kubernetes.io/auth-session-timeout- \
  alb.ingress.kubernetes.io/auth-idp-oidc-
# and, if you no longer need them:
kubectl delete secret flyte2-console-oidc -n flyte
kubectl delete role,rolebinding aws-lb-controller-oidc-secret -n flyte
```
(Remove the same block from your values file too, or the next `helm upgrade`
re-adds it.)
