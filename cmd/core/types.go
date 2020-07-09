package core

import "context"

type CommandFunc func(ctx context.Context, args []string, cmdCtx CommandContext) error
