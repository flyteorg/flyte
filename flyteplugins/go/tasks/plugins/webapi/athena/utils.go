package athena

import (
	"context"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/template"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"

	pluginsIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flytestdlib/utils"

	pb "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flytestdlib/logger"
)

func writeOutput(ctx context.Context, tCtx webapi.StatusContext, externalLocation string) error {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return err
	}

	if taskTemplate.Interface == nil || taskTemplate.Interface.Outputs == nil || taskTemplate.Interface.Outputs.Variables == nil {
		logger.Infof(ctx, "The task declares no outputs. Skipping writing the outputs.")
		return nil
	}

	resultsSchema, exists := taskTemplate.Interface.Outputs.Variables["results"]
	if !exists {
		logger.Infof(ctx, "The task declares no outputs. Skipping writing the outputs.")
		return nil
	}

	return tCtx.OutputWriter().Put(ctx, ioutils.NewInMemoryOutputReader(
		&pb.LiteralMap{
			Literals: map[string]*pb.Literal{
				"results": {
					Value: &pb.Literal_Scalar{
						Scalar: &pb.Scalar{Value: &pb.Scalar_Schema{
							Schema: &pb.Schema{
								Uri:  externalLocation,
								Type: resultsSchema.GetType().GetSchema(),
							},
						},
						},
					},
				},
			},
		}, nil, nil))
}

type QueryInfo struct {
	QueryString string
	Workgroup   string
	Catalog     string
	Database    string
}

func validateHiveQuery(hiveQuery pluginsIdl.QuboleHiveJob) error {
	if hiveQuery.Query == nil {
		return errors.Errorf(errors.BadTaskSpecification, "Query is a required field.")
	}

	if len(hiveQuery.Query.Query) == 0 {
		return errors.Errorf(errors.BadTaskSpecification, "Query statement is a required field.")
	}

	return nil
}

func validatePrestoQuery(prestoQuery pluginsIdl.PrestoQuery) error {
	if len(prestoQuery.Statement) == 0 {
		return errors.Errorf(errors.BadTaskSpecification, "Statement is a required field.")
	}

	return nil
}

func extractQueryInfo(ctx context.Context, tCtx webapi.TaskExecutionContextReader) (QueryInfo, error) {
	task, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return QueryInfo{}, err
	}

	switch task.Type {
	case "hive":
		custom := task.GetCustom()
		hiveQuery := pluginsIdl.QuboleHiveJob{}
		err := utils.UnmarshalStructToPb(custom, &hiveQuery)
		if err != nil {
			return QueryInfo{}, errors.Wrapf(ErrUser, err, "Expects a valid QubleHiveJob proto in custom field.")
		}

		if err = validateHiveQuery(hiveQuery); err != nil {
			return QueryInfo{}, errors.Wrapf(ErrUser, err, "Expects a valid QubleHiveJob proto in custom field.")
		}

		outputs, err := template.Render(ctx, []string{
			hiveQuery.Query.Query,
			hiveQuery.ClusterLabel,
		}, template.Parameters{
			TaskExecMetadata: tCtx.TaskExecutionMetadata(),
			Inputs:           tCtx.InputReader(),
			OutputPath:       tCtx.OutputWriter(),
			Task:             tCtx.TaskReader(),
		})
		if err != nil {
			return QueryInfo{}, err
		}

		return QueryInfo{
			QueryString: outputs[0],
			Database:    outputs[1],
		}, nil
	case "presto":
		custom := task.GetCustom()
		prestoQuery := pluginsIdl.PrestoQuery{}
		err := utils.UnmarshalStructToPb(custom, &prestoQuery)
		if err != nil {
			return QueryInfo{}, errors.Wrapf(ErrUser, err, "Expects a valid PrestoQuery proto in custom field.")
		}

		if err = validatePrestoQuery(prestoQuery); err != nil {
			return QueryInfo{}, errors.Wrapf(ErrUser, err, "Expects a valid PrestoQuery proto in custom field.")
		}

		outputs, err := template.Render(ctx, []string{
			prestoQuery.RoutingGroup,
			prestoQuery.Catalog,
			prestoQuery.Schema,
			prestoQuery.Statement,
		}, template.Parameters{
			TaskExecMetadata: tCtx.TaskExecutionMetadata(),
			Inputs:           tCtx.InputReader(),
			OutputPath:       tCtx.OutputWriter(),
			Task:             tCtx.TaskReader(),
		})
		if err != nil {
			return QueryInfo{}, err
		}

		return QueryInfo{
			Workgroup:   outputs[0],
			Catalog:     outputs[1],
			Database:    outputs[2],
			QueryString: outputs[3],
		}, nil
	}

	return QueryInfo{}, errors.Errorf(ErrUser, "Unexpected task type [%v].", task.Type)
}
