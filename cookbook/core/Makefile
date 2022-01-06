PREFIX=core
PWD=$(shell pwd)
include ../common/common.mk
include ../common/leaf.mk

.PHONY: extra_images
extra_images:
	make -C containerization/raw-containers-supporting-files build_images
