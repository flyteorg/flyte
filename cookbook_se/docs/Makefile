.PHONY: rsts
rsts:
	sphinx-apidoc --force --tocfile index --ext-autodoc --output-dir _rsts ../recipes

.PHONY: html
html:
	sphinx-build -b html . _build

.PHONY: clean
clean:
	rm -rf _build
	rm -rf auto_recipes
