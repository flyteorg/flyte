package deviceflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyteidl/clients/go/admin/tokenorchestrator"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	audience       = "audience"
	clientID       = "client_id"
	deviceCode     = "device_code"
	grantType      = "grant_type"
	scope          = "scope"
	grantTypeValue = "urn:ietf:params:oauth:grant-type:device_code"
)

const (
	errSlowDown    = "slow_down"
	errAuthPending = "authorization_pending"
)

// OAuthTokenOrError containing the token
type OAuthTokenOrError struct {
	*oauth2.Token
	Error string `json:"error,omitempty"`
}

// TokenOrchestrator implements the main logic to initiate device authorization flow
type TokenOrchestrator struct {
	Config Config
	tokenorchestrator.BaseTokenOrchestrator
}

// StartDeviceAuthorization will initiate the OAuth2 device authorization flow.
func (t TokenOrchestrator) StartDeviceAuthorization(ctx context.Context, dareq DeviceAuthorizationRequest) (*DeviceAuthorizationResponse, error) {
	v := url.Values{clientID: {dareq.ClientID}, scope: {dareq.Scope}, audience: {dareq.Audience}}
	httpReq, err := http.NewRequest("POST", t.ClientConfig.DeviceEndpoint, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	logger.Debugf(ctx, "Sending the following request to start device authorization %v with body %v", httpReq.URL, v.Encode())

	httpResp, err := ctxhttp.Do(ctx, nil, httpReq)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(io.LimitReader(httpResp.Body, 1<<20))
	httpResp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("device authorization request failed due to  %v", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, &oauth2.RetrieveError{
			Response: httpResp,
			Body:     body,
		}
	}

	daresp := &DeviceAuthorizationResponse{}
	err = json.Unmarshal(body, &daresp)
	if err != nil {
		return nil, err
	}

	if len(daresp.VerificationURIComplete) > 0 {
		fmt.Printf("Please open the browser at the url %v containing verification code\n", daresp.VerificationURIComplete)
	} else {
		fmt.Printf("Please open the browser at the url %v and enter following verification code %v\n", daresp.VerificationURI, daresp.UserCode)
	}
	return daresp, nil
}

// PollTokenEndpoint polls the token endpoint until the user authorizes/ denies the app or an error occurs other than slow_down or authorization_pending
func (t TokenOrchestrator) PollTokenEndpoint(ctx context.Context, tokReq DeviceAccessTokenRequest, pollInterval time.Duration) (*oauth2.Token, error) {
	v := url.Values{
		clientID:   {tokReq.ClientID},
		grantType:  {grantTypeValue},
		deviceCode: {tokReq.DeviceCode},
	}

	for {
		httpReq, err := http.NewRequest("POST", t.ClientConfig.Endpoint.TokenURL, strings.NewReader(v.Encode()))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		logger.Debugf(ctx, "Sending the following request to fetch the token %v with body %v", httpReq.URL, v.Encode())

		httpResp, err := ctxhttp.Do(ctx, nil, httpReq)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(io.LimitReader(httpResp.Body, 1<<20))
		httpResp.Body.Close()
		if err != nil {
			return nil, err
		}

		// We cannot check the status code since 400 is returned in case of errAuthPending and errSlowDown in which
		// case the polling has to still continue
		var tokResp DeviceAccessTokenResponse
		err = json.Unmarshal(body, &tokResp)
		if err != nil {
			return nil, err
		}

		// Unmarshalled response if it contains an error then check if we need to increase the polling interval
		if len(tokResp.Error) > 0 {
			if tokResp.Error == errSlowDown || tokResp.Error == errAuthPending {
				pollInterval = pollInterval * 2

			} else {
				return nil, fmt.Errorf("oauth error : %v", tokResp.Error)
			}
		} else {
			// Got the auth token in the response and save it in the cache
			err = t.TokenCache.SaveToken(&tokResp.Token)
			// Saving into the cache is only considered to be a warning in this case.
			if err != nil {
				logger.Warnf(ctx, "failed to save token in the token cache. Error: %w", err)
			}
			return &tokResp.Token, nil
		}
		fmt.Printf("Waiting for %v secs\n", pollInterval.Seconds())
		time.Sleep(pollInterval)
	}
}

// FetchTokenFromAuthFlow starts a webserver to listen to redirect callback from the authorization server at the end
// of the flow. It then launches the browser to authenticate the user.
func (t TokenOrchestrator) FetchTokenFromAuthFlow(ctx context.Context) (*oauth2.Token, error) {
	ctx, cancelNow := context.WithTimeout(ctx, t.Config.Timeout.Duration)
	defer cancelNow()

	var scopes string
	if len(t.ClientConfig.Scopes) > 0 {
		scopes = strings.Join(t.ClientConfig.Scopes, " ")
	}
	daReq := DeviceAuthorizationRequest{ClientID: t.ClientConfig.ClientID, Scope: scopes, Audience: t.ClientConfig.Audience}
	daResp, err := t.StartDeviceAuthorization(ctx, daReq)
	if err != nil {
		return nil, err
	}

	pollInterval := t.Config.PollInterval.Duration // default value of 5 sec poll interval if the authorization response doesn't have interval set
	if daResp.Interval > 0 {
		pollInterval = time.Duration(daResp.Interval) * time.Second
	}

	tokReq := DeviceAccessTokenRequest{ClientID: t.ClientConfig.ClientID, DeviceCode: daResp.DeviceCode, GrantType: grantType}
	return t.PollTokenEndpoint(ctx, tokReq, pollInterval)
}

// NewDeviceFlowTokenOrchestrator creates a new TokenOrchestrator that implements the main logic to start device authorization flow and fetch device code and then poll on the token endpoint until the device authorization is approved/denied by the user
func NewDeviceFlowTokenOrchestrator(baseOrchestrator tokenorchestrator.BaseTokenOrchestrator, cfg Config) (TokenOrchestrator, error) {
	return TokenOrchestrator{
		BaseTokenOrchestrator: baseOrchestrator,
		Config:                cfg,
	}, nil
}
