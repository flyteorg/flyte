package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
)

const (
	ErrIdpClient errors.ErrorCode = "IDP_REQUEST_FAILED"
)

/*
This struct represents what should be returned by an IDP according to the specification at
 https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse

Keep in mind that not all fields are necessarily populated, and additional fields may be present as well. This is a sample
response object returned from Okta for instance
	{
	  "sub": "abc123",
	  "name": "John Smith",
	  "locale": "US",
	  "preferred_username": "jsmith123@company.com",
	  "given_name": "John",
	  "family_name": "Smith",
	  "zoneinfo": "America/Los_Angeles",
	  "updated_at": 1568750854
	}
*/
type UserInfoResponse struct {
	Sub               string `json:"sub"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
	Picture           string `json:"picture"`
}

func postToIdp(ctx context.Context, client *http.Client, userInfoURL, accessToken string) (UserInfoResponse, error) {
	request, err := http.NewRequest(http.MethodPost, userInfoURL, nil)
	if err != nil {
		logger.Errorf(ctx, "Error creating user info request to IDP %s", err)
		return UserInfoResponse{}, errors.Wrapf(ErrIdpClient, err, "Error creating user info request to IDP")
	}
	request.Header.Set(DefaultAuthorizationHeader, fmt.Sprintf("%s %s", BearerScheme, accessToken))
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		logger.Errorf(ctx, "Error while requesting user info from IDP %s", err)
		return UserInfoResponse{}, errors.Wrapf(ErrIdpClient, err, "Error while requesting user info from IDP")
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			logger.Errorf(ctx, "Error closing response body %s", err)
		}
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Errorf(ctx, "Error reading user info response error %s", response.StatusCode, err)
		return UserInfoResponse{}, errors.Wrapf(ErrIdpClient, err, "Error reading user info response")
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		logger.Errorf(ctx, "Bad response code from IDP %d", response.StatusCode)
		return UserInfoResponse{}, errors.Errorf(ErrIdpClient,
			"Error reading user info response, code %d body %v", response.StatusCode, body)
	}

	responseObject := &UserInfoResponse{}
	err = json.Unmarshal(body, responseObject)
	if err != nil {
		return UserInfoResponse{}, errors.Wrapf(ErrIdpClient, err, "Could not unmarshal IDP response")
	}

	return *responseObject, nil
}
