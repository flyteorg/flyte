.PHONY: rsts
rsts:
	sphinx-apidoc --force --tocfile index --ext-autodoc --output-dir _rsts ../recipes

.PHONY: html
html:
	sphinx-build -b html . _build

.PHONY: clean
clean:
	rm -rf _build
	rm -rf auto_core_basic auto_core_intermediate auto_core_advanced auto_case_studies auto_plugins
