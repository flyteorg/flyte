package get

import (
	"context"
	"fmt"

	cmdcore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

const publicClientCfgDesc = "Get the public client config."

func PublicClientConfig(ctx context.Context, _ []string, cmdCtx cmdcore.CommandContext) error {
	resp, err := cmdCtx.ClientSet().AuthMetadataClient().GetPublicClientConfig(ctx, &service.PublicClientAuthConfigRequest{})
	if err != nil {
		return fmt.Errorf("failed to get public client config: %v", err)
	}
	fmt.Println(marshalOptions.Format(resp))
	return nil
}

const oauthMetadataDesc = "Get the OAuth2 metadata."

func OAuthMetadata(ctx context.Context, _ []string, cmdCtx cmdcore.CommandContext) error {
	resp, err := cmdCtx.ClientSet().AuthMetadataClient().GetOAuth2Metadata(ctx, &service.OAuth2MetadataRequest{})
	if err != nil {
		return fmt.Errorf("failed to get oauth metadata: %v", err)
	}
	fmt.Println(marshalOptions.Format(resp))
	return nil
}
