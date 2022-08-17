package clusterresource

import (
	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type strategicMergeFromPatch struct {
	from runtime.Object
}

// Type implements patch.
func (s *strategicMergeFromPatch) Type() types.PatchType {
	return types.StrategicMergePatchType
}

// Data implements Patch.
func (s *strategicMergeFromPatch) Data(obj client.Object) ([]byte, error) {
	originalJSON, err := json.Marshal(s.from)
	if err != nil {
		return nil, err
	}

	modifiedJSON, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return jsonpatch.CreateMergePatch(originalJSON, modifiedJSON)
}

// StrategicMergeFrom creates a Patch using the strategic-merge-patch strategy with the given object as base.
func StrategicMergeFrom(obj runtime.Object) client.Patch {
	return &strategicMergeFromPatch{obj}
}
