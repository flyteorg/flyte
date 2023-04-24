# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'FLYTEORG/BOILERPLATE' REPOSITORY:
#
# TO OPT OUT OF UPDATES, SEE https://github.com/flyteorg/boilerplate/blob/master/Readme.rst

current_dir := $(CURDIR)

.PHONY: end2end_execute
end2end_execute:
	$(current_dir)/end2end.sh $(current_dir)/functional-test-config.yaml --return_non_zero_on_failure

.PHONY: k8s_integration_execute
k8s_integration_execute:
	echo "pass"
