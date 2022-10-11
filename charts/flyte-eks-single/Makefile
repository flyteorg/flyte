
GENERATED_FILE="gen/eks_manifest.yaml"

clean:
	@[ -f $(GENERATED_FILE) ] && rm $(GENERATED_FILE) || true

.PHONY: helm
helm: clean
	helm template binary-tst-eks ./ -n flyte --render-subchart-notes --dependency-update --debug --create-namespace -a rbac.authorization.k8s.io/v1 -a networking.k8s.io/v1/Ingress -a apiextensions.k8s.io/v1/CustomResourceDefinition > $(GENERATED_FILE)

