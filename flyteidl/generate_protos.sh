#!/bin/bash
set -e

DIR=`pwd`
rm -rf $DIR/gen
LYFT_IMAGE="lyft/protocgenerator:8167e11d3b3439373c2f033080a4b550078884a2"
SWAGGER_CLI_IMAGE="ghcr.io/flyteorg/swagger-codegen-cli:latest"
PROTOC_GEN_DOC_IMAGE="pseudomuto/protoc-gen-doc:1.4.1"

# Override system locale during protos/docs generation to ensure consistent sorting (differences in system locale could e.g. lead to differently ordered docs)
export LC_ALL=C.UTF-8

docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/service --with_gateway -l go --go_source_relative
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/admin --with_gateway -l go --go_source_relative
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/core --with_gateway -l go --go_source_relative
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/event --with_gateway -l go --go_source_relative
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/plugins -l go --go_source_relative
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/datacatalog -l go --go_source_relative

# languages=("cpp" "java")
# idlfolders=("service" "admin" "core" "event" "plugins" "datacatalog")

# for lang in "${languages[@]}"
# do
#     for folder in "${idlfolders[@]}"
#     do
#         docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/$folder -l $lang
#     done
# done

# Buf migration
docker run -u $(id -u):$(id -g) -e "BUF_CACHE_DIR=/tmp/cache" --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf generate

# Unfortunately the python protoc plugin does not add __init__.py files to the generated code
# (as described in https://github.com/protocolbuffers/protobuf/issues/881). One of the
# suggestions is to manually create such files, which is what we do here:
find gen/pb_python -type d -exec touch {}/__init__.py \;

# Docs generated

# # Remove any currently generated core docs file
# ls -d protos/docs/core/* | grep -v index.rst | xargs rm
# # Use list of proto files in core directory to generate the RST files required for sphinx conversion. Additionally generate for google.protobuf.[timestamp | struct | duration].
# docker run --rm -u $(id -u):$(id -g) -v $DIR/protos:/protos:ro -v $DIR/protos/docs/core:/out:rw -v $DIR/tmp/doc_gen_deps:/tmp/doc_gen_deps:ro $PROTOC_GEN_DOC_IMAGE --doc_opt=protos/docs/restructuredtext.tmpl,core.rst -I=tmp/doc_gen_deps -I=protos `ls protos/flyteidl/core/*.proto | xargs` tmp/doc_gen_deps/google/protobuf/timestamp.proto tmp/doc_gen_deps/google/protobuf/duration.proto tmp/doc_gen_deps/google/protobuf/struct.proto

# # Remove any currently generated admin docs file
# ls -d protos/docs/admin/* | grep -v index.rst | xargs rm
# # Use list of proto files in admin directory to generate the RST files required for sphinx conversion. Additionally generate for google.protobuf.[duration | wrappers].
# docker run --rm -u $(id -u):$(id -g) -v $DIR/protos:/protos:ro -v $DIR/protos/docs/admin:/out:rw -v $DIR/tmp/doc_gen_deps:/tmp/doc_gen_deps:ro $PROTOC_GEN_DOC_IMAGE --doc_opt=protos/docs/withoutscalar_restructuredtext.tmpl,admin.rst -I=tmp/doc_gen_deps -I=protos `ls protos/flyteidl/admin/*.proto | xargs` tmp/doc_gen_deps/google/protobuf/duration.proto tmp/doc_gen_deps/google/protobuf/wrappers.proto

# # Remove any currently generated datacatalog docs file
# ls -d protos/docs/datacatalog/* | grep -v index.rst | xargs rm
# # Use list of proto files in datacatalog directory to generate the RST files required for sphinx conversion. Additionally generate for google.protobuf.[timestamp | struct | duration].
# docker run --rm -u $(id -u):$(id -g) -v $DIR/protos:/protos:ro -v $DIR/protos/docs/datacatalog:/out:rw -v $DIR/tmp/doc_gen_deps:/tmp/doc_gen_deps:ro $PROTOC_GEN_DOC_IMAGE --doc_opt=protos/docs/withoutscalar_restructuredtext.tmpl,datacatalog.rst -I=tmp/doc_gen_deps -I=protos `ls protos/flyteidl/datacatalog/*.proto | xargs` tmp/doc_gen_deps/google/protobuf/timestamp.proto tmp/doc_gen_deps/google/protobuf/duration.proto tmp/doc_gen_deps/google/protobuf/struct.proto

# # Remove any currently generated event docs file
# ls -d protos/docs/event/* | grep -v index.rst | xargs rm
# # Use list of proto files in event directory to generate the RST files required for sphinx conversion. Additionally generate for google.protobuf.[timestamp | struct | duration].
# docker run --rm -u $(id -u):$(id -g) -v $DIR/protos:/protos:ro -v $DIR/protos/docs/event:/out:rw -v $DIR/tmp/doc_gen_deps:/tmp/doc_gen_deps:ro $PROTOC_GEN_DOC_IMAGE --doc_opt=protos/docs/withoutscalar_restructuredtext.tmpl,event.rst -I=tmp/doc_gen_deps -I=protos `ls protos/flyteidl/event/*.proto | xargs` tmp/doc_gen_deps/google/protobuf/timestamp.proto tmp/doc_gen_deps/google/protobuf/duration.proto tmp/doc_gen_deps/google/protobuf/struct.proto

# # Remove any currently generated plugins docs file
# ls -d protos/docs/plugins/* | grep -v index.rst | xargs rm
# # Use list of proto files in plugins directory to generate the RST files required for sphinx conversion
# docker run --rm -u $(id -u):$(id -g) -v $DIR/protos:/protos:ro -v $DIR/protos/docs/plugins:/out:rw -v $DIR/tmp/doc_gen_deps:/tmp/doc_gen_deps:ro $PROTOC_GEN_DOC_IMAGE --doc_opt=protos/docs/withoutscalar_restructuredtext.tmpl,plugins.rst -I=protos -I=tmp/doc_gen_deps `ls protos/flyteidl/plugins/*.proto | xargs`

# # Remove any currently generated service docs file
# ls -d protos/docs/service/* | grep -v index.rst | xargs rm
# # Use list of proto files in service directory to generate the RST files required for sphinx conversion
# docker run --rm -u $(id -u):$(id -g) -v $DIR/protos:/protos:ro -v $DIR/protos/docs/service:/out:rw -v $DIR/tmp/doc_gen_deps:/tmp/doc_gen_deps:ro $PROTOC_GEN_DOC_IMAGE --doc_opt=protos/docs/withoutscalar_restructuredtext.tmpl,service.rst -I=protos -I=tmp/doc_gen_deps `ls protos/flyteidl/service/*.proto | xargs`

# Generate binary data from OpenAPI 2 file
docker run --rm -u $(id -u):$(id -g) -v $DIR/gen/pb-go/flyteidl/service:/service --entrypoint go-bindata $LYFT_IMAGE -pkg service -o /service/openapi.go -prefix /service/ -modtime 1562572800 /service/admin.swagger.json

# Generate JS code

# This section is used by Travis CI to ensure that the generation step was run
if [ -n "$DELTA_CHECK" ]; then
  DIRTY=$(git status --porcelain)
  if [ -n "$DIRTY" ]; then
    echo "FAILED: Protos updated without committing generated code."
    echo "Ensure make generate has run and all changes are committed."
    DIFF=$(git diff)
    echo "diff detected: $DIFF"
    DIFF=$(git diff --name-only)
    echo "files different: $DIFF"
    exit 1
  else
    echo "SUCCESS: Generated code is up to date."
  fi
fi
