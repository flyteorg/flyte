package cmd

import (
	"fmt"
	"strings"

	gotree "github.com/DiSiqueira/GoTree"
	"github.com/lyft/flytepropeller/cmd/kubectl-flyte/cmd/printers"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GetOpts struct {
	*RootOptions
	detailsEnabledFlag bool
	limit              int64
	chunkSize          int64
}

func NewGetCommand(opts *RootOptions) *cobra.Command {

	getOpts := &GetOpts{
		RootOptions: opts,
	}

	getCmd := &cobra.Command{
		Use:   "get [opts] [<workflow_name>]",
		Short: "Gets a single workflow or lists all workflows currently in execution",
		Long:  `use labels to filter`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				name := args[0]
				return getOpts.getWorkflow(name)
			}
			return getOpts.listWorkflows()
		},
	}

	getCmd.Flags().BoolVarP(&getOpts.detailsEnabledFlag, "details", "d", false, "If details of node execs are desired.")
	getCmd.Flags().Int64VarP(&getOpts.chunkSize, "chunk-size", "c", 100, "Use this much batch size.")
	getCmd.Flags().Int64VarP(&getOpts.limit, "limit", "l", -1, "Only get limit records. -1 => all records.")

	return getCmd
}

func (g *GetOpts) getWorkflow(name string) error {
	parts := strings.Split(name, "/")
	if len(parts) > 1 {
		g.ConfigOverrides.Context.Namespace = parts[0]
		name = parts[1]
	}
	w, err := g.flyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(g.ConfigOverrides.Context.Namespace).Get(name, v1.GetOptions{})
	if err != nil {
		return err
	}
	wp := printers.WorkflowPrinter{}
	tree := gotree.New("Workflow")
	if err := wp.Print(tree, w); err != nil {
		return err
	}
	fmt.Print(tree.Print())
	return nil
}

func (g *GetOpts) iterateOverWorkflows(f func(*v1alpha1.FlyteWorkflow) error, batchSize int64, limit int64) error {
	if limit > 0 && limit < batchSize {
		batchSize = limit
	}
	t, err := g.GetTimeoutSeconds()
	if err != nil {
		return err
	}
	opts := &v1.ListOptions{
		Limit:          batchSize,
		TimeoutSeconds: &t,
	}
	var counter int64
	for {
		wList, err := g.flyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(g.ConfigOverrides.Context.Namespace).List(*opts)
		if err != nil {
			return err
		}
		for _, w := range wList.Items {
			if err := f(&w); err != nil {
				return err
			}
			counter++
			if counter == limit {
				return nil
			}
		}
		if wList.Continue == "" {
			return nil
		}
		opts.Continue = wList.Continue
	}
}

func (g *GetOpts) listWorkflows() error {
	fmt.Printf("Listing workflows in [%s]\n", g.ConfigOverrides.Context.Namespace)
	wp := printers.WorkflowPrinter{}
	workflows := gotree.New("workflows")
	var counter int64
	var succeeded = 0
	var failed = 0
	var running = 0
	var waiting = 0
	err := g.iterateOverWorkflows(
		func(w *v1alpha1.FlyteWorkflow) error {
			counter++
			if err := wp.PrintShort(workflows, w); err != nil {
				return err
			}
			switch w.GetExecutionStatus().GetPhase() {
			case v1alpha1.WorkflowPhaseReady:
				waiting++
			case v1alpha1.WorkflowPhaseSuccess:
				succeeded++
			case v1alpha1.WorkflowPhaseFailed:
				failed++
			default:
				running++
			}
			if counter%g.chunkSize == 0 {
				fmt.Println("")
				fmt.Print(workflows.Print())
				workflows = gotree.New("\nworkflows")
			} else {
				fmt.Print(".")
			}
			return nil
		}, g.chunkSize, g.limit)
	if err != nil {
		return err
	}

	fmt.Print(workflows.Print())
	fmt.Printf("Found %d workflows\n", counter)
	fmt.Printf("Success: %d, Failed: %d, Running: %d, Waiting: %d\n", succeeded, failed, running, waiting)
	return nil
}
