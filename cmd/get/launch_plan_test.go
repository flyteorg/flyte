package get

import (
	"fmt"
	"os"
	"testing"

	"github.com/flyteorg/flytectl/pkg/filters"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/launchplan"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flytectl/pkg/ext/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	resourceListRequest    *admin.ResourceListRequest
	resourceGetRequest     *admin.ResourceListRequest
	objectGetRequest       *admin.ObjectGetRequest
	namedIDRequest         *admin.NamedEntityIdentifierListRequest
	launchPlanListResponse *admin.LaunchPlanList
	argsLp                 []string
	namedIdentifierList    *admin.NamedEntityIdentifierList
	launchPlan2            *admin.LaunchPlan
)

func getLaunchPlanSetup() {
	ctx = u.Ctx
	mockClient = u.MockClient
	// TODO: migrate to new command context from testutils
	cmdCtx = cmdCore.NewCommandContext(mockClient, u.MockOutStream)
	argsLp = []string{"launchplan1"}
	parameterMap := map[string]*core.Parameter{
		"numbers": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
				Description: "short desc",
			},
		},
		"numbers_count": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
				Description: "long description will be truncated in table",
			},
		},
		"run_local_at_count": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
				Description: "run_local_at_count",
			},
			Behavior: &core.Parameter_Default{
				Default: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 10,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	launchPlan1 := &admin.LaunchPlan{
		Id: &core.Identifier{
			Name:    "launchplan1",
			Version: "v1",
		},
		Spec: &admin.LaunchPlanSpec{
			DefaultInputs: &core.ParameterMap{
				Parameters: parameterMap,
			},
		},
		Closure: &admin.LaunchPlanClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
			ExpectedInputs: &core.ParameterMap{
				Parameters: parameterMap,
			},
		},
	}
	launchPlan2 = &admin.LaunchPlan{
		Id: &core.Identifier{
			Name:    "launchplan1",
			Version: "v2",
		},
		Spec: &admin.LaunchPlanSpec{
			DefaultInputs: &core.ParameterMap{
				Parameters: parameterMap,
			},
		},
		Closure: &admin.LaunchPlanClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			ExpectedInputs: &core.ParameterMap{
				Parameters: parameterMap,
			},
		},
	}

	launchPlans := []*admin.LaunchPlan{launchPlan2, launchPlan1}

	resourceListRequest = &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
	}

	resourceGetRequest = &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    argsLp[0],
		},
	}

	launchPlanListResponse = &admin.LaunchPlanList{
		LaunchPlans: launchPlans,
	}

	objectGetRequest = &admin.ObjectGetRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Project:      projectValue,
			Domain:       domainValue,
			Name:         argsLp[0],
			Version:      "v2",
		},
	}

	namedIDRequest = &admin.NamedEntityIdentifierListRequest{
		Project: projectValue,
		Domain:  domainValue,
	}

	var entities []*admin.NamedEntityIdentifier
	id1 := &admin.NamedEntityIdentifier{
		Project: projectValue,
		Domain:  domainValue,
		Name:    "launchplan1",
	}
	id2 := &admin.NamedEntityIdentifier{
		Project: projectValue,
		Domain:  domainValue,
		Name:    "launchplan2",
	}
	entities = append(entities, id1, id2)
	namedIdentifierList = &admin.NamedEntityIdentifierList{
		Entities: entities,
	}

	launchplan.DefaultConfig.Latest = false
	launchplan.DefaultConfig.Version = ""
	launchplan.DefaultConfig.ExecFile = ""
	launchplan.DefaultConfig.Filter = filters.Filters{}
}

