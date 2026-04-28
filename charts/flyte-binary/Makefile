
GENERATED_FILE="gen/binary_manifest.yaml"
DEVBOX_GENERATED_FILE="gen/devboxmanifest.yaml"

cleanhelm:
	@[ -f $(GENERATED_FILE) ] && rm $(GENERATED_FILE) || true

cleandevbox:
	@[ -f $(DEVBOX_GENERATED_FILE) ] && rm $(DEVBOX_GENERATED_FILE) || true

.PHONY: helm
helm: cleanhelm
	helm template binary-tst-deploy ./ -n flyte --dependency-update --debug --create-namespace -a rbac.authorization.k8s.io/v1 -a networking.k8s.io/v1/Ingress -a apiextensions.k8s.io/v1/CustomResourceDefinition > $(GENERATED_FILE)

.PHONY: devboxhelm
devboxhelm: cleandevbox
	helm template flytedevbox ./ -f flytectldemo.yaml -n flyte --dependency-update --debug --create-namespace -a rbac.authorization.k8s.io/v1 -a networking.k8s.io/v1/Ingress -a apiextensions.k8s.io/v1/CustomResourceDefinition > $(DEVBOX_GENERATED_FILE)
