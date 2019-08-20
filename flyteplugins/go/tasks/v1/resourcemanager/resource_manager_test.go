package resourcemanager

import (
	"context"
	"reflect"
	"testing"
)

func TestGetOrCreateResourceManagerFor(t *testing.T) {
	type args struct {
		ctx          context.Context
		resourceName string
	}
	tests := []struct {
		name    string
		args    args
		want    ResourceManager
		wantErr bool
	}{
		{name: "Simple", args: args{ctx: context.TODO(), resourceName: "simple"}, want: NoopResourceManager{}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOrCreateResourceManagerFor(tt.args.ctx, tt.args.resourceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrCreateResourceManagerFor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOrCreateResourceManagerFor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNoopResourceManager_AllocateResource(t *testing.T) {
	type args struct {
		ctx             context.Context
		namespace       string
		allocationToken string
	}
	tests := []struct {
		name    string
		n       NoopResourceManager
		args    args
		want    AllocationStatus
		wantErr bool
	}{
		{name: "Simple", n: NoopResourceManager{}, args: args{ctx: context.TODO(), namespace: "namespace", allocationToken: "token"}, want: AllocationStatusGranted, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NoopResourceManager{}
			got, err := n.AllocateResource(tt.args.ctx, tt.args.namespace, tt.args.allocationToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("NoopResourceManager.AllocateResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NoopResourceManager.AllocateResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNoopResourceManager_ReleaseResource(t *testing.T) {
	type args struct {
		ctx             context.Context
		namespace       string
		allocationToken string
	}
	tests := []struct {
		name    string
		n       NoopResourceManager
		args    args
		wantErr bool
	}{
		{name: "Simple", n: NoopResourceManager{}, args: args{ctx: context.TODO(), namespace: "namespace", allocationToken: "token"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NoopResourceManager{}
			if err := n.ReleaseResource(tt.args.ctx, tt.args.namespace, tt.args.allocationToken); (err != nil) != tt.wantErr {
				t.Errorf("NoopResourceManager.ReleaseResource() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
