// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	azsecrets "github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"

	mock "github.com/stretchr/testify/mock"
)

// AzureKeyVaultClient is an autogenerated mock type for the AzureKeyVaultClient type
type AzureKeyVaultClient struct {
	mock.Mock
}

type AzureKeyVaultClient_GetSecret struct {
	*mock.Call
}

func (_m AzureKeyVaultClient_GetSecret) Return(_a0 azsecrets.GetSecretResponse, _a1 error) *AzureKeyVaultClient_GetSecret {
	return &AzureKeyVaultClient_GetSecret{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *AzureKeyVaultClient) OnGetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) *AzureKeyVaultClient_GetSecret {
	c_call := _m.On("GetSecret", ctx, name, version, options)
	return &AzureKeyVaultClient_GetSecret{Call: c_call}
}

func (_m *AzureKeyVaultClient) OnGetSecretMatch(matchers ...interface{}) *AzureKeyVaultClient_GetSecret {
	c_call := _m.On("GetSecret", matchers...)
	return &AzureKeyVaultClient_GetSecret{Call: c_call}
}

// GetSecret provides a mock function with given fields: ctx, name, version, options
func (_m *AzureKeyVaultClient) GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	ret := _m.Called(ctx, name, version, options)

	var r0 azsecrets.GetSecretResponse
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *azsecrets.GetSecretOptions) azsecrets.GetSecretResponse); ok {
		r0 = rf(ctx, name, version, options)
	} else {
		r0 = ret.Get(0).(azsecrets.GetSecretResponse)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, *azsecrets.GetSecretOptions) error); ok {
		r1 = rf(ctx, name, version, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
