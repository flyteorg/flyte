package array

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/tests"
	"github.com/flyteorg/flyte/flytestdlib/utils"
	idlCore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

type AdvanceIteration func(ctx context.Context, tCtx core.TaskExecutionContext) error

func RunArrayTestsEndToEnd(t *testing.T, executor core.Plugin, iter AdvanceIteration) {
	inputs := coreutils.MustMakeLiteral(map[string]interface{}{
		"x": 5,
	}).GetMap()

	t.Run("Regular container task", func(t *testing.T) {
		template := tests.BuildTaskTemplate()
		tests.RunPluginEndToEndTest(t, executor, template, inputs, nil, nil, iter)
	})

	t.Run("Array of size 1. No cache", func(t *testing.T) {
		template := tests.BuildTaskTemplate()
		template.Interface = &idlCore.TypedInterface{
			Inputs: nil,
			Outputs: &idlCore.VariableMap{
				Variables: map[string]*idlCore.Variable{
					"x": {
						Type: &idlCore.LiteralType{
							Type: &idlCore.LiteralType_CollectionType{
								CollectionType: &idlCore.LiteralType{Type: &idlCore.LiteralType_Simple{Simple: idlCore.SimpleType_INTEGER}},
							},
						},
					},
				},
			},
		}

		var err error
		template.Custom, err = utils.MarshalPbToStruct(&plugins.ArrayJob{
			Parallelism: 10,
			Size:        1,
			SuccessCriteria: &plugins.ArrayJob_MinSuccesses{
				MinSuccesses: 1,
			},
		})

		assert.NoError(t, err)

		expectedOutputs := coreutils.MustMakeLiteral(map[string]interface{}{
			"x": []interface{}{5},
		}).GetMap()

		tests.RunPluginEndToEndTest(t, executor, template, inputs, expectedOutputs, nil, iter)
	})

	t.Run("Array of size 2. No cache", func(t *testing.T) {
		template := tests.BuildTaskTemplate()
		template.Interface = &idlCore.TypedInterface{
			Inputs: nil,
			Outputs: &idlCore.VariableMap{
				Variables: map[string]*idlCore.Variable{
					"x": {
						Type: &idlCore.LiteralType{
							Type: &idlCore.LiteralType_CollectionType{
								CollectionType: &idlCore.LiteralType{Type: &idlCore.LiteralType_Simple{Simple: idlCore.SimpleType_INTEGER}},
							},
						},
					},
				},
			},
		}

		var err error
		template.Custom, err = utils.MarshalPbToStruct(&plugins.ArrayJob{
			Parallelism: 10,
			Size:        2,
			SuccessCriteria: &plugins.ArrayJob_MinSuccesses{
				MinSuccesses: 1,
			},
		})

		assert.NoError(t, err)

		expectedOutputs := coreutils.MustMakeLiteral(map[string]interface{}{
			"x": []interface{}{5, 5},
		}).GetMap()

		tests.RunPluginEndToEndTest(t, executor, template, inputs, expectedOutputs, nil, iter)
	})
}
