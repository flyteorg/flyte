package assets

import (
	// This import is used to embed the admin.swagger.json file into the binary.
	_ "embed"
)

// AdminSwaggerFile is the admin.swagger.json file embedded into the binary.
//
//go:generate cp ../../../gen/pb-go/gateway/flyteidl/service/admin.swagger.json admin.swagger.json
//go:embed admin.swagger.json
var AdminSwaggerFile []byte
