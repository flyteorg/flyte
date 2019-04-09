package labeled

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/lyft/flytestdlib/contextutils"
)

var (
	ErrAlreadySet = fmt.Errorf("cannot set metric keys more than once")
	ErrEmpty      = fmt.Errorf("cannot set metric keys to an empty set")
	ErrNeverSet   = fmt.Errorf("must call SetMetricKeys prior to using labeled package")

	// Metric Keys to label metrics with. These keys get pulled from context if they are present. Use contextutils to fill
	// them in.
	metricKeys = make([]contextutils.Key, 0)

	// :(, we have to create a separate list to satisfy the MustNewCounterVec API as it accepts string only
	metricStringKeys = make([]string, 0)
	metricKeysAreSet = sync.Once{}
)

// Sets keys to use with labeled metrics. The values of these keys will be pulled from context at runtime.
func SetMetricKeys(keys ...contextutils.Key) {
	if len(keys) == 0 {
		panic(ErrEmpty)
	}

	ran := false
	metricKeysAreSet.Do(func() {
		ran = true
		metricKeys = keys
		for _, metricKey := range metricKeys {
			metricStringKeys = append(metricStringKeys, metricKey.String())
		}
	})

	if !ran && !reflect.DeepEqual(keys, metricKeys) {
		panic(ErrAlreadySet)
	}
}

func GetUnlabeledMetricName(metricName string) string {
	return metricName + "_unlabeled"
}
