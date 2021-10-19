.PHONY: rsts
rsts:
	sphinx-apidoc --force --tocfile index --ext-autodoc --output-dir _rsts ../../

.PHONY: html
html:
	sphinx-build -n -b html . _build && sphinx-build -n -b doctest . _build && sphinx-build -n -b linkcheck . _build

.PHONY: clean
clean:
	rm -rf _build
	rm -rf auto_*
	rm -rf auto/*
