package server

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/lib"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	admin2 "github.com/flyteorg/flyte/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type TriggerEngine struct {
	service     artifact.ArtifactRegistryServer
	adminClient service.AdminServiceClient
	store       StorageInterface

	// needs to be used
	scope promutils.Scope
}

// Evaluate the trigger and launch the workflow if it's true
func (e *TriggerEngine) evaluateAndHandleTrigger(ctx context.Context, trigger models.Trigger, incoming *artifact.Artifact) error {
	incomingKey := core.ArtifactKey{
		Project: incoming.GetArtifactId().GetArtifactKey().GetProject(),
		Domain:  incoming.GetArtifactId().GetArtifactKey().GetDomain(),
		Name:    incoming.GetArtifactId().GetArtifactKey().GetName(),
	}

	var triggeringArtifacts = make([]artifact.Artifact, 0)

	incomingPartitions := map[string]*core.LabelValue{}

	if incoming.GetArtifactId().GetPartitions().GetValue() != nil && len(incoming.GetArtifactId().GetPartitions().GetValue()) > 0 {
		for k, p := range incoming.GetArtifactId().GetPartitions().GetValue() {
			if len(p.GetStaticValue()) == 0 {
				logger.Warningf(ctx, "Trigger %s has non-static partition [%+v]", trigger.Name, incoming.GetArtifactId().GetPartitions().GetValue())
				return fmt.Errorf("trigger %s has non-static partition %s [%+v]", trigger.Name, k, p)
			}
			incomingPartitions[k] = p
		}
	}

	// Note the order of this is important. It needs to be the same order as the trigger.RunsOn
	// This is because binding data references an index in this array.
	for _, triggeringArtifactID := range trigger.RunsOn {
		// First check partitions.
		// They must either both have no partitions, or both have the same partitions.
		var thisIDPartitions = make(map[string]string)
		if triggeringArtifactID.GetPartitions().GetValue() != nil {
			for k := range triggeringArtifactID.GetPartitions().GetValue() {
				thisIDPartitions[k] = "placeholder"
			}
		}
		if len(thisIDPartitions) != len(incomingPartitions) {
			return fmt.Errorf("trigger %s has different number of partitions [%v] [%+v]", trigger.Name, incoming.GetArtifactId(), triggeringArtifactID)
		}

		// If the lengths match, they must still also have the same keys.
		// Build a query map of partitions for this triggering artifact while at it.
		queryPartitions := map[string]*core.LabelValue{}
		if len(thisIDPartitions) > 0 {
			for k := range thisIDPartitions {
				if incomingValue, ok := incomingPartitions[k]; !ok {
					return fmt.Errorf("trigger %s has different partitions [%v] [%+v]", trigger.Name, incoming.GetArtifactId(), triggeringArtifactID)
				} else {
					queryPartitions[k] = incomingValue
				}
			}
		}

		// See if it's the same one as incoming.
		if triggeringArtifactID.GetArtifactKey().Project == incomingKey.Project &&
			triggeringArtifactID.GetArtifactKey().Domain == incomingKey.Domain &&
			triggeringArtifactID.GetArtifactKey().Name == incomingKey.Name {
			triggeringArtifacts = append(triggeringArtifacts, *incoming)
			continue
		}

		// Otherwise, assume it's a different one
		// Construct a query and fetch it
		var lookupID = core.ArtifactID{
			ArtifactKey: triggeringArtifactID.ArtifactKey,
		}
		if len(queryPartitions) > 0 {
			lookupID.Dimensions = &core.ArtifactID_Partitions{
				Partitions: &core.Partitions{Value: queryPartitions},
			}
		}
		query := core.ArtifactQuery{
			Identifier: &core.ArtifactQuery_ArtifactId{
				ArtifactId: &lookupID,
			},
		}

		resp, err := e.service.GetArtifact(ctx, &artifact.GetArtifactRequest{
			Query: &query,
		})
		if err != nil {
			return fmt.Errorf("failed to get artifact [%+v]: %w", lookupID, err)
		}

		triggeringArtifacts = append(triggeringArtifacts, *resp.Artifact)

	}

	err := e.createExecution(
		ctx,
		triggeringArtifacts,
		trigger.LaunchPlan.Id,
		trigger.LaunchPlan.Spec.DefaultInputs,
	)

	return err
}

