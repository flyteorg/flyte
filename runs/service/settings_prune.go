package service

import (
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// hasAnyField reports whether m has any populated field. A message with none
// is exactly one protojson would serialize as {}.
func hasAnyField(m protoreflect.Message) bool {
	found := false
	m.Range(func(protoreflect.FieldDescriptor, protoreflect.Value) bool {
		found = true
		return false
	})
	return found
}

// pruneSettings removes message fields that contain no information, so the
// stored row does not depend on whether a client expressed INHERIT by omitting
// a field or by sending it present but empty. Both forms read back
// identically, so pruning only canonicalizes the stored bytes.
func pruneEmpty(m protoreflect.Message) {
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.Kind() != protoreflect.MessageKind || fd.IsMap() || fd.IsList() {
			return true
		}
		child := v.Message()
		pruneEmpty(child)

		if !hasAnyField(child) {
			m.Clear(fd)
		}
		return true
	})
}

// pruneSettings removes settings messages that carry no information, so a
// stored row does not depend on how the client spelled INHERIT.
func pruneSettings(s *settings.Settings) {
	if s == nil {
		return
	}
	pruneEmpty(s.ProtoReflect())
}
