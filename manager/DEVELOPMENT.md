# Development Guide

> Note: this development guide is stale, should use `make sandbox-run` command

Use following method to setup the whole system and run:

1. Create k3d cluster with

```sh
make cluster-create
```

2. Install `TaskAction` CR

```sh
make -C executor install
```

3. Deploy MinIO storage backend and port forward

```sh
kubectl apply -f dataproxy/deployment/minio.yaml
```

```sh
kubectl port-forward -n flyte-dataproxy svc/minio 9000:9000 9001:9001
```

4. Create `flyte-data` bucket in minio
   1. Login minio in `localhost:9001` with account and passward `minioadmin`
   2. Create bucket `flyte-data`

5. Start the manager

```sh
make -C manager run
```

6. Run `python example/basics/hello.py` in flyte-sdk with following config:

```yaml
admin:
  endpoint: dns:///localhost:8090
  insecure: True
image:
  builder: local
task:
  domain: development
  project: testproject
  org: testorg
```

7. Check if the pods are created. Each pod is an action:

```sh
❯ k get -n flyte pods
NAME                                                                       READY   STATUS      RESTARTS   AGE
testorg-testproject-development-run-1772092651-2h4g59o2hacijbxsm3oi9sx6    0/1     Completed   0          29m
testorg-testproject-development-run-1772092651-2you3apzuf5dpi8algfewpkgg   0/1     Completed   0          29m
testorg-testproject-development-run-1772092651-59nwdmghanjvxmk1ahgwnusv9   0/1     Completed   0          29m
testorg-testproject-development-run-1772092651-5oas2d435tc8quccoviknkpxd   0/1     Completed   0          29m
testorg-testproject-development-run-1772092651-5rw60n2g061co0jsyvc4prfyp   0/1     Completed   0          29m
testorg-testproject-development-run-1772092651-717du5er7pmdjd7dq2jiv89cb   0/1     Completed   0          29m
testorg-testproject-development-run-1772092651-aetk1g6slctsoi7qvr8vkc78u   0/1     Completed   0          29m
testorg-testproject-development-run-1772092651-br272e7cmghio54vju3jvgmaj   0/1     Completed   0          29m
testorg-testproject-development-run-1772092651-dv1i0ica99x5dhkg6cg5mos3e   0/1     Completed   0          29m
testorg-testproject-development-run-1772092651-dxvh90nrbft3cj2dh655bg371   0/1     Completed   0          29m
testorg-testproject-development-run-1772092651-run-1772092651              0/1     Completed   0          29m
```
