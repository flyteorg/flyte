// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	models "github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
)

// ArtifactRepo is an autogenerated mock type for the ArtifactRepo type
type ArtifactRepo struct {
	mock.Mock
}

type ArtifactRepo_Create struct {
	*mock.Call
}

func (_m ArtifactRepo_Create) Return(_a0 error) *ArtifactRepo_Create {
	return &ArtifactRepo_Create{Call: _m.Call.Return(_a0)}
}

func (_m *ArtifactRepo) OnCreate(ctx context.Context, in models.Artifact) *ArtifactRepo_Create {
	c_call := _m.On("Create", ctx, in)
	return &ArtifactRepo_Create{Call: c_call}
}

func (_m *ArtifactRepo) OnCreateMatch(matchers ...interface{}) *ArtifactRepo_Create {
	c_call := _m.On("Create", matchers...)
	return &ArtifactRepo_Create{Call: c_call}
}

// Create provides a mock function with given fields: ctx, in
func (_m *ArtifactRepo) Create(ctx context.Context, in models.Artifact) error {
	ret := _m.Called(ctx, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, models.Artifact) error); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type ArtifactRepo_Get struct {
	*mock.Call
}

func (_m ArtifactRepo_Get) Return(_a0 models.Artifact, _a1 error) *ArtifactRepo_Get {
	return &ArtifactRepo_Get{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ArtifactRepo) OnGet(ctx context.Context, in models.ArtifactKey) *ArtifactRepo_Get {
	c_call := _m.On("Get", ctx, in)
	return &ArtifactRepo_Get{Call: c_call}
}

func (_m *ArtifactRepo) OnGetMatch(matchers ...interface{}) *ArtifactRepo_Get {
	c_call := _m.On("Get", matchers...)
	return &ArtifactRepo_Get{Call: c_call}
}

// Get provides a mock function with given fields: ctx, in
func (_m *ArtifactRepo) Get(ctx context.Context, in models.ArtifactKey) (models.Artifact, error) {
	ret := _m.Called(ctx, in)

	var r0 models.Artifact
	if rf, ok := ret.Get(0).(func(context.Context, models.ArtifactKey) models.Artifact); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(models.Artifact)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.ArtifactKey) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ArtifactRepo_List struct {
	*mock.Call
}

func (_m ArtifactRepo_List) Return(_a0 []models.Artifact, _a1 error) *ArtifactRepo_List {
	return &ArtifactRepo_List{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ArtifactRepo) OnList(ctx context.Context, datasetKey models.DatasetKey, in models.ListModelsInput) *ArtifactRepo_List {
	c_call := _m.On("List", ctx, datasetKey, in)
	return &ArtifactRepo_List{Call: c_call}
}

func (_m *ArtifactRepo) OnListMatch(matchers ...interface{}) *ArtifactRepo_List {
	c_call := _m.On("List", matchers...)
	return &ArtifactRepo_List{Call: c_call}
}

// List provides a mock function with given fields: ctx, datasetKey, in
func (_m *ArtifactRepo) List(ctx context.Context, datasetKey models.DatasetKey, in models.ListModelsInput) ([]models.Artifact, error) {
	ret := _m.Called(ctx, datasetKey, in)

	var r0 []models.Artifact
	if rf, ok := ret.Get(0).(func(context.Context, models.DatasetKey, models.ListModelsInput) []models.Artifact); ok {
		r0 = rf(ctx, datasetKey, in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Artifact)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, models.DatasetKey, models.ListModelsInput) error); ok {
		r1 = rf(ctx, datasetKey, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ArtifactRepo_Update struct {
	*mock.Call
}

func (_m ArtifactRepo_Update) Return(_a0 error) *ArtifactRepo_Update {
	return &ArtifactRepo_Update{Call: _m.Call.Return(_a0)}
}

func (_m *ArtifactRepo) OnUpdate(ctx context.Context, artifact models.Artifact) *ArtifactRepo_Update {
	c_call := _m.On("Update", ctx, artifact)
	return &ArtifactRepo_Update{Call: c_call}
}

func (_m *ArtifactRepo) OnUpdateMatch(matchers ...interface{}) *ArtifactRepo_Update {
	c_call := _m.On("Update", matchers...)
	return &ArtifactRepo_Update{Call: c_call}
}

// Update provides a mock function with given fields: ctx, artifact
func (_m *ArtifactRepo) Update(ctx context.Context, artifact models.Artifact) error {
	ret := _m.Called(ctx, artifact)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, models.Artifact) error); ok {
		r0 = rf(ctx, artifact)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
