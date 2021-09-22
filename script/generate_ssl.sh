#!/bin/bash

script="$0"
basename="$(dirname $script)"
 
echo "Script name $script resides in $basename directory."

openssl genrsa -des3 -out "$basename/rootCA.key" 2048
openssl req -x509 -new -nodes -key "$basename/rootCA.key" -sha256 -days 1024 -out "$basename/rootCA.pem"
openssl req -new -sha256 -nodes -out "$basename/server.csr" -newkey rsa:2048 -keyout "$basename/server.key" -config <( cat "$basename/server.csr.cnf" )
openssl x509 -req -in "$basename/server.csr" -CA "$basename/rootCA.pem" -CAkey "$basename/rootCA.key" -CAcreateserial -out "$basename/server.crt" -days 500 -sha256 -extfile "$basename/v3.ext"
