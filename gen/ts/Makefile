PLACEHOLDER_NPM := \"version\": \"0.0.0-develop\"

.PHONY: update_npmversion
update_npmversion:
	grep "$(PLACEHOLDER_NPM)" "package.json"
	sed -i "s/$(PLACEHOLDER_NPM)/\"version\":  \"${VERSION}\"/g" "package.json"
