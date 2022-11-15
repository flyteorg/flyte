
GENERATED_FILE="gen/binary_manifest.yaml"
DEMO_GENERATED_FILE="gen/demomanifest.yaml"

cleanhelm:
	@[ -f $(GENERATED_FILE) ] && rm $(GENERATED_FILE) || true

cleandemo:
	@[ -f $(DEMO_GENERATED_FILE) ] && rm $(DEMO_GENERATED_FILE) || true

.PHONY: helm
helm: cleanhelm
	helm template binary-tst-deploy ./ -n flyte --dependency-update --debug --create-namespace -a rbac.authorization.k8s.io/v1 -a networking.k8s.io/v1/Ingress -a apiextensions.k8s.io/v1/CustomResourceDefinition > $(GENERATED_FILE)

.PHONY: demohelm
demohelm: cleandemo
	helm template flytedemo ./ -f flytectldemo.yaml -n flyte --dependency-update --debug --create-namespace -a rbac.authorization.k8s.io/v1 -a networking.k8s.io/v1/Ingress -a apiextensions.k8s.io/v1/CustomResourceDefinition > $(DEMO_GENERATED_FILE)
