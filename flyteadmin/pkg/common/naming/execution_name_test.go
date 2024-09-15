package naming

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyte/flyteadmin/scheduler/identifier"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const AllowedExecutionIDAlphabetStr = "abcdefghijklmnopqrstuvwxyz"
const AllowedExecutionIDAlphanumericStr = "abcdefghijklmnopqrstuvwxyz1234567890"
const AllowedExecutionIDFriendlyNameStr = "abcdefghijklmnopqrstuvwxyz-"

var AllowedExecutionIDAlphabets = []rune(AllowedExecutionIDAlphabetStr)
var AllowedExecutionIDAlphanumerics = []rune(AllowedExecutionIDAlphanumericStr)
var AllowedExecutionIDFriendlyNameChars = []rune(AllowedExecutionIDFriendlyNameStr)

func TestGetExecutionName(t *testing.T) {
	originalConfigProvider := configProvider
	defer func() { configProvider = originalConfigProvider }()

	mockConfigProvider := &runtimeMocks.MockApplicationProvider{}
	configProvider = mockConfigProvider

	t.Run("general name", func(t *testing.T) {
		appConfig := runtimeInterfaces.ApplicationConfig{
			FeatureGates: runtimeInterfaces.FeatureGates{
				EnableFriendlyNames: false,
			},
		}
		mockConfigProvider.SetTopLevelConfig(appConfig)

		randString := GetExecutionName(time.Now().UnixNano())
		assert.Len(t, randString, ExecutionIDLength)
		assert.Contains(t, AllowedExecutionIDAlphabets, rune(randString[0]))
		for i := 1; i < len(randString); i++ {
			assert.Contains(t, AllowedExecutionIDAlphanumerics, rune(randString[i]))
		}
	})

	t.Run("friendly name", func(t *testing.T) {
		appConfig := runtimeInterfaces.ApplicationConfig{
			FeatureGates: runtimeInterfaces.FeatureGates{
				EnableFriendlyNames: true,
			},
		}
		mockConfigProvider.SetTopLevelConfig(appConfig)

		randString := GetExecutionName(time.Now().UnixNano())
		assert.LessOrEqual(t, len(randString), ExecutionIDLengthLimit)
		for i := 0; i < len(randString); i++ {
			assert.Contains(t, AllowedExecutionIDFriendlyNameChars, rune(randString[i]))
		}
		hyphenCount := strings.Count(randString, "-")
		assert.Equal(t, 3, hyphenCount, "FriendlyName should contain exactly three hyphens")
		words := strings.Split(randString, "-")
		assert.Equal(t, 4, len(words), "FriendlyName should be split into exactly four words")
	})

	t.Run("deterministic name", func(t *testing.T) {
		hashValue := identifier.HashScheduledTimeStamp(context.Background(), &core.Identifier{
			Project: "Project",
			Domain:  "Domain",
			Name:    "Name",
			Version: "Version",
		}, time.Time{})

		name := GetExecutionName(int64(hashValue))
		assert.Equal(t, name, "carpet-juliet-kentucky-kentucky")
	})
}
