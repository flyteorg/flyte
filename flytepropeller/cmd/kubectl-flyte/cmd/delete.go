package cmd

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytepropeller/pkg/controller"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeleteOpts struct {
	*RootOptions
	force        bool
	allCompleted bool
	chunkSize    int64
	limit        int64
}

func NewDeleteCommand(opts *RootOptions) *cobra.Command {

	deleteOpts := &DeleteOpts{
		RootOptions: opts,
	}

	// deleteCmd represents the delete command
	deleteCmd := &cobra.Command{
		Use:   "delete <opts> [workflow-name]",
		Short: "delete a workflow",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				name := args[0]
				return deleteOpts.deleteWorkflow(name)
			}

			return deleteOpts.deleteCompletedWorkflows()
		},
	}

	deleteCmd.Flags().BoolVarP(&deleteOpts.force, "force", "f", false, "Enable force deletion to remove finalizers from a workflow.")
	deleteCmd.Flags().BoolVarP(&deleteOpts.allCompleted, "all-completed", "a", false, "Delete all the workflows that have completed. Cannot be used with --force.")
	deleteCmd.Flags().Int64VarP(&deleteOpts.chunkSize, "chunk-size", "c", 100, "When using all-completed, provide a chunk size to retrieve at once from the server.")
	deleteCmd.Flags().Int64VarP(&deleteOpts.limit, "limit", "l", -1, "Only iterate over max limit records.")

	return deleteCmd
}

func (d *DeleteOpts) deleteCompletedWorkflows() error {
	if d.force && d.allCompleted {
		return fmt.Errorf("cannot delete multiple workflows with --force")
	}
	if !d.allCompleted {
		return fmt.Errorf("all completed | workflow name is required")
	}

	t, err := d.GetTimeoutSeconds()
	if err != nil {
		return err
	}

	p := v1.DeletePropagationBackground
	return d.flyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(d.ConfigOverrides.Context.Namespace).DeleteCollection(
		context.TODO(),
		v1.DeleteOptions{PropagationPolicy: &p}, v1.ListOptions{
			TimeoutSeconds: &t,
			LabelSelector:  v1.FormatLabelSelector(controller.CompletedWorkflowsLabelSelector()),
		},
	)

}

func (d *DeleteOpts) deleteWorkflow(name string) error {
	p := v1.DeletePropagationBackground
	if err := d.flyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(d.ConfigOverrides.Context.Namespace).Delete(context.TODO(), name, v1.DeleteOptions{PropagationPolicy: &p}); err != nil {
		return err
	}
	if d.force {
		w, err := d.flyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(d.ConfigOverrides.Context.Namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return err
		}
		w.SetFinalizers([]string{})
		if _, err := d.flyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(d.ConfigOverrides.Context.Namespace).Update(context.TODO(), w, v1.UpdateOptions{}); err != nil {
			return err
		}
	}
	return nil
}
