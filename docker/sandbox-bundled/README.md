# Flyte Deployment Sandbox


Testing mysql changes.

Running the deployment/service manually

Need to port forward because the new port is not exposed through the docker container yet.
Don't want to coopt the postgres port for now.
```
kf port-forward service/flyte-sandbox-mysql 30003:3306
```


