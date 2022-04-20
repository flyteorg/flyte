package common

import "github.com/flyteorg/flytestdlib/contextutils"

var RuntimeTypeKey = contextutils.Key("runtime_type")
var RuntimeVersionKey = contextutils.Key("runtime_version")

const (
	AuditFieldsContextKey contextutils.Key = "audit_fields"
	PrincipalContextKey   contextutils.Key = "principal"
	ErrorKindKey          contextutils.Key = "error_kind"
)

const MaxResponseStatusBytes = 32000

// DefaultProducerID is used in older versions of propeller which hard code this producer id.
// See https://github.com/flyteorg/flytepropeller/blob/eaf084934de5d630cd4c11aae15ecae780cc787e/pkg/controller/nodes/task/transformer.go#L114
const DefaultProducerID = "propeller"
