package naming

import (
	"fmt"

	"github.com/wolfeidau/humanhash"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
)

const ExecutionIDLength = 20
const ExecutionIDLengthLimit = 63
const ExecutionStringFormat = "a%s"

var configProvider runtimeInterfaces.ApplicationConfiguration = runtime.NewApplicationConfigurationProvider()

/* #nosec */
func GetExecutionName(seed int64) string {
	rand.Seed(seed)
	config := configProvider.GetTopLevelConfig()
	if config.FeatureGates.EnableFriendlyNames {
		hashKey := []byte(rand.String(ExecutionIDLength))
		// Ignoring the error as it's guaranteed hash key longer than result in this context.
		result, _ := humanhash.Humanize(hashKey, 4)
		return result
	}
	return fmt.Sprintf(ExecutionStringFormat, rand.String(ExecutionIDLength-1))
}
