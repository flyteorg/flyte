.PHONY: rsts
rsts:
	sphinx-apidoc --force --tocfile index --ext-autodoc --output-dir _rsts ../../

.PHONY: html
html:
	sphinx-build -n -W -b html . _build

.PHONY: linkcheck
linkcheck:
	sphinx-build -n -b linkcheck . _build

.PHONY: doctest
doctest:
	sphinx-build -n -b doctest . _build

.PHONY: clean
clean:
	rm -rf _build
	rm -rf auto_*
	rm -rf auto/*
