package common

import "github.com/flyteorg/flytestdlib/contextutils"

var RuntimeTypeKey = contextutils.Key("runtime_type")
var RuntimeVersionKey = contextutils.Key("runtime_version")

const (
	AuditFieldsContextKey contextutils.Key = "audit_fields"
	PrincipalContextKey   contextutils.Key = "principal"
)

const MaxResponseStatusBytes = 32000
