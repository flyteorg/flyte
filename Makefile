.PHONY: kustomize
kustomize:
	bash script/kustomize.sh

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

# Use this target to build the rsts directory only. In order to build the entire flyte docs, use update_ref_docs && all_docs
.PHONY: generate-local-docs
generate-local-docs:
	@docker run -t -v `pwd`:/base lyft/docbuilder:v2.2.0 sphinx-build -E -b html /base/rsts/. /base/_build

# Builds the entire doc tree. Assumes update_ref_docs has run and that all externals rsts are in _rsts/ dir
.PHONY: generate-docs
generate-docs: generate-dependent-repo-docs
	@FLYTEKIT_VERSION=0.7.0 ./script/generate_docs.sh

# updates referenced docs from other repositories (e.g. flyteidl, flytekit)
.PHONY: generate-dependent-repo-docs
generate-dependent-repo-docs:
	@FLYTEKIT_VERSION=0.7.0 FLYTEIDL_VERSION=0.17.27 ./script/update_ref_docs.sh
