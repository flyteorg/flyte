package catalog

import (
	"context"
	"reflect"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue/mocks"
	"github.com/flyteorg/flytestdlib/bitarray"
	"github.com/stretchr/testify/mock"
)

var exampleInterface = &core.TypedInterface{
	Inputs: &core.VariableMap{
		Variables: map[string]*core.Variable{
			"a": {
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
	},
}
var input1 = &core.LiteralMap{
	Literals: map[string]*core.Literal{
		"a": {
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Integer{
								Integer: 1,
							},
						},
					},
				},
			},
		},
	},
}
var input2 = &core.LiteralMap{
	Literals: map[string]*core.Literal{
		"a": {
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Integer{
								Integer: 2,
							},
						},
					},
				},
			},
		},
	},
}

func TestAsyncClientImpl_Download(t *testing.T) {
	ctx := context.Background()

	q := &mocks.IndexedWorkQueue{}
	info := &mocks.WorkItemInfo{}
	info.OnItem().Return(NewReaderWorkItem(Key{}, &mocks2.OutputWriter{}))
	info.OnStatus().Return(workqueue.WorkStatusSucceeded)
	q.OnGetMatch(mock.Anything).Return(info, true, nil)
	q.OnQueueMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	ow := &mocks2.OutputWriter{}
	ow.OnGetOutputPrefixPath().Return("/prefix/")
	ow.OnGetOutputPath().Return("/prefix/outputs.pb")

	tests := []struct {
		name             string
		reader           workqueue.IndexedWorkQueue
		requests         []DownloadRequest
		wantOutputFuture DownloadFuture
		wantErr          bool
	}{
		{"DownloadQueued", q, []DownloadRequest{
			{
				Key:    Key{},
				Target: ow,
			},
		}, newDownloadFuture(ResponseStatusReady, nil, bitarray.NewBitSet(1), 1, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := AsyncClientImpl{
				Reader: tt.reader,
			}
			gotOutputFuture, err := c.Download(ctx, tt.requests...)
			if (err != nil) != tt.wantErr {
				t.Errorf("AsyncClientImpl.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOutputFuture, tt.wantOutputFuture) {
				t.Errorf("AsyncClientImpl.Download() = %v, want %v", gotOutputFuture, tt.wantOutputFuture)
			}
		})
	}
}

func TestAsyncClientImpl_Upload(t *testing.T) {
	ctx := context.Background()

	inputHash1 := "{UNSPECIFIED     {} [] 0}:-0-DNhkpTTPC5YDtRGb4yT-PFxgMSgHzHrKAQKgQGEfGRY"
	inputHash2 := "{UNSPECIFIED     {} [] 0}:-1-26M4dwarvBVJqJSUC4JC1GtRYgVBIAmQfsFSdLVMlAc"

	q := &mocks.IndexedWorkQueue{}
	info := &mocks.WorkItemInfo{}
	info.OnItem().Return(NewReaderWorkItem(Key{}, &mocks2.OutputWriter{}))
	info.OnStatus().Return(workqueue.WorkStatusSucceeded)
	q.OnGet(inputHash1).Return(info, true, nil)
	q.OnGet(inputHash2).Return(info, true, nil)
	q.OnQueueMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	inputReader1 := &mocks2.InputReader{}
	inputReader1.OnGetMatch(mock.Anything).Return(input1, nil)
	inputReader2 := &mocks2.InputReader{}
	inputReader2.OnGetMatch(mock.Anything).Return(input2, nil)

	tests := []struct {
		name          string
		requests      []UploadRequest
		wantPutFuture UploadFuture
		wantErr       bool
	}{
		{
			"UploadSucceeded",
			// The second request has the same Key.Identifier and Key.Cache version but a different
			// Key.InputReader. This should lead to a different WorkItemID in the queue.
			// See https://github.com/flyteorg/flyte/issues/3787 for more details
			[]UploadRequest{
				{
					Key: Key{
						TypedInterface: *exampleInterface,
						InputReader:    inputReader1,
					},
				},
				{
					Key: Key{
						TypedInterface: *exampleInterface,
						InputReader:    inputReader2,
					},
				},
			},
			newUploadFuture(ResponseStatusReady, nil),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := AsyncClientImpl{
				Writer: q,
			}
			gotPutFuture, err := c.Upload(ctx, tt.requests...)
			if (err != nil) != tt.wantErr {
				t.Errorf("AsyncClientImpl.Sidecar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPutFuture, tt.wantPutFuture) {
				t.Errorf("AsyncClientImpl.Sidecar() = %v, want %v", gotPutFuture, tt.wantPutFuture)
			}
			expectedWorkItemIDs := []string{inputHash1, inputHash2}
			gottenWorkItemIDs := make([]string, 0)
			for _, mockCall := range q.Calls {
				if mockCall.Method == "Get" {
					gottenWorkItemIDs = append(gottenWorkItemIDs, mockCall.Arguments[0].(string))
				}
			}
			if !reflect.DeepEqual(gottenWorkItemIDs, expectedWorkItemIDs) {
				t.Errorf("Retrieved workitem IDs = %v, want %v", gottenWorkItemIDs, expectedWorkItemIDs)
			}
		})
	}
}

func TestAsyncClientImpl_Start(t *testing.T) {
	type fields struct {
		Reader workqueue.IndexedWorkQueue
		Writer workqueue.IndexedWorkQueue
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := AsyncClientImpl{
				Reader: tt.fields.Reader,
				Writer: tt.fields.Writer,
			}
			if err := c.Start(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("AsyncClientImpl.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
