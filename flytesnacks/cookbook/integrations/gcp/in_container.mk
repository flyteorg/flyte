SERIALIZED_PB_OUTPUT_DIR := /tmp/output

.PHONY: clean
clean:
	rm -rf $(SERIALIZED_PB_OUTPUT_DIR)/*

$(SERIALIZED_PB_OUTPUT_DIR): clean
	mkdir -p $(SERIALIZED_PB_OUTPUT_DIR)

.PHONY: serialize
serialize: $(SERIALIZED_PB_OUTPUT_DIR)
	pyflyte --config /root/sandbox.config serialize workflows -f $(SERIALIZED_PB_OUTPUT_DIR)

.PHONY: fast_serialize
fast_serialize: $(SERIALIZED_PB_OUTPUT_DIR)
	pyflyte --config /root/sandbox.config serialize fast workflows -f $(SERIALIZED_PB_OUTPUT_DIR)
