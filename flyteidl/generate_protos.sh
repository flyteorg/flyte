#!/bin/bash

DIR=`pwd`
rm -rf $DIR/gen
LYFT_IMAGE="lyft/protocgenerator:8167e11d3b3439373c2f033080a4b550078884a2"
SWAGGER_CLI_IMAGE="docker.io/lyft/swagger-codegen-cli:dc5ce6ec6d7d4d980fa882d6bd13a83cba3be3c3"

docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/service --with_gateway -l go --go_source_relative
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/admin --with_gateway -l go --go_source_relative --validate_out
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/core --with_gateway -l go --go_source_relative --validate_out
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/event --with_gateway -l go --go_source_relative --validate_out
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/plugins -l go --go_source_relative --validate_out
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/datacatalog -l go --go_source_relative --validate_out

languages=("python" "cpp" "java")
idlfolders=("service" "admin" "core" "event" "plugins" "datacatalog")

for lang in "${languages[@]}"
do
    for folder in "${idlfolders[@]}"
    do
        docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs $LYFT_IMAGE -i ./protos -d protos/flyteidl/$folder -l $lang
    done
done

# Docs generated

# Get list of proto files in core directory
core_proto_files=`ls protos/flyteidl/core/*.proto |xargs`
# Remove any currently generated file
ls -d protos/docs/core/* | grep -v index.rst | xargs rm
# Use the list to generate the RST files required for sphinx conversion. Additionally generate for google.protobuf.[timestamp | struct | duration].
protoc --doc_out=protos/docs/core --doc_opt=restructuredtext,core.rst -I=tmp/doc_gen_deps -I=protos `echo $core_proto_files` tmp/doc_gen_deps/google/protobuf/timestamp.proto tmp/doc_gen_deps/google/protobuf/duration.proto tmp/doc_gen_deps/google/protobuf/struct.proto

# Get list of proto files in admin directory
admin_proto_files=`ls protos/flyteidl/admin/*.proto |xargs`
# Remove any currently generated file
ls -d protos/docs/admin/* | grep -v index.rst | xargs rm
# Use the list to generate the RST files required for sphinx conversion
protoc --doc_out=protos/docs/admin --doc_opt=protos/docs/withoutscalar_restructuredtext.tmpl,admin.rst -I=tmp/doc_gen_deps -I=protos `echo $admin_proto_files` tmp/doc_gen_deps/google/protobuf/duration.proto

# Get list of proto files in datacatlog directory
datacatalog_proto_files=`ls protos/flyteidl/datacatalog/*.proto |xargs`
# Remove any currently generated file
ls -d protos/docs/datacatalog/* | grep -v index.rst | xargs rm
# Use the list to generate the RST files required for sphinx conversion
protoc --doc_out=protos/docs/datacatalog --doc_opt=protos/docs/withoutscalar_restructuredtext.tmpl,datacatalog.rst -I=tmp/doc_gen_deps -I=protos `echo $datacatalog_proto_files` tmp/doc_gen_deps/google/protobuf/timestamp.proto tmp/doc_gen_deps/google/protobuf/duration.proto tmp/doc_gen_deps/google/protobuf/struct.proto

# Get list of proto files in event directory
event_proto_files=`ls protos/flyteidl/event/*.proto |xargs`
# Remove any currently generated file
ls -d protos/docs/event/* | grep -v index.rst | xargs rm
# Use the list to generate the RST files required for sphinx conversion
protoc --doc_out=protos/docs/event --doc_opt=protos/docs/withoutscalar_restructuredtext.tmpl,event.rst -I=tmp/doc_gen_deps -I=protos `echo $event_proto_files` tmp/doc_gen_deps/google/protobuf/timestamp.proto tmp/doc_gen_deps/google/protobuf/duration.proto tmp/doc_gen_deps/google/protobuf/struct.proto

# Get list of proto files in plugins directory.
plugins_proto_files=`ls protos/flyteidl/plugins/*.proto | xargs`
# Remove any currently generated file
ls -d protos/docs/plugins/* |grep -v index.rst| xargs rm
# Use the list to generate the RST files required for sphinx conversion
protoc --doc_out=protos/docs/plugins --doc_opt=protos/docs/withoutscalar_restructuredtext.tmpl,plugins.rst -I=protos -I=tmp/doc_gen_deps `echo $plugins_proto_files`

# Get list of proto files in service directory.
service_proto_files=`ls protos/flyteidl/service/*.proto | xargs`
# Remove any currently generated file
ls -d protos/docs/service/* |grep -v index.rst| xargs rm
# Use the list to generate the RST files required for sphinx conversion
protoc --doc_out=protos/docs/service --doc_opt=protos/docs/withoutscalar_restructuredtext.tmpl,service.rst -I=protos -I=tmp/doc_gen_deps `echo $service_proto_files`

# Generate binary data from OpenAPI 2 file
docker run --rm -u $(id -u):$(id -g) -v $DIR/gen/pb-go/flyteidl/service:/service --entrypoint go-bindata $LYFT_IMAGE -pkg service -o /service/openapi.go -prefix /service/ -modtime 1562572800 /service/admin.swagger.json

# Generate JS code
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs schottra/docker-protobufjs:v0.0.2 --module-name flyteidl -d protos/flyteidl/core  -d protos/flyteidl/event -d protos/flyteidl/admin -d protos/flyteidl/service  -- --root flyteidl -t static-module -w es6 --no-delimited --force-long --no-convert -p /defs/protos

# Generate GO API client code
docker run --rm -u $(id -u):$(id -g) --rm -v $DIR:/defs $SWAGGER_CLI_IMAGE generate -i /defs/gen/pb-go/flyteidl/service/admin.swagger.json -l go -o /defs/gen/pb-go/flyteidl/service/flyteadmin --additional-properties=packageName=flyteadmin

# # Generate Python API client code
docker run --rm -u $(id -u):$(id -g) --rm -v $DIR:/defs $SWAGGER_CLI_IMAGE generate -i /defs/gen/pb-go/flyteidl/service/admin.swagger.json -l python -o /defs/gen/pb_python/flyteidl/service/flyteadmin --additional-properties=packageName=flyteadmin

# Remove documentation generated from the swagger-codegen-cli
rm -rf gen/pb-go/flyteidl/service/flyteadmin/docs
rm -rf gen/pb_python/flyteidl/service/flyteadmin/docs


# Unfortunately, the `--grpc-gateway-out` plugin doesn’t yet support the `source_relative` option. Until it does, we need to move the files from the autogenerated location to the source_relative location.
cp -r gen/pb-go/github.com/flyteorg/flyteidl/gen/* gen/
rm -rf gen/pb-go/github.com

# Copy the validate.py protos.
mkdir -p gen/pb_python/validate
cp -r validate/* gen/pb_python/validate/

# This section is used by Travis CI to ensure that the generation step was run
if [ -n "$DELTA_CHECK" ]; then
  DIRTY=$(git status --porcelain)
  if [ -n "$DIRTY" ]; then
    echo "FAILED: Protos updated without commiting generated code."
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
