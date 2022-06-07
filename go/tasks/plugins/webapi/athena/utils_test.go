package athena

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flytestdlib/storage"

	mocks3 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	pb "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flytestdlib/utils"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi/mocks"
)

func Test_writeOutput(t *testing.T) {
	ctx := context.Background()
	t.Run("No Outputs", func(t *testing.T) {
		taskReader := &mocks2.TaskReader{}
		taskReader.OnRead(ctx).Return(&core.TaskTemplate{}, nil)

		statusContext := &mocks.StatusContext{}
		statusContext.OnTaskReader().Return(taskReader)

		err := writeOutput(context.Background(), statusContext, "s3://my-external-bucket/key")
		assert.NoError(t, err)
	})

	t.Run("No Output named results", func(t *testing.T) {
		taskReader := &mocks2.TaskReader{}
		taskReader.OnRead(ctx).Return(&core.TaskTemplate{
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"myOutput": &core.Variable{},
					},
				},
			},
		}, nil)

		statusContext := &mocks.StatusContext{}
		statusContext.OnTaskReader().Return(taskReader)

		err := writeOutput(context.Background(), statusContext, "s3://my-external-bucket/key")
		assert.NoError(t, err)
	})

	t.Run("Valid Qubole", func(t *testing.T) {
		statusContext := &mocks.StatusContext{}
		taskReader := &mocks2.TaskReader{}
		hive := &plugins.QuboleHiveJob{
			ClusterLabel: "mydb",
			Query: &plugins.HiveQuery{
				Query: "Select * from mytable",
			},
		}

		st, err := utils.MarshalPbToStruct(hive)
		if !assert.NoError(t, err) {
			assert.FailNowf(t, "expected to be able to marshal", "")
		}

		taskReader.OnRead(ctx).Return(&core.TaskTemplate{
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"results": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Schema{
									Schema: &core.SchemaType{
										Columns: []*core.SchemaType_SchemaColumn{},
									},
								},
							},
						},
					},
				},
			},
			Custom: st,
		}, nil)

		statusContext.OnTaskReader().Return(taskReader)

		ow := &mocks3.OutputWriter{}
		externalLocation := "s3://my-external-bucket/key"
		ow.OnPut(ctx, ioutils.NewInMemoryOutputReader(
			&pb.LiteralMap{
				Literals: map[string]*pb.Literal{
					"results": {
						Value: &pb.Literal_Scalar{
							Scalar: &pb.Scalar{
								Value: &pb.Scalar_Schema{
									Schema: &pb.Schema{
										Uri: externalLocation,
										Type: &core.SchemaType{
											Columns: []*core.SchemaType_SchemaColumn{},
										},
									},
								},
							},
						},
					},
				},
			}, nil, nil)).Return(nil)
		statusContext.OnOutputWriter().Return(ow)

		err = writeOutput(context.Background(), statusContext, externalLocation)
		assert.NoError(t, err)
	})
}

func Test_ExtractQueryInfo(t *testing.T) {
	ctx := context.Background()
	validProtos := []struct {
		message  proto.Message
		taskType string
	}{
		{
			message: &plugins.QuboleHiveJob{
				ClusterLabel: "mydb",
				Query: &plugins.HiveQuery{
					Query: "Select * from mytable",
				},
			},
			taskType: "hive",
		},
		{
			message: &plugins.PrestoQuery{
				Statement:    "Select * from mytable",
				Schema:       "mytable",
				RoutingGroup: "primary",
				Catalog:      "catalog",
			},
			taskType: "presto",
		},
	}

	for _, validProto := range validProtos {
		t.Run(fmt.Sprintf("Valid %v", validProto.taskType), func(t *testing.T) {
			tCtx := &mocks.TaskExecutionContextReader{}
			taskReader := &mocks2.TaskReader{}
			st, err := utils.MarshalPbToStruct(validProto.message)
			if !assert.NoError(t, err) {
				assert.FailNowf(t, "expected to be able to marshal", "")
			}

			taskReader.OnRead(ctx).Return(&core.TaskTemplate{
				Type: validProto.taskType,
				Interface: &core.TypedInterface{
					Outputs: &core.VariableMap{
						Variables: map[string]*core.Variable{
							"results": {
								Type: &core.LiteralType{
									Type: &core.LiteralType_Schema{
										Schema: &core.SchemaType{
											Columns: []*core.SchemaType_SchemaColumn{},
										},
									},
								},
							},
						},
					},
				},
				Custom: st,
			}, nil)

			tCtx.OnTaskReader().Return(taskReader)

			tMeta := &mocks2.TaskExecutionMetadata{}
			tCtx.OnTaskExecutionMetadata().Return(tMeta)

			tID := &mocks2.TaskExecutionID{}
			tMeta.OnGetTaskExecutionID().Return(tID)

			tID.OnGetGeneratedName().Return("generated-name")

			ow := &mocks3.OutputWriter{}
			tCtx.OnOutputWriter().Return(ow)
			ow.OnGetOutputPrefixPath().Return("s3://another")
			ow.OnGetRawOutputPrefix().Return("s3://another/output")
			ow.OnGetCheckpointPrefix().Return("/checkpoint")
			ow.OnGetPreviousCheckpointsPrefix().Return("/prev")

			ir := &mocks3.InputReader{}
			tCtx.OnInputReader().Return(ir)
			ir.OnGetInputPath().Return(storage.DataReference("s3://something"))
			ir.OnGetInputPrefixPath().Return(storage.DataReference("s3://something/2"))
			ir.OnGet(ctx).Return(nil, nil)

			q, err := extractQueryInfo(ctx, tCtx)
			assert.NoError(t, err)
			assert.True(t, len(q.QueryString) > 0)
		})
	}
}
