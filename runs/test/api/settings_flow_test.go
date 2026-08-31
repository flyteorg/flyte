package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings/settingsconnect"
)

// mustJSON unmarshals a protojson literal from settings_customer_flow.md into
// msg, failing the test on bad input
func mustJSON(t *testing.T, in string, msg proto.Message) {
	t.Helper()
	require.NoError(t, protojson.Unmarshal([]byte(in), msg))
}

// expectEqual asserts actual matches the doc's expected JSON, printing both on failure.
func expectEqual(t *testing.T, step string, expected, actual proto.Message) {
	t.Helper()
	if !proto.Equal(expected, actual) {
		t.Errorf("%s mismatch\nexpected: %s\nactual:   %s",
			step, protojson.Format(expected), protojson.Format(actual))
	}
}

// TestSettingsCustomerFlow replays flyteidl2/settings/settings_customer_flow.md
// against the running test server. Every JSON literal below is copied from that
// document, so a failure here means the two have drifted apart.
func TestSettingsCustomerFlow(t *testing.T) {
	ctx := context.Background()
	cleanupTestDB(t)

	client := settingsconnect.NewSettingsServiceClient(newClient(), endpoint)

	// --- Starting state: seed the two rows from the doc's Database table ---
	seedOrg := &settings.CreateSettingsRequest{}
	mustJSON(t, `{
	  "key": { "org": "acme" },
	  "settings": {
	    "run": { "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "default" } },
	    "taskResource": { "min": { "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "500m" } } },
	    "environmentVariables": { "state": "SETTING_STATE_VALUE", "mapValue": { "entries": { "LOG_LEVEL": "info", "REGION": "us-east-1" } } }
	  }
	}`, seedOrg)
	_, err := client.CreateSettings(ctx, connect.NewRequest(seedOrg))
	require.NoError(t, err)

	seedProject := &settings.CreateSettingsRequest{}
	mustJSON(t, `{
	  "key": { "org": "acme", "domain": "production", "project": "analytics" },
	  "settings": {
	    "taskResource": { "min": { "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "1000m" } } }
	  }
	}`, seedProject)
	_, err = client.CreateSettings(ctx, connect.NewRequest(seedProject))
	require.NoError(t, err)

	// --- Step 1: GetSettings at org scope ---
	step1Req := &settings.GetSettingsRequest{}
	mustJSON(t, `{ "key": { "org": "acme" } }`, step1Req)
	step1, err := client.GetSettings(ctx,
		connect.NewRequest(step1Req))
	require.NoError(t, err)

	step1Want := &settings.GetSettingsResponse{}
	mustJSON(t, `{
	  "settingsRecord": {
	    "key": { "org": "acme" },
	    "settings": {
	      "run": { "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "default" } },
	      "taskResource": { "min": { "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "500m" } } },
	      "environmentVariables": { "state": "SETTING_STATE_VALUE", "mapValue": { "entries": { "LOG_LEVEL": "info", "REGION": "us-east-1" } } }
	    }
	  }
	}`, step1Want)
	expectEqual(t, "step 1", step1Want, step1.Msg)

	// --- Step 2a: GetSettingsForEdit at project scope ---
	step2aReq := &settings.GetSettingsForEditRequest{}
	mustJSON(t, `{ "key": { "org": "acme", "domain": "production", "project": "analytics" } }`, step2aReq)
	step2a, err := client.GetSettingsForEdit(ctx, connect.NewRequest(step2aReq))
	require.NoError(t, err)

	step2aWant := &settings.GetSettingsForEditResponse{}
	mustJSON(t, `{
	  "requestedKey": { "org": "acme", "domain": "production", "project": "analytics" },
	  "levels": [
	    {
	      "key": { "org": "acme" },
	      "settings": {
	        "run": { "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "default" } },
	        "taskResource": { "min": { "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "500m" } } },
	        "environmentVariables": { "state": "SETTING_STATE_VALUE", "mapValue": { "entries": { "LOG_LEVEL": "info", "REGION": "us-east-1" } } }
	      },
	      "version": "1"
	    },
	    {
	      "key": { "org": "acme", "domain": "production" },
	      "settings": {}
	    },
	    {
	      "key": { "org": "acme", "domain": "production", "project": "analytics" },
	      "settings": {
	        "taskResource": { "min": { "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "1000m" } } }
	      },
	      "version": "1"
	    }
	  ]
	}`, step2aWant)
	expectEqual(t, "step 2a", step2aWant, step2a.Msg)

	// --- Step 2b: UpdateSettings at project scope ---
	step2bReq := &settings.UpdateSettingsRequest{}
	mustJSON(t, `{
	  "key": { "org": "acme", "domain": "production", "project": "analytics" },
	  "settings": {
	    "run": { "defaultQueue": {} },
	    "taskResource": { "min": { "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "2000m" } } },
	    "environmentVariables": { "state": "SETTING_STATE_VALUE", "mapValue": { "entries": { "LOG_LEVEL": "debug" } } }
	  },
	  "version": "1"
	}`, step2bReq)
	step2b, err := client.UpdateSettings(ctx, connect.NewRequest(step2bReq))
	require.NoError(t, err)

	step2bWant := &settings.UpdateSettingsResponse{}
	mustJSON(t, `{
	  "settingsRecord": {
	    "key": { "org": "acme", "domain": "production", "project": "analytics" },
	    "settings": {
	      "taskResource": { "min": { "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "2000m" } } },
	      "environmentVariables": { "state": "SETTING_STATE_VALUE", "mapValue": { "entries": { "LOG_LEVEL": "debug" } } }
	    },
	    "version": "2"
	  }
	}`, step2bWant)
	expectEqual(t, "step 2b", step2bWant, step2b.Msg)

	// --- Step 3: GetSettings at project scope ---
	step3Req := &settings.GetSettingsRequest{}
	mustJSON(t, `{ "key": { "org": "acme", "domain": "production", "project": "analytics" } }`, step3Req)
	step3, err := client.GetSettings(ctx, connect.NewRequest(step3Req))
	require.NoError(t, err)

	step3Want := &settings.GetSettingsResponse{}
	mustJSON(t, `{
	  "settingsRecord": {
	    "key": { "org": "acme", "domain": "production", "project": "analytics" },
	    "settings": {
	      "run": { "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "default" } },
	      "taskResource": { "min": { "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "2000m", "scopeLevel": "SCOPE_LEVEL_PROJECT" } } },
	      "environmentVariables": { "state": "SETTING_STATE_VALUE", "mapValue": { "entries": { "LOG_LEVEL": "debug", "REGION": "us-east-1" } }, "scopeLevel": "SCOPE_LEVEL_PROJECT" }
	    }
	  }
	}`, step3Want)
	expectEqual(t, "step 3", step3Want, step3.Msg)

	// --- Step 4: CreateSettings at domain scope ---
	step4Req := &settings.CreateSettingsRequest{}
	mustJSON(t, `{
	  "key": { "org": "acme", "domain": "production" },
	  "settings": {
	    "run": { "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "fast-queue" } },
	    "taskResource": { "min": { "cpu": {} } },
	    "environmentVariables": {}
	  }
	}`, step4Req)
	step4, err := client.CreateSettings(ctx, connect.NewRequest(step4Req))
	require.NoError(t, err)

	step4Want := &settings.CreateSettingsResponse{}
	mustJSON(t, `{
	  "settingsRecord": {
	    "key": { "org": "acme", "domain": "production" },
	    "settings": {
	      "run": { "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "fast-queue" } }
	    },
	    "version": "1"
	  }
	}`, step4Want)
	expectEqual(t, "step 4", step4Want, step4.Msg)

	// --- Step 5: GetSettings at project scope, after the domain exists ---
	step5, err := client.GetSettings(ctx, connect.NewRequest(step3Req))
	require.NoError(t, err)

	step5Want := &settings.GetSettingsResponse{}
	mustJSON(t, `{
	  "settingsRecord": {
	    "key": { "org": "acme", "domain": "production", "project": "analytics" },
	    "settings": {
	      "run": { "defaultQueue": { "state": "SETTING_STATE_VALUE", "stringValue": "fast-queue", "scopeLevel": "SCOPE_LEVEL_DOMAIN" } },
	      "taskResource": { "min": { "cpu": { "state": "SETTING_STATE_VALUE", "quantityValue": "2000m", "scopeLevel": "SCOPE_LEVEL_PROJECT" } } },
	      "environmentVariables": { "state": "SETTING_STATE_VALUE", "mapValue": { "entries": { "LOG_LEVEL": "debug", "REGION": "us-east-1" } }, "scopeLevel": "SCOPE_LEVEL_PROJECT" }
	    }
	  }
	}`, step5Want)
	expectEqual(t, "step 5", step5Want, step5.Msg)
}
