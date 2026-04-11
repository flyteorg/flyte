package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizationServerType_String(t *testing.T) {
	assert.Equal(t, "Self", AuthorizationServerTypeSelf.String())
	assert.Equal(t, "External", AuthorizationServerTypeExternal.String())
	assert.Equal(t, "AuthorizationServerType(99)", AuthorizationServerType(99).String())
}

func TestAuthorizationServerTypeString(t *testing.T) {
	v, err := AuthorizationServerTypeString("Self")
	require.NoError(t, err)
	assert.Equal(t, AuthorizationServerTypeSelf, v)

	v, err = AuthorizationServerTypeString("External")
	require.NoError(t, err)
	assert.Equal(t, AuthorizationServerTypeExternal, v)

	_, err = AuthorizationServerTypeString("bogus")
	assert.Error(t, err)
}

func TestAuthorizationServerType_JSON(t *testing.T) {
	b, err := json.Marshal(AuthorizationServerTypeExternal)
	require.NoError(t, err)
	assert.JSONEq(t, `"External"`, string(b))

	var out AuthorizationServerType
	require.NoError(t, json.Unmarshal([]byte(`"Self"`), &out))
	assert.Equal(t, AuthorizationServerTypeSelf, out)

	assert.Error(t, json.Unmarshal([]byte(`"bogus"`), &out))
	assert.Error(t, json.Unmarshal([]byte(`42`), &out))
}

func TestAuthorizationServerType_IsA(t *testing.T) {
	assert.True(t, AuthorizationServerTypeSelf.IsAAuthorizationServerType())
	assert.True(t, AuthorizationServerTypeExternal.IsAAuthorizationServerType())
	assert.False(t, AuthorizationServerType(99).IsAAuthorizationServerType())
	assert.ElementsMatch(t,
		[]AuthorizationServerType{AuthorizationServerTypeSelf, AuthorizationServerTypeExternal},
		AuthorizationServerTypeValues())
}

func TestSameSite_String(t *testing.T) {
	assert.Equal(t, "DefaultMode", SameSiteDefaultMode.String())
	assert.Equal(t, "LaxMode", SameSiteLaxMode.String())
	assert.Equal(t, "StrictMode", SameSiteStrictMode.String())
	assert.Equal(t, "NoneMode", SameSiteNoneMode.String())
	assert.Equal(t, "SameSite(99)", SameSite(99).String())
}

func TestSameSiteString(t *testing.T) {
	v, err := SameSiteString("StrictMode")
	require.NoError(t, err)
	assert.Equal(t, SameSiteStrictMode, v)

	_, err = SameSiteString("bogus")
	assert.Error(t, err)
}

func TestSameSite_JSON(t *testing.T) {
	b, err := json.Marshal(SameSiteNoneMode)
	require.NoError(t, err)
	assert.JSONEq(t, `"NoneMode"`, string(b))

	var out SameSite
	require.NoError(t, json.Unmarshal([]byte(`"LaxMode"`), &out))
	assert.Equal(t, SameSiteLaxMode, out)

	assert.Error(t, json.Unmarshal([]byte(`"bogus"`), &out))
}

func TestSameSite_IsA(t *testing.T) {
	assert.True(t, SameSiteDefaultMode.IsASameSite())
	assert.True(t, SameSiteNoneMode.IsASameSite())
	assert.False(t, SameSite(99).IsASameSite())
	assert.Len(t, SameSiteValues(), 4)
}

func TestThirdPartyConfigOptions_IsEmpty(t *testing.T) {
	assert.True(t, ThirdPartyConfigOptions{}.IsEmpty())
	assert.False(t, ThirdPartyConfigOptions{
		FlyteClientConfig: FlyteClientConfig{ClientID: "x"},
	}.IsEmpty())
	assert.False(t, ThirdPartyConfigOptions{
		FlyteClientConfig: FlyteClientConfig{Scopes: []string{"all"}},
	}.IsEmpty())
}

func TestMustParseURL(t *testing.T) {
	u := MustParseURL("https://example.com/path")
	require.NotNil(t, u)
	assert.Equal(t, "example.com", u.Host)

	assert.Panics(t, func() { MustParseURL("://bogus") })
}

func TestDefaultConfig(t *testing.T) {
	require.NotNil(t, DefaultConfig)
	assert.Equal(t, "flyte-authorization", DefaultConfig.HTTPAuthorizationHeader)
	assert.Equal(t, AuthorizationServerTypeSelf, DefaultConfig.AppAuth.AuthServerType)
	assert.Equal(t, SameSiteDefaultMode, DefaultConfig.UserAuth.CookieSetting.SameSitePolicy)
	assert.Contains(t, DefaultConfig.UserAuth.OpenID.Scopes, "openid")
}
