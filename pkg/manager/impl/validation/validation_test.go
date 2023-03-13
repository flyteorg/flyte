package validation

import (
	"fmt"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestGetMissingArgumentError(t *testing.T) {
	err := shared.GetMissingArgumentError("foo")
	assert.EqualError(t, err, "missing foo")
	assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
}

func TestValidateMaxLengthStringField(t *testing.T) {
	err := ValidateMaxLengthStringField("abcdefg", "foo", 6)
	assert.EqualError(t, err, "foo cannot exceed 6 characters")
	assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
}

func TestValidateMaxMapLengthField(t *testing.T) {
	labels := map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	}
	err := ValidateMaxMapLengthField(labels, "foo", 2)
	assert.EqualError(t, err, "foo map cannot exceed 2 entries")
	assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
}

func TestValidateIdentifier(t *testing.T) {
	err := ValidateIdentifier(&core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Domain:       "domain",
		Name:         "name",
	}, common.Task)
	assert.EqualError(t, err, "missing project")

	err = ValidateIdentifier(&core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Name:         "name",
	}, common.Task)
	assert.EqualError(t, err, "missing domain")

	err = ValidateIdentifier(&core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "domain",
	}, common.Task)
	assert.EqualError(t, err, "missing name")

	err = ValidateIdentifier(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
	}, common.Task)
	assert.EqualError(t, err, "unexpected resource type workflow for identifier "+
		"[resource_type:WORKFLOW project:\"project\" domain:\"domain\" ], expected task instead")
}

func TestValidateNamedEntityIdentifierListRequest(t *testing.T) {
	assert.Nil(t, ValidateNamedEntityIdentifierListRequest(admin.NamedEntityIdentifierListRequest{
		Project: "project",
		Domain:  "domain",
		Limit:   2,
	}))

	assert.NotNil(t, ValidateNamedEntityIdentifierListRequest(admin.NamedEntityIdentifierListRequest{
		Domain: "domain",
		Limit:  2,
	}))

	assert.NotNil(t, ValidateNamedEntityIdentifierListRequest(admin.NamedEntityIdentifierListRequest{
		Project: "project",
		Limit:   2,
	}))

	assert.NotNil(t, ValidateNamedEntityIdentifierListRequest(admin.NamedEntityIdentifierListRequest{
		Project: "project",
		Domain:  "domain",
	}))
}

func TestValidateDescriptionEntityIdentifierGetRequest(t *testing.T) {
	assert.Nil(t, ValidateDescriptionEntityGetRequest(admin.ObjectGetRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
			Version:      "v1",
		},
	}))

	assert.NotNil(t, ValidateDescriptionEntityGetRequest(admin.ObjectGetRequest{
		Id: &core.Identifier{
			Project: "project",
		},
	}))

	assert.NotNil(t, ValidateDescriptionEntityGetRequest(admin.ObjectGetRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "project",
		},
	}))
}

func TestValidateDescriptionEntityListRequest(t *testing.T) {
	assert.Nil(t, ValidateDescriptionEntityListRequest(admin.DescriptionEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Limit: 1,
	}))

	assert.NotNil(t, ValidateDescriptionEntityListRequest(admin.DescriptionEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	}))

	assert.NotNil(t, ValidateDescriptionEntityListRequest(admin.DescriptionEntityListRequest{
		Id: nil,
	}))

	assert.NotNil(t, ValidateDescriptionEntityListRequest(admin.DescriptionEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id:           nil,
	}))

	assert.NotNil(t, ValidateDescriptionEntityListRequest(admin.DescriptionEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Domain: "domain",
		},
	}))

	assert.NotNil(t, ValidateDescriptionEntityListRequest(admin.DescriptionEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
		},
	}))

	assert.NotNil(t, ValidateDescriptionEntityListRequest(admin.DescriptionEntityListRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
		},
	}))
}

func TestValidateVersion(t *testing.T) {
	err := ValidateVersion("")
	assert.EqualError(t, err, "missing version")

	t.Run("url safe versions only", func(t *testing.T) {
		assert.NoError(t, ValidateVersion("Foo123"))
		for _, reservedChar := range uriReservedChars {
			invalidVersion := fmt.Sprintf("foo%c", reservedChar)
			assert.NotNil(t, ValidateVersion(invalidVersion))
		}
	})
}

func TestValidateListTaskRequest(t *testing.T) {
	request := admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Limit: 10,
	}
	assert.NoError(t, ValidateResourceListRequest(request))
}