func TestGetLaunchPlanFuncWithError(t *testing.T) {
	t.Run("failure fetch latest", func(t *testing.T) {
		setup()
		getLaunchPlanSetup()
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		launchplan.DefaultConfig.Latest = true
		launchplan.DefaultConfig.Filter = filters.Filters{}
		mockFetcher.OnFetchLPLatestVersionMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error fetching latest version"))
		_, err = FetchLPForName(ctx, mockFetcher, "lpName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching version ", func(t *testing.T) {
		setup()
		getLaunchPlanSetup()
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		launchplan.DefaultConfig.Version = "v1"
		launchplan.DefaultConfig.Filter = filters.Filters{}
		mockFetcher.OnFetchLPVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("error fetching version"))
		_, err = FetchLPForName(ctx, mockFetcher, "lpName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching all version ", func(t *testing.T) {
		setup()
		getLaunchPlanSetup()
		launchplan.DefaultConfig.Filter = filters.Filters{}
		launchplan.DefaultConfig.Filter = filters.Filters{}
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		mockFetcher.OnFetchAllVerOfLPMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error fetching all version"))
		_, err = FetchLPForName(ctx, mockFetcher, "lpName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching ", func(t *testing.T) {
		setup()
		getLaunchPlanSetup()
		mockClient.OnListLaunchPlansMatch(ctx, resourceGetRequest).Return(nil, fmt.Errorf("error fetching all version"))
		mockClient.OnGetLaunchPlanMatch(ctx, objectGetRequest).Return(nil, fmt.Errorf("error fetching lanuch plan"))
		mockClient.OnListLaunchPlanIdsMatch(ctx, namedIDRequest).Return(nil, fmt.Errorf("error listing lanuch plan ids"))
		err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching list", func(t *testing.T) {
		setup()
		getLaunchPlanSetup()
		argsLp = []string{}
		mockClient.OnListLaunchPlansMatch(ctx, resourceListRequest).Return(nil, fmt.Errorf("error fetching all version"))
		err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
		assert.NotNil(t, err)
	})
}

func TestGetLaunchPlanFunc(t *testing.T) {
	setup()
	getLaunchPlanSetup()
	mockClient.OnListLaunchPlansMatch(ctx, resourceGetRequest).Return(launchPlanListResponse, nil)
	mockClient.OnGetLaunchPlanMatch(ctx, objectGetRequest).Return(launchPlan2, nil)
	mockClient.OnListLaunchPlanIdsMatch(ctx, namedIDRequest).Return(namedIdentifierList, nil)
	err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "ListLaunchPlans", ctx, resourceGetRequest)
	tearDownAndVerify(t, `[
	{
		"id": {
			"name": "launchplan1",
			"version": "v2"
		},
		"spec": {
			"defaultInputs": {
				"parameters": {
					"numbers": {
						"var": {
							"type": {
								"collectionType": {
									"simple": "INTEGER"
								}
							},
							"description": "short desc"
						}
					},
					"numbers_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							},
							"description": "long description will be truncated in table"
						}
					},
					"run_local_at_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							},
							"description": "run_local_at_count"
						},
						"default": {
							"scalar": {
								"primitive": {
									"integer": "10"
								}
							}
						}
					}
				}
			}
		},
		"closure": {
			"expectedInputs": {
				"parameters": {
					"numbers": {
						"var": {
							"type": {
								"collectionType": {
									"simple": "INTEGER"
								}
							},
							"description": "short desc"
						}
					},
					"numbers_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							},
							"description": "long description will be truncated in table"
						}
					},
					"run_local_at_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							},
							"description": "run_local_at_count"
						},
						"default": {
							"scalar": {
								"primitive": {
									"integer": "10"
								}
							}
						}
					}
				}
			},
			"createdAt": "1970-01-01T00:00:01Z"
		}
	},
	{
		"id": {
			"name": "launchplan1",
			"version": "v1"
		},
		"spec": {
			"defaultInputs": {
				"parameters": {
					"numbers": {
						"var": {
							"type": {
								"collectionType": {
									"simple": "INTEGER"
								}
							},
							"description": "short desc"
						}
					},
					"numbers_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							},
							"description": "long description will be truncated in table"
						}
					},
					"run_local_at_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							},
							"description": "run_local_at_count"
						},
						"default": {
							"scalar": {
								"primitive": {
									"integer": "10"
								}
							}
						}
					}
				}
			}
		},
		"closure": {
			"expectedInputs": {
				"parameters": {
					"numbers": {
						"var": {
							"type": {
								"collectionType": {
									"simple": "INTEGER"
								}
							},
							"description": "short desc"
						}
					},
					"numbers_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							},
							"description": "long description will be truncated in table"
						}
					},
					"run_local_at_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							},
							"description": "run_local_at_count"
						},
						"default": {
							"scalar": {
								"primitive": {
									"integer": "10"
								}
							}
						}
					}
				}
			},
			"createdAt": "1970-01-01T00:00:00Z"
		}
	}
]`)
}

func TestGetLaunchPlanFuncLatest(t *testing.T) {
	setup()
	getLaunchPlanSetup()
	launchplan.DefaultConfig.Latest = true
	launchplan.DefaultConfig.Filter = filters.Filters{}
	mockClient.OnListLaunchPlansMatch(ctx, resourceGetRequest).Return(launchPlanListResponse, nil)
	mockClient.OnGetLaunchPlanMatch(ctx, objectGetRequest).Return(launchPlan2, nil)
	err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "ListLaunchPlans", ctx, resourceGetRequest)
	tearDownAndVerify(t, `{
	"id": {
		"name": "launchplan1",
		"version": "v2"
	},
	"spec": {
		"defaultInputs": {
			"parameters": {
				"numbers": {
					"var": {
						"type": {
							"collectionType": {
								"simple": "INTEGER"
							}
						},
						"description": "short desc"
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "long description will be truncated in table"
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "run_local_at_count"
					},
					"default": {
						"scalar": {
							"primitive": {
								"integer": "10"
							}
						}
					}
				}
			}
		}
	},
	"closure": {
		"expectedInputs": {
			"parameters": {
				"numbers": {
					"var": {
						"type": {
							"collectionType": {
								"simple": "INTEGER"
							}
						},
						"description": "short desc"
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "long description will be truncated in table"
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "run_local_at_count"
					},
					"default": {
						"scalar": {
							"primitive": {
								"integer": "10"
							}
						}
					}
				}
			}
		},
		"createdAt": "1970-01-01T00:00:01Z"
	}
}`)
}

func TestGetLaunchPlanWithVersion(t *testing.T) {
	setup()
	getLaunchPlanSetup()
	launchplan.DefaultConfig.Version = "v2"
	mockClient.OnListLaunchPlansMatch(ctx, resourceListRequest).Return(launchPlanListResponse, nil)
	mockClient.OnGetLaunchPlanMatch(ctx, objectGetRequest).Return(launchPlan2, nil)
	mockClient.OnListLaunchPlanIdsMatch(ctx, namedIDRequest).Return(namedIdentifierList, nil)
	err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "GetLaunchPlan", ctx, objectGetRequest)
	tearDownAndVerify(t, `{
	"id": {
		"name": "launchplan1",
		"version": "v2"
	},
	"spec": {
		"defaultInputs": {
			"parameters": {
				"numbers": {
					"var": {
						"type": {
							"collectionType": {
								"simple": "INTEGER"
							}
						},
						"description": "short desc"
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "long description will be truncated in table"
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "run_local_at_count"
					},
					"default": {
						"scalar": {
							"primitive": {
								"integer": "10"
							}
						}
					}
				}
			}
		}
	},
	"closure": {
		"expectedInputs": {
			"parameters": {
				"numbers": {
					"var": {
						"type": {
							"collectionType": {
								"simple": "INTEGER"
							}
						},
						"description": "short desc"
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "long description will be truncated in table"
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "run_local_at_count"
					},
					"default": {
						"scalar": {
							"primitive": {
								"integer": "10"
							}
						}
					}
				}
			}
		},
		"createdAt": "1970-01-01T00:00:01Z"
	}
}`)
}

func TestGetLaunchPlans(t *testing.T) {
	setup()
	getLaunchPlanSetup()
	mockClient.OnListLaunchPlansMatch(ctx, resourceListRequest).Return(launchPlanListResponse, nil)
	mockClient.OnGetLaunchPlanMatch(ctx, objectGetRequest).Return(launchPlan2, nil)
	argsLp = []string{}
	err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
	assert.Nil(t, err)
	tearDownAndVerify(t, `[{"id": {"name": "launchplan1","version": "v2"},"spec": {"defaultInputs": {"parameters": {"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:01Z"}},{"id": {"name": "launchplan1","version": "v1"},"spec": {"defaultInputs": {"parameters": {"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:00Z"}}]`)
}

func TestGetLaunchPlansWithExecFile(t *testing.T) {
	setup()
	getLaunchPlanSetup()
	mockClient.OnListLaunchPlansMatch(ctx, resourceListRequest).Return(launchPlanListResponse, nil)
	mockClient.OnGetLaunchPlanMatch(ctx, objectGetRequest).Return(launchPlan2, nil)
	mockClient.OnListLaunchPlanIdsMatch(ctx, namedIDRequest).Return(namedIdentifierList, nil)
	launchplan.DefaultConfig.Version = "v2"
	launchplan.DefaultConfig.ExecFile = testDataFolder + "exec_file"
	err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
	os.Remove(launchplan.DefaultConfig.ExecFile)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "GetLaunchPlan", ctx, objectGetRequest)
	tearDownAndVerify(t, `{
	"id": {
		"name": "launchplan1",
		"version": "v2"
	},
	"spec": {
		"defaultInputs": {
			"parameters": {
				"numbers": {
					"var": {
						"type": {
							"collectionType": {
								"simple": "INTEGER"
							}
						},
						"description": "short desc"
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "long description will be truncated in table"
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "run_local_at_count"
					},
					"default": {
						"scalar": {
							"primitive": {
								"integer": "10"
							}
						}
					}
				}
			}
		}
	},
	"closure": {
		"expectedInputs": {
			"parameters": {
				"numbers": {
					"var": {
						"type": {
							"collectionType": {
								"simple": "INTEGER"
							}
						},
						"description": "short desc"
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "long description will be truncated in table"
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						},
						"description": "run_local_at_count"
					},
					"default": {
						"scalar": {
							"primitive": {
								"integer": "10"
							}
						}
					}
				}
			}
		},
		"createdAt": "1970-01-01T00:00:01Z"
	}
}`)
}

func TestGetLaunchPlanTableFunc(t *testing.T) {
	setup()
	getLaunchPlanSetup()
	mockClient.OnListLaunchPlansMatch(ctx, resourceGetRequest).Return(launchPlanListResponse, nil)
	mockClient.OnGetLaunchPlanMatch(ctx, objectGetRequest).Return(launchPlan2, nil)
	mockClient.OnListLaunchPlanIdsMatch(ctx, namedIDRequest).Return(namedIdentifierList, nil)
	config.GetConfig().Output = "table"
	err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "ListLaunchPlans", ctx, resourceGetRequest)
	tearDownAndVerify(t, `
--------- ------------- ------ ------- ---------- --------------------------- --------- 
| VERSION | NAME        | TYPE | STATE | SCHEDULE | INPUTS                    | OUTPUTS | 
--------- ------------- ------ ------- ---------- --------------------------- --------- 
| v2      | launchplan1 |      |       |          | numbers: short desc       |         |
|         |             |      |       |          | numbers_count: long de... |         |
|         |             |      |       |          | run_local_at_count        |         | 
--------- ------------- ------ ------- ---------- --------------------------- --------- 
| v1      | launchplan1 |      |       |          | numbers: short desc       |         |
|         |             |      |       |          | numbers_count: long de... |         |
|         |             |      |       |          | run_local_at_count        |         | 
--------- ------------- ------ ------- ---------- --------------------------- --------- 
2 rows`)
}
