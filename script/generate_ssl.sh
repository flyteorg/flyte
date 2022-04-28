#!/bin/bash

script="$0"
basename="$(dirname $script)"
output="$basename/certificate"

echo "Script name $script resides in $basename directory. Output would be written to: $output"

mkdir -p $output

openssl genrsa -des3 -out "$output/rootCA.key" 2048
openssl req -x509 -new -nodes -key "$output/rootCA.key" -sha256 -days 1024 -out "$output/rootCA.pem"
openssl req -new -sha256 -nodes -out "$output/server.csr" -newkey rsa:2048 -keyout "$output/server.key" -config <( cat "$basename/server.csr.cnf" )
openssl x509 -req -in "$output/server.csr" -CA "$output/rootCA.pem" -CAkey "$output/rootCA.key" -CAcreateserial -out "$output/server.crt" -days 500 -sha256 -extfile "$basename/v3.ext"
