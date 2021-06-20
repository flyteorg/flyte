export REPOSITORY=flyte

define PIP_COMPILE
pip-compile $(1) --upgrade --verbose
endef

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

.PHONY: kustomize
kustomize: 
	KUSTOMIZE_VERSION=3.9.2 bash script/generate_kustomize.sh

.PHONY: helm
helm: 
	bash script/generate_helm.sh

.PHONY: release_automation
release_automation:
	mkdir -p release
	bash script/release.sh

.PHONY: deploy_sandbox
deploy_sandbox:
	bash script/deploy.sh

# launch dockernetes and execute tests
.PHONY: end2end
end2end:
	@end2end/launch_dockernetes.sh

# execute tests in the current kubernetes context
.PHONY: end2end_execute
end2end_execute:
	@end2end/execute.sh

.PHONY: install-piptools
install-piptools:
	pip install -U pip-tools

.PHONY: doc-requirements.txt
doc-requirements.txt: doc-requirements.in install-piptools
	$(call PIP_COMPILE,doc-requirements.in)

.PHONY: requirements.txt
requirements.txt: requirements.in install-piptools
	$(call PIP_COMPILE,requirements.in)

.PHONY: stats
stats:
	@generate-dashboard -o deployment/stats/prometheus/flytepropeller-dashboard.json stats/flytepropeller_dashboard.py
	@generate-dashboard -o deployment/stats/prometheus/flyteadmin-dashboard.json stats/flyteadmin_dashboard.py
	@generate-dashboard -o deployment/stats/prometheus/flyteuser-dashboard.json stats/flyteuser_dashboard.py

.PHONY: prepare_artifacts
prepare_artifacts:
	bash script/prepare_artifacts.sh

.PHONY: helm_install
helm_install:
	helm install flyte --debug ./helm  -f helm/values-sandbox.yaml --create-namespace --namespace=flyte

.PHONY: helm_upgrade
helm_upgrade:
	helm upgrade flyte --debug ./helm  -f helm/values-sandbox.yaml --create-namespace --namespace=flyte
	bash script/prepare_artifacts.sh

.PHONY: docs
docs:
	make -C rsts clean html SPHINXOPTS=-W