func TestValidateListTaskRequest_MissingProject(t *testing.T) {
	request := admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Domain: "domain",
			Name:   "name",
		},
		Limit: 10,
	}
	assert.EqualError(t, ValidateResourceListRequest(request), "missing project")
}

func TestValidateListTaskRequest_MissingDomain(t *testing.T) {
	request := admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Name:    "name",
		},
		Limit: 10,
	}
	assert.EqualError(t, ValidateResourceListRequest(request), "missing domain")
}

func TestValidateListTaskRequest_MissingName(t *testing.T) {
	request := admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
		},
		Limit: 10,
	}
	assert.NoError(t, ValidateResourceListRequest(request))
}

func TestValidateListTaskRequest_MissingLimit(t *testing.T) {
	request := admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	}
	assert.EqualError(t, ValidateResourceListRequest(request), "invalid value for limit")
}

func TestValidateParameterMap(t *testing.T) {
	t.Run("valid field", func(t *testing.T) {
		exampleMap := core.ParameterMap{
			Parameters: map[string]*core.Parameter{
				"foo": {
					Var: &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
					},
					Behavior: &core.Parameter_Default{
						Default: coreutils.MustMakeLiteral("foo-value"),
					},
				},
			},
		}
		err := validateParameterMap(&exampleMap, "foo")
		assert.NoError(t, err)
	})
	t.Run("invalid because missing required and defaults", func(t *testing.T) {
		exampleMap := core.ParameterMap{
			Parameters: map[string]*core.Parameter{
				"foo": {
					Var: &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
					},
					Behavior: nil, // neither required or defaults
				},
			},
		}
		err := validateParameterMap(&exampleMap, "some text")
		assert.Error(t, err)
	})
	t.Run("valid with required true", func(t *testing.T) {
		exampleMap := core.ParameterMap{
			Parameters: map[string]*core.Parameter{
				"foo": {
					Var: &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
					},
					Behavior: &core.Parameter_Required{
						Required: true,
					},
				},
			},
		}
		err := validateParameterMap(&exampleMap, "some text")
		assert.NoError(t, err)
	})
	t.Run("invalid because not required and no default provided", func(t *testing.T) {
		exampleMap := core.ParameterMap{
			Parameters: map[string]*core.Parameter{
				"foo": {
					Var: &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
					},
					Behavior: &core.Parameter_Required{
						Required: false,
					},
				},
			},
		}
		err := validateParameterMap(&exampleMap, "some text")
		assert.Error(t, err)
	})
	t.Run("valid datetime field", func(t *testing.T) {
		exampleMap := core.ParameterMap{
			Parameters: map[string]*core.Parameter{
				"foo": {
					Var: &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_DATETIME}},
					},
					Behavior: &core.Parameter_Default{
						Default: coreutils.MustMakeLiteral(time.Now()),
					},
				},
			},
		}
		err := validateParameterMap(&exampleMap, "some text")
		assert.NoError(t, err)
	})
}

func TestValidateToken(t *testing.T) {
	offset, err := ValidateToken("")
	assert.Nil(t, err)
	assert.Equal(t, 0, offset)

	offset, err = ValidateToken("1")
	assert.Nil(t, err)
	assert.Equal(t, 1, offset)

	_, err = ValidateToken("foo")
	assert.NotNil(t, err)

	_, err = ValidateToken("-1")
	assert.NotNil(t, err)
}

func TestValidateActiveLaunchPlanRequest(t *testing.T) {
	err := ValidateActiveLaunchPlanRequest(
		admin.ActiveLaunchPlanRequest{
			Id: &admin.NamedEntityIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
		},
	)
	assert.Nil(t, err)

	err = ValidateActiveLaunchPlanRequest(
		admin.ActiveLaunchPlanRequest{
			Id: &admin.NamedEntityIdentifier{
				Domain: "d",
				Name:   "n",
			},
		},
	)
	assert.Error(t, err)

	err = ValidateActiveLaunchPlanRequest(
		admin.ActiveLaunchPlanRequest{
			Id: &admin.NamedEntityIdentifier{
				Project: "p",
				Name:    "n",
			},
		},
	)
	assert.Error(t, err)

	err = ValidateActiveLaunchPlanRequest(
		admin.ActiveLaunchPlanRequest{
			Id: &admin.NamedEntityIdentifier{
				Project: "p",
				Domain:  "d",
			},
		},
	)
	assert.Error(t, err)
}