func (e *TriggerEngine) generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	charset := "abcdefghijklmnopqrstuvwxyz"
	var result string
	for i := 0; i < length; i++ {
		randomIndex := rand.Intn(len(charset))
		result += string(charset[randomIndex])
	}

	return result
}

func (e *TriggerEngine) getSpec(_ context.Context, launchPlanID *core.Identifier) admin.ExecutionSpec {

	var spec = admin.ExecutionSpec{
		LaunchPlan:            launchPlanID,
		Metadata:              nil,
		NotificationOverrides: nil,
		Labels:                nil,
		Annotations:           nil,
		SecurityContext:       nil,
	}
	return spec
}

func (e *TriggerEngine) createExecution(ctx context.Context, triggeringArtifacts []artifact.Artifact, launchPlanID *core.Identifier, defaultInputs *core.ParameterMap) error {

	resolvedInputs, err := e.resolveInputs(ctx, triggeringArtifacts, defaultInputs, launchPlanID)
	if err != nil {
		return fmt.Errorf("failed to resolve inputs: %w", err)
	}

	spec := e.getSpec(ctx, launchPlanID)

	resp, err := e.adminClient.CreateExecution(ctx, &admin.ExecutionCreateRequest{
		Project: launchPlanID.Project,
		Domain:  launchPlanID.Domain,
		Name:    e.generateRandomString(12),
		Spec:    &spec,
		Inputs:  &core.LiteralMap{Literals: resolvedInputs},
	})
	if err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}

	logger.Infof(ctx, "Created execution %v", resp)
	return nil
}

func (e *TriggerEngine) resolveInputs(ctx context.Context, triggeringArtifacts []artifact.Artifact, defaultInputs *core.ParameterMap, launchPlanID *core.Identifier) (map[string]*core.Literal, error) {
	// Process inputs that have defaults separately as these may be used to fill in other inputs
	var defaults = map[string]*core.Literal{}
	for k, v := range defaultInputs.Parameters {
		if v.GetDefault() != nil {
			defaults[k] = v.GetDefault()
		}
	}

	var inputs = map[string]*core.Literal{}
	for k, v := range defaultInputs.Parameters {
		if v.GetDefault() != nil {
			continue
		}
		if v == nil {
			return nil, fmt.Errorf("parameter [%s] is nil", k)
		}

		convertedLiteral, err := e.convertParameterToLiteral(ctx, triggeringArtifacts, *v, defaults, launchPlanID)
		if err != nil {
			logger.Errorf(ctx, "Error converting parameter [%s] [%v] to literal: %v", k, v, err)
			return nil, err
		}
		inputs[k] = &convertedLiteral
	}

	for k, v := range inputs {
		defaults[k] = v
	}
	return defaults, nil
}

func (e *TriggerEngine) convertParameterToLiteral(ctx context.Context, triggeringArtifacts []artifact.Artifact,
	value core.Parameter, otherInputs map[string]*core.Literal, launchPlanID *core.Identifier) (core.Literal, error) {

	if value.GetArtifactQuery() != nil {
		return e.convertArtifactQueryToLiteral(ctx, triggeringArtifacts, *value.GetArtifactQuery(), otherInputs, value, launchPlanID)

	} else if value.GetArtifactId() != nil {
		return core.Literal{}, fmt.Errorf("artifact id not supported yet")
	}
	return core.Literal{}, fmt.Errorf("trying to convert non artifact Parameter to Literal")

}

var DurationRegex = regexp.MustCompile(`P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+S)?`)

func ParseDuration(str string) time.Duration {
	matches := DurationRegex.FindStringSubmatch(str)

	years := ParseInt64(matches[1])
	months := ParseInt64(matches[2])
	days := ParseInt64(matches[3])
	hours := ParseInt64(matches[4])
	minutes := ParseInt64(matches[5])
	seconds := ParseInt64(matches[6])

	hour := int64(time.Hour)
	minute := int64(time.Minute)
	second := int64(time.Second)
	return time.Duration(years*24*365*hour + months*30*24*hour + days*24*hour + hours*hour + minutes*minute + seconds*second)
}

