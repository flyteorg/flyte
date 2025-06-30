package get

import (
	"context"
	"fmt"

	cmdcore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

const userInfoDesc = "Get the user info."

func UserInfo(ctx context.Context, _ []string, cmdCtx cmdcore.CommandContext) error {
	resp, err := cmdCtx.ClientSet().IdentityClient().UserInfo(ctx, &service.UserInfoRequest{})
	if err != nil {
		return fmt.Errorf("failed to get user info: %v", err)
	}
	fmt.Println(marshalOptions.Format(resp))
	return nil
}
