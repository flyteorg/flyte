package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/clierrors"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

//go:generate pflags NamedEntityConfig --default-var namedEntityConfig

var (
	namedEntityConfig = &NamedEntityConfig{}
)

type NamedEntityConfig struct {
	Archive     bool   `json:"archive" pflag:",archive named entity."`
	Activate    bool   `json:"activate" pflag:",activate the named entity."`
	Description string `json:"description" pflag:",description of the named entity."`
}

func (n NamedEntityConfig) UpdateNamedEntity(ctx context.Context, name string, project string, domain string, rsType core.ResourceType, cmdCtx cmdCore.CommandContext) error {
	archiveProject := n.Archive
	activateProject := n.Activate
	if activateProject == archiveProject && activateProject {
		return fmt.Errorf(clierrors.ErrInvalidStateUpdate)
	}
	var nameEntityState admin.NamedEntityState
	if activateProject {
		nameEntityState = admin.NamedEntityState_NAMED_ENTITY_ACTIVE
	} else if archiveProject {
		nameEntityState = admin.NamedEntityState_NAMED_ENTITY_ARCHIVED
	}
	_, err := cmdCtx.AdminClient().UpdateNamedEntity(ctx, &admin.NamedEntityUpdateRequest{
		ResourceType: rsType,
		Id: &admin.NamedEntityIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
		Metadata: &admin.NamedEntityMetadata{
			Description: n.Description,
			State:       nameEntityState,
		},
	})
	return err
}