func ParseInt64(value string) int64 {
	if len(value) == 0 {
		return 0
	}
	parsed, err := strconv.Atoi(value[:len(value)-1])
	if err != nil {
		return 0
	}
	return int64(parsed)
}

func (e *TriggerEngine) parseStringAsDateAndApplyTransform(ctx context.Context, dateFormat string, dateStr string, transform string) (time.Time, error) {
	t, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date [%s] as date : %w", dateStr, err)
	}

	if transform != "" {
		firstRune, runeSize := utf8.DecodeRuneInString(transform)
		op := string(firstRune)
		transformStr := transform[runeSize:]

		if op == "-" {
			td := ParseDuration(transformStr)
			t = t.Add(-td)
			logger.Infof(ctx, "Applying transform [%s] new date is [%s] from %s", transform, t, transformStr)
		} else {
			logger.Warningf(ctx, "Ignoring transform [%s]", transformStr)
		}
	}
	return t, nil
}

func (e *TriggerEngine) convertArtifactQueryToLiteral(ctx context.Context, triggeringArtifacts []artifact.Artifact,
	aq core.ArtifactQuery, otherInputs map[string]*core.Literal, p core.Parameter, launchPlanID *core.Identifier) (core.Literal, error) {

	// if it's a binding and it has a partition key
	// this is the only time we need to transform the type.
	// and it's only ever going to be a datetime
	if bnd := aq.GetBinding(); bnd != nil {
		a := triggeringArtifacts[bnd.GetIndex()]
		if bnd.GetPartitionKey() != "" {
			if partitions := a.GetArtifactId().GetPartitions().GetValue(); partitions != nil {

				if sv, ok := partitions[bnd.GetPartitionKey()]; ok {
					dateStr := sv.GetStaticValue()

					t, err := e.parseStringAsDateAndApplyTransform(ctx, lib.DateFormat, dateStr, bnd.GetTransform())
					if err != nil {
						return core.Literal{}, fmt.Errorf("failed to parse [%s] transform %s] for [%+v]: %w", dateStr, bnd.GetTransform(), a.GetArtifactId(), err)
					}

					// convert time to timestamp
					ts := timestamppb.New(t)

					return core.Literal{
						Value: &core.Literal_Scalar{
							Scalar: &core.Scalar{
								Value: &core.Scalar_Primitive{
									Primitive: &core.Primitive{
										Value: &core.Primitive_Datetime{
											Datetime: ts,
										},
									},
								},
							},
						},
					}, nil
				}
			}
			return core.Literal{}, fmt.Errorf("partition key [%s] not found in artifact [%+v]", bnd.GetPartitionKey(), a.GetArtifactId())
		}

		// this means it's bound to the whole triggering artifact, not a partition. Just pull out the literal and use it
		idlLit := triggeringArtifacts[bnd.GetIndex()].GetSpec().GetValue()
		return *idlLit, nil

		// this is a real query - will need to first iterate through the partitions to see if we need to fill in any
	} else if queryID := aq.GetArtifactId(); queryID != nil {
		searchPartitions := map[string]string{}
		if len(queryID.GetPartitions().GetValue()) > 0 {
			for partitionKey, lv := range queryID.GetPartitions().GetValue() {
				if lv.GetStaticValue() != "" {
					searchPartitions[partitionKey] = lv.GetStaticValue()

				} else if lv.GetInputBinding() != nil {
					inputVar := lv.GetInputBinding().GetVar()
					if val, ok := otherInputs[inputVar]; ok {
						strVal, err := lib.RenderLiteral(val)
						if err != nil {
							return core.Literal{}, fmt.Errorf("failed to render input [%s] for partition [%s] with error: %w", inputVar, partitionKey, err)
						}
						searchPartitions[partitionKey] = strVal
					} else {
						return core.Literal{}, fmt.Errorf("input binding [%s] not found in input data", inputVar)
					}

					// This is like AnArtifact.query(time_partition=Upstream.time_partition - timedelta(days=1)) or
					//   AnArtifact.query(region=Upstream.region)
				} else if triggerBinding := lv.GetTriggeredBinding(); triggerBinding != nil {
					a := triggeringArtifacts[triggerBinding.GetIndex()]
					aP := a.GetArtifactId().GetPartitions().GetValue()
					var searchValue = aP[triggerBinding.GetPartitionKey()].GetStaticValue()

					if triggerBinding.GetTransform() != "" {
						logger.Infof(ctx, "Transform detected [%s] value [%s] with transform [%s], assuming datetime",
							triggerBinding.GetPartitionKey(), searchValue, triggerBinding.GetTransform())
						t, err := e.parseStringAsDateAndApplyTransform(ctx, lib.DateFormat, searchValue, triggerBinding.GetTransform())
						if err != nil {
							return core.Literal{}, fmt.Errorf("failed to parse [%s] transform %s] for [%+v]: %w", searchValue, triggerBinding.GetTransform(), a.GetArtifactId(), err)
						}
						logger.Debugf(ctx, "Transformed [%s] to [%s]", aP[triggerBinding.GetPartitionKey()].GetStaticValue(), searchValue)
						searchValue = t.Format(lib.DateFormat)
					}
					searchPartitions[partitionKey] = searchValue
				}
			}
		}

		artifactID := core.ArtifactID{
			ArtifactKey: &core.ArtifactKey{
				Project: launchPlanID.Project,
				Domain:  launchPlanID.Domain,
				Name:    queryID.ArtifactKey.Name,
			},
		}
		if len(searchPartitions) > 0 {
			artifactID.Dimensions = &core.ArtifactID_Partitions{
				Partitions: models.PartitionsToIdl(searchPartitions),
			}
		}

		resp, err := e.service.GetArtifact(ctx, &artifact.GetArtifactRequest{
			Query: &core.ArtifactQuery{
				Identifier: &core.ArtifactQuery_ArtifactId{
					ArtifactId: &artifactID,
				},
			},
		})
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				if len(p.GetVar().GetType().GetUnionType().GetVariants()) > 0 {
					for _, v := range p.GetVar().GetType().GetUnionType().GetVariants() {
						if v.GetSimple() == core.SimpleType_NONE {
							return core.Literal{
								Value: &core.Literal_Scalar{
									Scalar: &core.Scalar{
										Value: &core.Scalar_NoneType{
											NoneType: &core.Void{},
										},
									},
								},
							}, nil
						}
					}
				}
			}
		}

		return *resp.Artifact.Spec.Value, nil
	}

	return core.Literal{}, fmt.Errorf("query was neither binding nor artifact id [%v]", aq)
}

