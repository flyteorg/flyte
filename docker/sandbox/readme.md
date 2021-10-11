to build local image 

```
docker build --target=default --tag flyte_arm:normal  -f docker/sandbox/Dockerfile .

docker build --target=dind --tag flyte_arm:dind  -f docker/sandbox/Dockerfile .
```


docker buildx build --platform linux/amd64,linux/arm64 --target=default --tag flyte_sandbox-test  -f docker/sandbox/Dockerfile  .