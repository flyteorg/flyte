package common

import "github.com/flyteorg/flyte/flytestdlib/contextutils"

var RuntimeTypeKey = contextutils.Key("runtime_type")
var RuntimeVersionKey = contextutils.Key("runtime_version")

const (
	AuditFieldsContextKey contextutils.Key = "audit_fields"
	PrincipalContextKey   contextutils.Key = "principal"
	ErrorKindKey          contextutils.Key = "error_kind"
)

const MaxResponseStatusBytes = 32000

// Grpc proxied over nginx has a status message length limit. Responses with
// status message lengths longer than about 4800 bytes (found experimentally)
// result in a 502 error from nginx. We do not know where this limit is coming
// from, and whether it is impacted by other headers. For now, we are setting
// it to 4 Kib. Error messages longer than this will be truncated.
const MaxErrorResponseStatusBytes = 4 * 1024

// DefaultProducerID is used in older versions of propeller which hard code this producer id.
// See https://github.com/flyteorg/flytepropeller/blob/eaf084934de5d630cd4c11aae15ecae780cc787e/pkg/controller/nodes/task/transformer.go#L114
const DefaultProducerID = "propeller"
const EagerSecretName = "EAGER_DEFAULT"
