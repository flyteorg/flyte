package update

import (
	"context"
	"fmt"
	"os"

	"github.com/flyteorg/flytectl/clierrors"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
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
	Force       bool   `json:"force" pflag:",do not ask for an acknowledgement during updates."`
}

func (cfg NamedEntityConfig) UpdateNamedEntity(ctx context.Context, name string, project string, domain string, rsType core.ResourceType, cmdCtx cmdCore.CommandContext) error {
	if cfg.Activate && cfg.Archive {
		return fmt.Errorf(clierrors.ErrInvalidStateUpdate)
	}

	id := &admin.NamedEntityIdentifier{
		Project: project,
		Domain:  domain,
		Name:    name,
	}

	namedEntity, err := cmdCtx.AdminClient().GetNamedEntity(ctx, &admin.NamedEntityGetRequest{
		ResourceType: rsType,
		Id:           id,
	})
	if err != nil {
		return fmt.Errorf("update metadata for %s: could not fetch metadata: %w", name, err)
	}

	oldMetadata, newMetadata := composeNamedMetadataEdits(cfg, namedEntity.Metadata)
	patch, err := DiffAsYaml(diffPathBefore, diffPathAfter, oldMetadata, newMetadata)
	if err != nil {
		panic(err)
	}
	if patch == "" {
		fmt.Printf("No changes detected. Skipping the update.\n")
		return nil
	}

	fmt.Printf("The following changes are to be applied.\n%s\n", patch)

	if cfg.DryRun {
		fmt.Printf("skipping UpdateNamedEntity request (dryRun)\n")
		return nil
	}

	if !cfg.Force && !cmdUtil.AskForConfirmation("Continue?", os.Stdin) {
		return fmt.Errorf("update aborted by user")
	}

	_, err = cmdCtx.AdminClient().UpdateNamedEntity(ctx, &admin.NamedEntityUpdateRequest{
		ResourceType: rsType,
		Id:           id,
		Metadata:     newMetadata,
	})
	if err != nil {
		return fmt.Errorf("update metadata for %s: update failed: %w", name, err)
	}

	return nil
}

func composeNamedMetadataEdits(config NamedEntityConfig, current *admin.NamedEntityMetadata) (old *admin.NamedEntityMetadata, new *admin.NamedEntityMetadata) {
	old = &admin.NamedEntityMetadata{}
	new = &admin.NamedEntityMetadata{}

	switch {
	case config.Activate && config.Archive:
		panic("cannot both activate and archive")
	case config.Activate:
		old.State = current.State
		new.State = admin.NamedEntityState_NAMED_ENTITY_ACTIVE
	case config.Archive:
		old.State = current.State
		new.State = admin.NamedEntityState_NAMED_ENTITY_ARCHIVED
	}

	if config.Description != "" {
		old.Description = current.Description
		new.Description = config.Description
	}

	return old, new
}
