define PIP_COMPILE
pip-compile $(1) --upgrade --verbose
endef

.PHONY: kustomize
kustomize:
	KUSTOMIZE_VERSION=3.9.2 bash script/generate_kustomize.sh

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

# Builds the entire doc tree. Assumes update_ref_docs has run and that all externals rsts are in _rsts/ dir
.PHONY: generate-docs
generate-docs: generate-dependent-repo-docs
	@FLYTEKIT_VERSION=0.15.4 ./script/generate_docs.sh

# updates referenced docs from other repositories (e.g. flyteidl, flytekit)
.PHONY: generate-dependent-repo-docs
generate-dependent-repo-docs:
	@FLYTEKIT_VERSION=0.15.4 FLYTEIDL_VERSION=0.18.11 ./script/update_ref_docs.sh

.PHONY: install-piptools
install-piptools:
	pip install -U pip-tools

.PHONY: doc-requirements.txt
doc-requirements.txt: doc-requirements.in install-piptools
	$(call PIP_COMPILE,doc-requirements.in)