func (e *TriggerEngine) EvaluateNewArtifact(ctx context.Context, artifact *artifact.Artifact) ([]core.WorkflowExecutionIdentifier, error) {

	if artifact.GetArtifactId().GetArtifactKey() == nil {
		// metric
		return nil, fmt.Errorf("artifact or its key cannot be nil")
	}

	triggers, err := e.store.GetTriggersByArtifactKey(ctx, *artifact.ArtifactId.ArtifactKey)
	if err != nil {
		logger.Errorf(ctx, "Failed to get triggers for artifact [%+v]: %v", artifact.ArtifactId.ArtifactKey, err)
		return nil, err
	}

	for _, trigger := range triggers {
		err := e.evaluateAndHandleTrigger(ctx, trigger, artifact)
		if err != nil {
			logger.Errorf(ctx, "Failed to evaluate trigger [%s]: %v", trigger.Name, err)
			return nil, err
		}
	}

	// todo: capture return IDs
	return nil, nil
}

func NewTriggerEngine(ctx context.Context, storage StorageInterface, service artifact.ArtifactRegistryServer, scope promutils.Scope) (TriggerEngine, error) {
	cfg := admin2.GetConfig(ctx)
	clients, err := admin2.NewClientsetBuilder().WithConfig(cfg).Build(ctx)
	if err != nil {
		return TriggerEngine{}, fmt.Errorf("failed to initialize clientset. Error: %w", err)
	}

	return TriggerEngine{
		service:     service,
		adminClient: clients.AdminClient(),
		store:       storage,
		scope:       scope,
	}, nil
}
