package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	gotree "github.com/DiSiqueira/GoTree"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/spf13/cobra"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytepropeller/cmd/kubectl-flyte/cmd/printers"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type GetOpts struct {
	*RootOptions
	detailsEnabledFlag bool
	limit              int64
	chunkSize          int64
	showQuota          bool
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
			ctx := context.Background()

			if len(args) > 0 {
				name := args[0]
				return getOpts.getWorkflow(ctx, name)
			}
			return getOpts.listWorkflows()
		},
	}

	getCmd.Flags().BoolVarP(&getOpts.detailsEnabledFlag, "details", "d", false, "If details of node execs are desired.")
	getCmd.Flags().BoolVarP(&getOpts.showQuota, "show-quota", "q", false, "Shows resource quota usage for that resource.")
	getCmd.Flags().Int64VarP(&getOpts.chunkSize, "chunk-size", "c", 100, "Use this much batch size.")
	getCmd.Flags().Int64VarP(&getOpts.limit, "limit", "l", -1, "Only get limit records. -1 => all records.")

	return getCmd
}

func (g *GetOpts) getWorkflow(ctx context.Context, name string) error {
	parts := strings.Split(name, "/")
	if len(parts) > 1 {
		g.ConfigOverrides.Context.Namespace = parts[0]
		name = parts[1]
	}
	w, err := g.flyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(g.ConfigOverrides.Context.Namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return err
	}
	wp := printers.WorkflowPrinter{}
	tree := gotree.New("Workflow")
	w.DataReferenceConstructor = storage.URLPathConstructor{}
	if err := wp.Print(ctx, tree, w); err != nil {
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
		wList, err := g.flyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(g.ConfigOverrides.Context.Namespace).List(context.TODO(), *opts)
		if err != nil {
			return err
		}
		for _, w := range wList.Items {
			_w := w
			if err := f(&_w); err != nil {
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

func (g *GetOpts) iterateOverQuotas(f func(quota *v12.ResourceQuota) error, batchSize int64, limit int64) error {
	if limit > 0 && limit < batchSize {
		batchSize = limit
	}
	t, err := g.GetTimeoutSeconds()
	if err != nil {
		return err
	}
	opts := v1.ListOptions{
		Limit:          batchSize,
		TimeoutSeconds: &t,
	}

	var counter int64
	for {
		rq, err := g.kubeClient.CoreV1().ResourceQuotas(g.ConfigOverrides.Context.Namespace).List(context.TODO(), opts)
		if err != nil {
			return err
		}
		for _, r := range rq.Items {
			_r := r
			if err := f(&_r); err != nil {
				return err
			}
			counter++
			if counter == limit {
				return nil
			}
		}
		if rq.Continue == "" {
			return nil
		}
		opts.Continue = rq.Continue
	}
}

type perNSCounter struct {
	ns        string
	total     int
	succeeded int
	failed    int
	waiting   int
	running   int
	hard      v12.ResourceList
	used      v12.ResourceList
}

func Header() string {
	return fmt.Sprintf("|%40s|%7s|%7s|%7s|%7s|%7s|%40s|", "Namespace", "Total", "Success", "Failed", "Running", "Waiting", "QuotasUsage")
}

func (v perNSCounter) CalculateQuotaString() string {
	sb := strings.Builder{}
	for k, q := range v.hard {
		uq := v.used[k]
		used, ok := uq.AsInt64()
		if !ok {
			continue
		}
		hard, ok := q.AsInt64()
		if !ok {
			continue
		}
		per := float64(used) / float64(hard) * 100.0
		sb.WriteString(fmt.Sprintf("%s=%.2f%%,", k, per))
	}
	return sb.String()
}

func (v perNSCounter) String(quotas bool) string {
	s := "-"
	if quotas {
		s = v.CalculateQuotaString()
	}
	return fmt.Sprintf("|%40s|%7d|%7d|%7d|%7d|%7d|%40s|", v.ns, v.total, v.succeeded, v.failed, v.running, v.waiting, s)
}

type SortableNSDist []*perNSCounter

func (s SortableNSDist) Len() int {
	return len(s)
}

func (s SortableNSDist) Less(i, j int) bool {
	if s[i].total == s[j].total {
		return s[i].running < s[j].running
	}
	return s[i].total < s[j].total
}

func (s SortableNSDist) Swap(i, j int) {
	w := s[i]
	s[i] = s[j]
	s[j] = w
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
	perNS := make(map[string]*perNSCounter)
	err := g.iterateOverWorkflows(
		func(w *v1alpha1.FlyteWorkflow) error {
			counter++
			if err := wp.PrintShort(workflows, w); err != nil {
				return err
			}
			if _, ok := perNS[w.Namespace]; !ok {
				perNS[w.Namespace] = &perNSCounter{
					ns: w.Namespace,
				}
			}
			c := perNS[w.Namespace]
			c.total++
			switch w.GetExecutionStatus().GetPhase() {
			case v1alpha1.WorkflowPhaseReady:
				c.waiting++
				waiting++
			case v1alpha1.WorkflowPhaseSuccess:
				c.succeeded++
				succeeded++
			case v1alpha1.WorkflowPhaseFailed:
				c.failed++
				failed++
			default:
				running++
				c.running++
			}
			if counter%g.chunkSize == 0 {
				if g.detailsEnabledFlag {
					fmt.Println("")
					fmt.Print(workflows.Print())
				}
				workflows = gotree.New("\nworkflows")
			} else {
				fmt.Print(".")
			}
			return nil
		}, g.chunkSize, g.limit)
	if err != nil {
		return err
	}

	if g.showQuota {
		err := g.iterateOverQuotas(func(q *v12.ResourceQuota) error {
			if _, ok := perNS[q.Namespace]; !ok {
				perNS[q.Namespace] = &perNSCounter{
					ns: q.Namespace,
				}
			}
			c := perNS[q.Namespace]
			c.hard = q.Status.Hard
			c.used = q.Status.Used
			return nil
		}, g.chunkSize, g.limit)
		if err != nil {
			return err
		}
	}

	if g.detailsEnabledFlag {
		fmt.Print(workflows.Print())
	}
	fmt.Printf("\nFound %d workflows\n", counter)
	fmt.Printf("Success: %d, Failed: %d, Running: %d, Waiting: %d\n", succeeded, failed, running, waiting)

	perNSDist := make(SortableNSDist, 0, len(perNS))
	for _, v := range perNS {
		perNSDist = append(perNSDist, v)
	}
	sort.Sort(perNSDist)

	fmt.Println(Header())
	for _, v := range perNSDist {
		fmt.Println(v.String(g.showQuota))
	}
	return nil
}
