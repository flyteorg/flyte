#!/bin/bash
set -x

TMP_DEPS_FOLDER=tmp/doc_gen_deps
TMP_DEPS_PROTOBUF_FOLDER=tmp/protocolbuffers
TMP_DEPS_GRPC_GATEWAY_FOLDER=tmp/grpc-gateway
TMP_DEPS_K8S_IO=tmp/k8s.io

# clear all the tmp deps folder
rm -rf $TMP_DEPS_FOLDER $TMP_DEPS_PROTOBUF_FOLDER $TMP_DEPS_GRPC_GATEWAY_FOLDER $TMP_DEPS_K8S_IO

# clone deps and move them to TMP_DEPS_FOLDER
# googleapi deps
git clone --depth 1 https://github.com/googleapis/googleapis $TMP_DEPS_FOLDER
rm -rf $TMP_DEPS_FOLDER/.git

# protobuf deps
git clone --depth 1 https://github.com/protocolbuffers/protobuf $TMP_DEPS_PROTOBUF_FOLDER
cp -r $TMP_DEPS_PROTOBUF_FOLDER/src/* $TMP_DEPS_FOLDER
rm -rf $TMP_DEPS_PROTOBUF_FOLDER

# grpc-gateway deps
git -c advice.detachedHead=false clone --depth 1 --branch v1.15.2 https://github.com/grpc-ecosystem/grpc-gateway $TMP_DEPS_GRPC_GATEWAY_FOLDER #v1.15.2 is used to keep the grpc-gateway version in sync with generated protos which is using the LYFT image
cp -r $TMP_DEPS_GRPC_GATEWAY_FOLDER/protoc-gen-swagger $TMP_DEPS_FOLDER
rm -rf $TMP_DEPS_GRPC_GATEWAY_FOLDER

# k8 dependencies
git clone --depth 1 https://github.com/kubernetes/api $TMP_DEPS_K8S_IO/api
git clone --depth 1 https://github.com/kubernetes/apimachinery $TMP_DEPS_K8S_IO/apimachinery
cp -r $TMP_DEPS_K8S_IO $TMP_DEPS_FOLDER
rm -rf $TMP_DEPS_K8S_IO
