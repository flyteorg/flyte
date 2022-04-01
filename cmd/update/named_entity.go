package update

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytectl/clierrors"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
)

//go:generate pflags NamedEntityConfig --default-var namedEntityConfig --bind-default-var

var (
	namedEntityConfig = &NamedEntityConfig{}
)

type NamedEntityConfig struct {
	Archive     bool   `json:"archive" pflag:",archive named entity."`
	Activate    bool   `json:"activate" pflag:",activate the named entity."`
	Description string `json:"description" pflag:",description of the named entity."`
	DryRun      bool   `json:"dryRun" pflag:",execute command without making any modifications."`
}

func (cfg NamedEntityConfig) UpdateNamedEntity(ctx context.Context, name string, project string, domain string, rsType core.ResourceType, cmdCtx cmdCore.CommandContext) error {
	archiveProject := cfg.Archive
	activateProject := cfg.Activate
	if activateProject == archiveProject && activateProject {
		return fmt.Errorf(clierrors.ErrInvalidStateUpdate)
	}
	var nameEntityState admin.NamedEntityState
	if activateProject {
		nameEntityState = admin.NamedEntityState_NAMED_ENTITY_ACTIVE
	} else if archiveProject {
		nameEntityState = admin.NamedEntityState_NAMED_ENTITY_ARCHIVED
	}
	if cfg.DryRun {
		logger.Infof(ctx, "skipping UpdateNamedEntity request (dryRun)")
	} else {
		_, err := cmdCtx.AdminClient().UpdateNamedEntity(ctx, &admin.NamedEntityUpdateRequest{
			ResourceType: rsType,
			Id: &admin.NamedEntityIdentifier{
				Project: project,
				Domain:  domain,
				Name:    name,
			},
			Metadata: &admin.NamedEntityMetadata{
				Description: cfg.Description,
				State:       nameEntityState,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
