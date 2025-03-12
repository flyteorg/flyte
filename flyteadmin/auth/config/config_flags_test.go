// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

var dereferencableKindsConfig = map[reflect.Kind]struct{}{
	reflect.Array: {}, reflect.Chan: {}, reflect.Map: {}, reflect.Ptr: {}, reflect.Slice: {},
}

// Checks if t is a kind that can be dereferenced to get its underlying type.
func canGetElementConfig(t reflect.Kind) bool {
	_, exists := dereferencableKindsConfig[t]
	return exists
}

// This decoder hook tests types for json unmarshaling capability. If implemented, it uses json unmarshal to build the
// object. Otherwise, it'll just pass on the original data.
func jsonUnmarshalerHookConfig(_, to reflect.Type, data interface{}) (interface{}, error) {
	unmarshalerType := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	if to.Implements(unmarshalerType) || reflect.PtrTo(to).Implements(unmarshalerType) ||
		(canGetElementConfig(to.Kind()) && to.Elem().Implements(unmarshalerType)) {

		raw, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("Failed to marshal Data: %v. Error: %v. Skipping jsonUnmarshalHook", data, err)
			return data, nil
		}

		res := reflect.New(to).Interface()
		err = json.Unmarshal(raw, &res)
		if err != nil {
			fmt.Printf("Failed to umarshal Data: %v. Error: %v. Skipping jsonUnmarshalHook", data, err)
			return data, nil
		}

		return res, nil
	}

	return data, nil
}

