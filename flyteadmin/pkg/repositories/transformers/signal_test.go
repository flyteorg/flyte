package transformers

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"

	"github.com/golang/protobuf/proto"
)

var (
	booleanType = core.LiteralType{
		Type: &core.LiteralType_Simple{
			Simple: core.SimpleType_BOOLEAN,
		},
	}

	booleanValue = core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_Boolean{
							Boolean: false,
						},
					},
				},
			},
		},
	}

	signalKey = models.SignalKey{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		SignalID: "signal",
	}

	signalID = core.SignalIdentifier{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		SignalId: "signal",
	}
)

func TestCreateSignalModel(t *testing.T) {
	booleanTypeBytes, _ := proto.Marshal(&booleanType)
	booleanValueBytes, _ := proto.Marshal(&booleanValue)

	tests := []struct {
		name  string
		model models.Signal
		proto admin.Signal
	}{
		{
			name:  "Empty",
			model: models.Signal{},
			proto: admin.Signal{},
		},
		{
			name: "Full",
			model: models.Signal{
				SignalKey: signalKey,
				Type:      booleanTypeBytes,
				Value:     booleanValueBytes,
			},
			proto: admin.Signal{
				Id:    &signalID,
				Type:  &booleanType,
				Value: &booleanValue,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			signalModel, err := CreateSignalModel(test.proto.Id, test.proto.Type, test.proto.Value)
			assert.NoError(t, err)

			assert.Equal(t, test.model, signalModel)
		})
	}
}

func TestFromSignalModel(t *testing.T) {
	booleanTypeBytes, _ := proto.Marshal(&booleanType)
	booleanValueBytes, _ := proto.Marshal(&booleanValue)

	tests := []struct {
		name  string
		model models.Signal
		proto admin.Signal
	}{
		{
			name:  "Empty",
			model: models.Signal{},
			proto: admin.Signal{},
		},
		{
			name: "Full",
			model: models.Signal{
				SignalKey: signalKey,
				Type:      booleanTypeBytes,
				Value:     booleanValueBytes,
			},
			proto: admin.Signal{
				Id:    &signalID,
				Type:  &booleanType,
				Value: &booleanValue,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			signal, err := FromSignalModel(test.model)
			assert.NoError(t, err)

			assert.True(t, proto.Equal(&test.proto, &signal))
		})
	}
}

func TestFromSignalModels(t *testing.T) {
	booleanTypeBytes, _ := proto.Marshal(&booleanType)
	booleanValueBytes, _ := proto.Marshal(&booleanValue)

	signalModels := []models.Signal{
		models.Signal{},
		models.Signal{
			SignalKey: signalKey,
			Type:      booleanTypeBytes,
			Value:     booleanValueBytes,
		},
	}

	signals := []*admin.Signal{
		&admin.Signal{},
		&admin.Signal{
			Id:    &signalID,
			Type:  &booleanType,
			Value: &booleanValue,
		},
	}

	s, err := FromSignalModels(signalModels)
	assert.NoError(t, err)

	assert.Len(t, s, len(signals))
	for idx, signal := range signals {
		assert.True(t, proto.Equal(signal, s[idx]))
	}
}
