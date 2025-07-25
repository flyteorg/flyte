FLYTE_INTERNAL_TASK_DOMAIN="development" \
FLYTE_INTERNAL_TASK_PROJECT="testproject" \
_U_ORG_NAME="testorg" \
_U_RUN_BASE="s3://bucket/metadata/v2/testorg/testproject/development/rcf7sw8frpsbz7gcmk9v/bswd03zvzi8892d19tyoab8eq/1" \
ACTION_NAME="a0" \
RUN_NAME="run-1" \
unionai-actor-tester \
RUN_POOL="true" \
a0 --inputs s3://bucket/metadata/v2/testorg/testproject/development/rcf7sw8frpsbz7gcmk9v/bswd03zvzi8892d19tyoab8eq/inputs.pb --outputs-path s3://bucket/metadata/v2/testorg/testproject/development/rcf7sw8frpsbz7gcmk9v/bswd03zvzi8892d19tyoab8eq/1 --version 904a87c5b877afff041d3ff38e75b362 --raw-data-path s3://bucket/metadata/v2/testorg/testproject/development/rcf7sw8frpsbz7gcmk9v/bswd03zvzi8892d19tyoab8eq/1/di/rcf7sw8frpsbz7gcmk9v-bswd03zvzi8892d19tyoab8eq-0 --checkpoint-path s3://bucket/metadata/v2/testorg/testproject/development/rcf7sw8frpsbz7gcmk9v/bswd03zvzi8892d19tyoab8eq/1/di/rcf7sw8frpsbz7gcmk9v-bswd03zvzi8892d19tyoab8eq-0/_flytecheckpoints --prev-checkpoint "" --run-name {{.runName}} --name {{.actionName}} --image-cache H4sIAAAAAAAC/wXB0QqDIBQA0H+571tKmuQPbBAE9QPjZuok65bOh4j+fedcEFb09hOJlrKDvqBf0rvLJ73WkU3xKHKwZvSgwX9NegaqyhZow/CgnCsXz5/Vs5wYt5zV3NXGoRK8nRkqpaSQ2DYN3PcfUD/ec2gAAAA= --tgz s3://bucket/testproject/development/SBFIPRNYO6X76BA5H7ZY45NTMI======/fast5d5a40b1e9323acc394b96feb9aa34cf.tar.gz --dest . --resolver flyte._internal.resolvers.default.DefaultTaskResolver mod examples.reuse.reusable instance square