func decode_Config(input, result interface{}) error {
	config := &mapstructure.DecoderConfig{
		TagName:          "json",
		WeaklyTypedInput: true,
		Result:           result,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			jsonUnmarshalerHookConfig,
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func join_Config(arr interface{}, sep string) string {
	listValue := reflect.ValueOf(arr)
	strs := make([]string, 0, listValue.Len())
	for i := 0; i < listValue.Len(); i++ {
		strs = append(strs, fmt.Sprintf("%v", listValue.Index(i)))
	}

	return strings.Join(strs, sep)
}

func testDecodeJson_Config(t *testing.T, val, result interface{}) {
	assert.NoError(t, decode_Config(val, result))
}

func testDecodeRaw_Config(t *testing.T, vStringSlice, result interface{}) {
	assert.NoError(t, decode_Config(vStringSlice, result))
}

func TestConfig_GetPFlagSet(t *testing.T) {
	val := Config{}
	cmdFlags := val.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())
}

func TestConfig_SetFlags(t *testing.T) {
	actual := Config{}
	cmdFlags := actual.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())

	t.Run("Test_httpAuthorizationHeader", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("httpAuthorizationHeader", testValue)
			if vString, err := cmdFlags.GetString("httpAuthorizationHeader"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.HTTPAuthorizationHeader)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_grpcAuthorizationHeader", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("grpcAuthorizationHeader", testValue)
			if vString, err := cmdFlags.GetString("grpcAuthorizationHeader"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.GrpcAuthorizationHeader)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_disableForHttp", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("disableForHttp", testValue)
			if vBool, err := cmdFlags.GetBool("disableForHttp"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.DisableForHTTP)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_disableForGrpc", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("disableForGrpc", testValue)
			if vBool, err := cmdFlags.GetBool("disableForGrpc"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.DisableForGrpc)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_httpProxyURL", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.HTTPProxyURL.String()

			cmdFlags.Set("httpProxyURL", testValue)
			if vString, err := cmdFlags.GetString("httpProxyURL"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.HTTPProxyURL)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.redirectUrl", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.UserAuth.RedirectURL.String()

			cmdFlags.Set("userAuth.redirectUrl", testValue)
			if vString, err := cmdFlags.GetString("userAuth.redirectUrl"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.RedirectURL)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.openId.clientId", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("userAuth.openId.clientId", testValue)
			if vString, err := cmdFlags.GetString("userAuth.openId.clientId"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.OpenID.ClientID)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.openId.clientSecretName", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("userAuth.openId.clientSecretName", testValue)
			if vString, err := cmdFlags.GetString("userAuth.openId.clientSecretName"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.OpenID.ClientSecretName)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.openId.clientSecretFile", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("userAuth.openId.clientSecretFile", testValue)
			if vString, err := cmdFlags.GetString("userAuth.openId.clientSecretFile"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.OpenID.DeprecatedClientSecretFile)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.openId.baseUrl", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.UserAuth.OpenID.BaseURL.String()

			cmdFlags.Set("userAuth.openId.baseUrl", testValue)
			if vString, err := cmdFlags.GetString("userAuth.openId.baseUrl"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.OpenID.BaseURL)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.openId.scopes", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := join_Config(DefaultConfig.UserAuth.OpenID.Scopes, ",")

			cmdFlags.Set("userAuth.openId.scopes", testValue)
			if vStringSlice, err := cmdFlags.GetStringSlice("userAuth.openId.scopes"); err == nil {
				testDecodeRaw_Config(t, join_Config(vStringSlice, ","), &actual.UserAuth.OpenID.Scopes)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.httpProxyURL", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.UserAuth.HTTPProxyURL.String()

			cmdFlags.Set("userAuth.httpProxyURL", testValue)
			if vString, err := cmdFlags.GetString("userAuth.httpProxyURL"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.HTTPProxyURL)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.cookieHashKeySecretName", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("userAuth.cookieHashKeySecretName", testValue)
			if vString, err := cmdFlags.GetString("userAuth.cookieHashKeySecretName"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.CookieHashKeySecretName)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.cookieBlockKeySecretName", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("userAuth.cookieBlockKeySecretName", testValue)
			if vString, err := cmdFlags.GetString("userAuth.cookieBlockKeySecretName"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.CookieBlockKeySecretName)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.cookieSetting.sameSitePolicy", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("userAuth.cookieSetting.sameSitePolicy", testValue)
			if vString, err := cmdFlags.GetString("userAuth.cookieSetting.sameSitePolicy"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.CookieSetting.SameSitePolicy)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.cookieSetting.domain", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("userAuth.cookieSetting.domain", testValue)
			if vString, err := cmdFlags.GetString("userAuth.cookieSetting.domain"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.CookieSetting.Domain)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_userAuth.idpQueryParameter", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("userAuth.idpQueryParameter", testValue)
			if vString, err := cmdFlags.GetString("userAuth.idpQueryParameter"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.UserAuth.IDPQueryParameter)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.selfAuthServer.issuer", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("appAuth.selfAuthServer.issuer", testValue)
			if vString, err := cmdFlags.GetString("appAuth.selfAuthServer.issuer"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.SelfAuthServer.Issuer)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.selfAuthServer.accessTokenLifespan", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.AppAuth.SelfAuthServer.AccessTokenLifespan.String()

			cmdFlags.Set("appAuth.selfAuthServer.accessTokenLifespan", testValue)
			if vString, err := cmdFlags.GetString("appAuth.selfAuthServer.accessTokenLifespan"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.SelfAuthServer.AccessTokenLifespan)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.selfAuthServer.refreshTokenLifespan", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.AppAuth.SelfAuthServer.RefreshTokenLifespan.String()

			cmdFlags.Set("appAuth.selfAuthServer.refreshTokenLifespan", testValue)
			if vString, err := cmdFlags.GetString("appAuth.selfAuthServer.refreshTokenLifespan"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.SelfAuthServer.RefreshTokenLifespan)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.selfAuthServer.authorizationCodeLifespan", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.AppAuth.SelfAuthServer.AuthorizationCodeLifespan.String()

			cmdFlags.Set("appAuth.selfAuthServer.authorizationCodeLifespan", testValue)
			if vString, err := cmdFlags.GetString("appAuth.selfAuthServer.authorizationCodeLifespan"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.SelfAuthServer.AuthorizationCodeLifespan)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.selfAuthServer.claimSymmetricEncryptionKeySecretName", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("appAuth.selfAuthServer.claimSymmetricEncryptionKeySecretName", testValue)
			if vString, err := cmdFlags.GetString("appAuth.selfAuthServer.claimSymmetricEncryptionKeySecretName"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.SelfAuthServer.ClaimSymmetricEncryptionKeySecretName)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.selfAuthServer.tokenSigningRSAKeySecretName", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("appAuth.selfAuthServer.tokenSigningRSAKeySecretName", testValue)
			if vString, err := cmdFlags.GetString("appAuth.selfAuthServer.tokenSigningRSAKeySecretName"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.SelfAuthServer.TokenSigningRSAKeySecretName)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.selfAuthServer.oldTokenSigningRSAKeySecretName", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("appAuth.selfAuthServer.oldTokenSigningRSAKeySecretName", testValue)
			if vString, err := cmdFlags.GetString("appAuth.selfAuthServer.oldTokenSigningRSAKeySecretName"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.SelfAuthServer.OldTokenSigningRSAKeySecretName)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.externalAuthServer.baseUrl", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.AppAuth.ExternalAuthServer.BaseURL.String()

			cmdFlags.Set("appAuth.externalAuthServer.baseUrl", testValue)
			if vString, err := cmdFlags.GetString("appAuth.externalAuthServer.baseUrl"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.ExternalAuthServer.BaseURL)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.externalAuthServer.allowedAudience", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := join_Config(DefaultConfig.AppAuth.ExternalAuthServer.AllowedAudience, ",")

			cmdFlags.Set("appAuth.externalAuthServer.allowedAudience", testValue)
			if vStringSlice, err := cmdFlags.GetStringSlice("appAuth.externalAuthServer.allowedAudience"); err == nil {
				testDecodeRaw_Config(t, join_Config(vStringSlice, ","), &actual.AppAuth.ExternalAuthServer.AllowedAudience)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.externalAuthServer.metadataUrl", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.AppAuth.ExternalAuthServer.MetadataEndpointURL.String()

			cmdFlags.Set("appAuth.externalAuthServer.metadataUrl", testValue)
			if vString, err := cmdFlags.GetString("appAuth.externalAuthServer.metadataUrl"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.ExternalAuthServer.MetadataEndpointURL)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.externalAuthServer.httpProxyURL", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.AppAuth.ExternalAuthServer.HTTPProxyURL.String()

			cmdFlags.Set("appAuth.externalAuthServer.httpProxyURL", testValue)
			if vString, err := cmdFlags.GetString("appAuth.externalAuthServer.httpProxyURL"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.ExternalAuthServer.HTTPProxyURL)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.externalAuthServer.retryAttempts", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("appAuth.externalAuthServer.retryAttempts", testValue)
			if vInt, err := cmdFlags.GetInt("appAuth.externalAuthServer.retryAttempts"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vInt), &actual.AppAuth.ExternalAuthServer.RetryAttempts)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.externalAuthServer.retryDelay", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := DefaultConfig.AppAuth.ExternalAuthServer.RetryDelay.String()

			cmdFlags.Set("appAuth.externalAuthServer.retryDelay", testValue)
			if vString, err := cmdFlags.GetString("appAuth.externalAuthServer.retryDelay"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.ExternalAuthServer.RetryDelay)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.thirdPartyConfig.flyteClient.clientId", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("appAuth.thirdPartyConfig.flyteClient.clientId", testValue)
			if vString, err := cmdFlags.GetString("appAuth.thirdPartyConfig.flyteClient.clientId"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.ThirdParty.FlyteClientConfig.ClientID)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.thirdPartyConfig.flyteClient.redirectUri", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("appAuth.thirdPartyConfig.flyteClient.redirectUri", testValue)
			if vString, err := cmdFlags.GetString("appAuth.thirdPartyConfig.flyteClient.redirectUri"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.ThirdParty.FlyteClientConfig.RedirectURI)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.thirdPartyConfig.flyteClient.scopes", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := join_Config(DefaultConfig.AppAuth.ThirdParty.FlyteClientConfig.Scopes, ",")

			cmdFlags.Set("appAuth.thirdPartyConfig.flyteClient.scopes", testValue)
			if vStringSlice, err := cmdFlags.GetStringSlice("appAuth.thirdPartyConfig.flyteClient.scopes"); err == nil {
				testDecodeRaw_Config(t, join_Config(vStringSlice, ","), &actual.AppAuth.ThirdParty.FlyteClientConfig.Scopes)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_appAuth.thirdPartyConfig.flyteClient.audience", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("appAuth.thirdPartyConfig.flyteClient.audience", testValue)
			if vString, err := cmdFlags.GetString("appAuth.thirdPartyConfig.flyteClient.audience"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.AppAuth.ThirdParty.FlyteClientConfig.Audience)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_rbac.enabled", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("rbac.enabled", testValue)
			if vBool, err := cmdFlags.GetBool("rbac.enabled"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.Rbac.Enabled)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_rbac.bypassMethodPatterns", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := join_Config(DefaultConfig.Rbac.BypassMethodPatterns, ",")

			cmdFlags.Set("rbac.bypassMethodPatterns", testValue)
			if vStringSlice, err := cmdFlags.GetStringSlice("rbac.bypassMethodPatterns"); err == nil {
				testDecodeRaw_Config(t, join_Config(vStringSlice, ","), &actual.Rbac.BypassMethodPatterns)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_rbac.tokenScopeRoleResolver.enabled", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("rbac.tokenScopeRoleResolver.enabled", testValue)
			if vBool, err := cmdFlags.GetBool("rbac.tokenScopeRoleResolver.enabled"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.Rbac.TokenScopeRoleResolver.Enabled)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_rbac.tokenClaimRoleResolver.enabled", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("rbac.tokenClaimRoleResolver.enabled", testValue)
			if vBool, err := cmdFlags.GetBool("rbac.tokenClaimRoleResolver.enabled"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.Rbac.TokenClaimRoleResolver.Enabled)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
}
