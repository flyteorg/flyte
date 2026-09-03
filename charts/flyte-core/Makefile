.PHONY: lint template

lint:
	helm dependency build .
	helm lint .

template:
	helm dependency build .
	helm template flyte-core . --namespace flyte --include-crds
