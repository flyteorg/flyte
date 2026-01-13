#!/bin/bash

ENDPOINT="${ENDPOINT:-http://localhost:8088}"
PROJECT="${PROJECT:-testproject}"
DOMAIN="${DOMAIN:-development}"
ORG="${ORG:-testorg}"
FILENAME="${FILENAME:-test-file.txt}"
FILENAME_ROOT="${FILENAME_ROOT:-test-upload}"

# Create a simple MD5 hash (16 bytes in base64)
# "test-content-hash" -> base64 encoded
CONTENT_MD5="dGVzdC1jb250ZW50LWhhc2g="

buf curl --schema . $ENDPOINT/flyteidl2.dataproxy.DataProxyService/CreateUploadLocation --data @- <<EOF
{
    "project": "$PROJECT",
    "domain": "$DOMAIN",
    "org": "$ORG",
    "filename": "$FILENAME",
    "filename_root": "$FILENAME_ROOT",
    "content_md5": "$CONTENT_MD5",
    "expires_in": "1800s",
    "content_length": 1024
}
EOF
