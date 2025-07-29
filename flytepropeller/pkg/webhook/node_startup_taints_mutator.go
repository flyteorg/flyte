package webhook

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

var (
	defaultNodeStartupTaintsConfig = NodeStartupTaintsConfig{
		Enabled: false,
		StartupTaints: []corev1.Taint{
			{
				Key:    "startup-taint.cluster-autoscaler.kubernetes.io/unionai-node-not-ready",
				Value:  "true",
				Effect: corev1.TaintEffectNoSchedule,
			},
		},
		LabelSelector: metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "flyte.org/node-role",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"worker"},
				},
			},
		},
	}

	nodeStartupTaintsConfigSection = config.MustRegisterSubsection("nodeStartupTaints", &defaultNodeStartupTaintsConfig)
)

type NodeStartupTaintsConfig struct {
	Enabled       bool                 `json:"enabled" pflag:",Enable node startup taint mutation."`
	StartupTaints []corev1.Taint       `json:"startupTaints" pflag:",Configuration for the startup taint to apply."`
	LabelSelector metav1.LabelSelector `json:"labelSelector" pflag:",Label selector to use for the webhook."`
}

const (
	NodeStartupTaintsMutatorID = "node-startup-taints"
)

type NodeStartupTaintsMutator struct {
	cfg *NodeStartupTaintsConfig
}

func (m NodeStartupTaintsMutator) ID() string {
	return NodeStartupTaintsMutatorID
}

func (m NodeStartupTaintsMutator) Mutate(ctx context.Context, node *corev1.Node) (newNode *corev1.Node, changed bool, err *admission.Response) {
	if !m.cfg.Enabled {
		return node, false, nil
	}

	// Check if the node matches our label selector
	if !m.matchesLabelSelector(node) {
		logger.Infof(ctx, "Node [%v] does not match label selector", node.Name)
		return node, false, nil
	}

	// Track taints in map to do quick lookups, and also to avoid duplicate keys
	taintsMap := make(map[string]corev1.Taint)
	for _, taint := range node.Spec.Taints {
		taintsMap[taint.Key] = taint
	}

	for _, taint := range m.cfg.StartupTaints {
		if _, exists := taintsMap[taint.Key]; exists {
			if taintsMap[taint.Key].Value == taint.Value && taintsMap[taint.Key].Effect == taint.Effect {
				logger.Infof(ctx, "Taint [%v] already exists on node [%v]", taint.Key, node.Name)
				continue
			}
			logger.Infof(ctx, "Taint [%v] already exists on node [%v] but with different value or effect", taint.Key, node.Name)
		}

		logger.Infof(ctx, "Adding or updating startup taint [%v=%v:%v] to node [%v]",
			taint.Key, taint.Value, taint.Effect, node.Name)

		taintsMap[taint.Key] = taint
		changed = true
	}

	// Convert the map back to a slice
	node.Spec.Taints = make([]corev1.Taint, 0, len(taintsMap))
	for _, taint := range taintsMap {
		node.Spec.Taints = append(node.Spec.Taints, taint)
	}

	return node, changed, nil
}

// matchesLabelSelector checks if the node matches the configured label selector
func (m NodeStartupTaintsMutator) matchesLabelSelector(node *corev1.Node) bool {
	selector, err := metav1.LabelSelectorAsSelector(&m.cfg.LabelSelector)
	if err != nil {
		return false
	}
	return selector.Matches(labels.Set(node.Labels))
}

func (m NodeStartupTaintsMutator) LabelSelector() *metav1.LabelSelector {
	return &m.cfg.LabelSelector
}

func GetNodeStartupTaintsConfig() *NodeStartupTaintsConfig {
	return nodeStartupTaintsConfigSection.GetConfig().(*NodeStartupTaintsConfig)
}

func NewNodeStartupTaintsMutator(cfg *NodeStartupTaintsConfig) (NodeStartupTaintsMutator, error) {
	return NodeStartupTaintsMutator{
		cfg: cfg,
	}, nil
}