func TestValidateActiveLaunchPlanListRequest(t *testing.T) {
	err := ValidateActiveLaunchPlanListRequest(
		admin.ActiveLaunchPlanListRequest{
			Project: "p",
			Domain:  "d",
			Limit:   2,
		},
	)
	assert.Nil(t, err)

	err = ValidateActiveLaunchPlanListRequest(
		admin.ActiveLaunchPlanListRequest{
			Domain: "d",
		},
	)
	assert.Error(t, err)

	err = ValidateActiveLaunchPlanListRequest(
		admin.ActiveLaunchPlanListRequest{
			Project: "p",
		},
	)
	assert.Error(t, err)

	err = ValidateActiveLaunchPlanListRequest(
		admin.ActiveLaunchPlanListRequest{
			Project: "p",
			Domain:  "d",
			Limit:   0,
		},
	)
	assert.Error(t, err)
}

func TestValidateOutputData(t *testing.T) {
	t.Run("no output data", func(t *testing.T) {
		assert.NoError(t, ValidateOutputData(nil, 100))
	})
	outputData := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": {
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Integer{
									Integer: 4,
								},
							},
						},
					},
				},
			},
		},
	}
	t.Run("output data within threshold", func(t *testing.T) {
		assert.NoError(t, ValidateOutputData(outputData, int64(10000000)))
	})
	t.Run("output data greater than threshold", func(t *testing.T) {
		err := ValidateOutputData(outputData, int64(1))
		assert.Equal(t, codes.ResourceExhausted, err.(errors.FlyteAdminError).Code())
	})
}

func TestValidateDatetime(t *testing.T) {
	t.Run("no datetime", func(t *testing.T) {
		assert.EqualError(t, ValidateDatetime(nil), "Found invalid nil datetime")
	})
	t.Run("datetime with valid format and value", func(t *testing.T) {
		assert.NoError(t, ValidateDatetime(&core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Datetime{Datetime: timestamppb.Now()},
						},
					},
				},
			},
		}))
	})
	t.Run("datetime with value below min", func(t *testing.T) {
		// given
		timestamp := timestamppb.Timestamp{Seconds: -62135596801, Nanos: 999999999} // = 0000-12-31T23:59:59.999999999Z
		expectedErrStr := "before 0001-01-01"

		// when
		result := ValidateDatetime(&core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Datetime{Datetime: &timestamp},
						},
					},
				},
			},
		})

		// then
		assert.NotNil(t, result)
		assert.Containsf(t, result.Error(), expectedErrStr, "Found unexpected error message")
	})
	t.Run("datetime with value equals Instant.MIN", func(t *testing.T) {
		// given
		timestamp := timestamppb.Timestamp{Seconds: -31557014167219200, Nanos: 0} // = -1000000000-01-01T00:00Z
		expectedErrStr := "timestamp (seconds:-31557014167219200) before 0001-01-01"

		// when
		result := ValidateDatetime(&core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Datetime{Datetime: &timestamp},
						},
					},
				},
			},
		})

		// then
		assert.NotNil(t, result)
		assert.Containsf(t, result.Error(), expectedErrStr, "Found unexpected error message")
	})
	t.Run("datetime with min valid value", func(t *testing.T) {
		timestamp := timestamppb.Timestamp{Seconds: -62135596800, Nanos: 0} // = 0001-01-01T00:00:00Z

		assert.NoError(t, ValidateDatetime(&core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Datetime{Datetime: &timestamp},
						},
					},
				},
			},
		}))
	})
	t.Run("datetime with max valid value", func(t *testing.T) {
		timestamp := timestamppb.Timestamp{Seconds: 253402300799, Nanos: 999999999} // = 9999-12-31T23:59:59.999999999Z

		assert.NoError(t, ValidateDatetime(&core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Datetime{Datetime: &timestamp},
						},
					},
				},
			},
		}))
	})
	t.Run("datetime with value above max", func(t *testing.T) {
		// given
		timestamp := timestamppb.Timestamp{Seconds: 253402300800, Nanos: 0} // = 0000-12-31T23:59:59.999999999Z
		expectedErrStr := "timestamp (seconds:253402300800) after 9999-12-31"

		// when
		result := ValidateDatetime(&core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Datetime{Datetime: &timestamp},
						},
					},
				},
			},
		})

		// then
		assert.NotNil(t, result)
		assert.Containsf(t, result.Error(), expectedErrStr, "Found unexpected error message")
	})
}
