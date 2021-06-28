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

.PHONY: fast_register
fast_register: fast_serialize
	flyte-cli fast-register-files -h ${FLYTE_HOST} ${INSECURE_FLAG} -p ${PROJECT} -d development --kubernetes-service-account ${SERVICE_ACCOUNT} --output-location-prefix ${OUTPUT_DATA_PREFIX} --additional-distribution-dir ${ADDL_DISTRIBUTION_DIR} $(SERIALIZED_PB_OUTPUT_DIR)/*
