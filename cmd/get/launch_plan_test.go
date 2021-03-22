package get

import (
	"os"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	resourceListRequest    *admin.ResourceListRequest
	objectGetRequest       *admin.ObjectGetRequest
	namedIDRequest         *admin.NamedEntityIdentifierListRequest
	launchPlanListResponse *admin.LaunchPlanList
	argsLp                 []string
)

func getLaunchPlanSetup() {
	ctx = testutils.Ctx
	cmdCtx = testutils.CmdCtx
	mockClient = testutils.MockClient
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
			},
		},
		"numbers_count": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
		"run_local_at_count": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
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
	launchPlan2 := &admin.LaunchPlan{
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
			Name:    argsLp[0],
		},
		SortBy: &admin.Sort{
			Key:       "created_at",
			Direction: admin.Sort_DESCENDING,
		},
		Limit: 100,
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
		SortBy: &admin.Sort{
			Key:       "name",
			Direction: admin.Sort_ASCENDING,
		},
		Limit: 100,
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
	namedIdentifierList := &admin.NamedEntityIdentifierList{
		Entities: entities,
	}

	mockClient.OnListLaunchPlansMatch(ctx, resourceListRequest).Return(launchPlanListResponse, nil)
	mockClient.OnGetLaunchPlanMatch(ctx, objectGetRequest).Return(launchPlan2, nil)
	mockClient.OnListLaunchPlanIdsMatch(ctx, namedIDRequest).Return(namedIdentifierList, nil)

	launchPlanConfig.Latest = false
	launchPlanConfig.Version = ""
	launchPlanConfig.ExecFile = ""
}

func TestGetLaunchPlanFunc(t *testing.T) {
	setup()
	getLaunchPlanSetup()
	err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "ListLaunchPlans", ctx, resourceListRequest)
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
							}
						}
					},
					"numbers_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							}
						}
					},
					"run_local_at_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							}
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
							}
						}
					},
					"numbers_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							}
						}
					},
					"run_local_at_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							}
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
							}
						}
					},
					"numbers_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							}
						}
					},
					"run_local_at_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							}
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
							}
						}
					},
					"numbers_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							}
						}
					},
					"run_local_at_count": {
						"var": {
							"type": {
								"simple": "INTEGER"
							}
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
	launchPlanConfig.Latest = true
	err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "ListLaunchPlans", ctx, resourceListRequest)
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
						}
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
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
						}
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
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
	launchPlanConfig.Version = "v2"
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
						}
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
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
						}
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
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
	argsLp = []string{}
	err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "ListLaunchPlanIds", ctx, namedIDRequest)
	tearDownAndVerify(t, `[
	{
		"project": "dummyProject",
		"domain": "dummyDomain",
		"name": "launchplan1"
	},
	{
		"project": "dummyProject",
		"domain": "dummyDomain",
		"name": "launchplan2"
	}
]`)
}

func TestGetLaunchPlansWithExecFile(t *testing.T) {
	setup()
	getLaunchPlanSetup()
	launchPlanConfig.Version = "v2"
	launchPlanConfig.ExecFile = testDataFolder + "exec_file"
	err = getLaunchPlanFunc(ctx, argsLp, cmdCtx)
	os.Remove(launchPlanConfig.ExecFile)
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
						}
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
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
						}
					}
				},
				"numbers_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
					}
				},
				"run_local_at_count": {
					"var": {
						"type": {
							"simple": "INTEGER"
						}
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
