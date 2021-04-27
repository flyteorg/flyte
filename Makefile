define PIP_COMPILE
pip-compile $(1) --upgrade --verbose
endef

.PHONY: kustomize
kustomize: 
	KUSTOMIZE_VERSION=3.9.2 bash script/generate_kustomize.sh

.PHONY: release_automation
release_automation:
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