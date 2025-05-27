package webhook

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var (
	defaultManagedImageConfig = ManagedImagesConfig{
		Enabled: false,
		Images: []ManagedImageConfig{
			{
				Regex:              "managed.cr.union.ai/sglang:stable",
				ReplacementPattern: "public.ecr.aws/g1m2l3c1/union-serving-staging/sglang:main-opt",
				Tolerations: []v1.Toleration{
					{
						Key:      "union.ai/containerd-snapshotter",
						Operator: v1.TolerationOpEqual,
						Effect:   v1.TaintEffectNoSchedule,
						Value:    "nydus",
					},
				},
				NodeSelector: map[string]string{
					"union.ai/containerd-snapshotter": "nydus",
				},
			},
		},
	}

	managedImageConfigSection = config.MustRegisterSubsection("images", &defaultManagedImageConfig)
)

type ManagedImagesConfig struct {
	Enabled       bool                 `json:"enabled" pflag:",Enable managed images."`
	Images        []ManagedImageConfig `json:"images" pflag:",List of managed images to use for the webhook."`
	LabelSelector metav1.LabelSelector `json:"labelSelector" pflag:",Label selector to use for the webhook."`
}

type ManagedImageConfig struct {
	Regex              string            `json:"regex" pflag:",Regex to match the image name."`
	ReplacementPattern string            `json:"replacementPattern" pflag:",Replacement pattern to use for the image name."`
	Tolerations        []v1.Toleration   `json:"tolerations" pflag:",Tolerations for the image to match."`
	NodeSelector       map[string]string `json:"nodeSelector" pflag:",Node selector for the image to match."`
	compiledRegex      *regexp.Regexp
}

func (m *ManagedImageConfig) CompiledRegex() (*regexp.Regexp, error) {
	if m.compiledRegex != nil {
		return m.compiledRegex, nil
	}

	compiled, err := regexp.Compile(m.Regex)
	if err != nil {
		return nil, err
	}

	m.compiledRegex = compiled
	return m.compiledRegex, nil
}

const (
	ManagedImageMutatorID = "managed-image"
)

type ManagedImageMutator struct {
	cfg *ManagedImagesConfig
}

func (m ManagedImageMutator) ID() string {
	return ManagedImageMutatorID
}

func (m ManagedImageMutator) Mutate(ctx context.Context, p *v1.Pod) (newP *v1.Pod, changed bool, err *admission.Response) {
	containerChanged := false
	nodeSelectors := make(map[string]string, len(m.cfg.Images))
	tolerations := make([]v1.Toleration, 0, len(m.cfg.Images))

	for _, imgConfig := range m.cfg.Images {
		imgConfigApplied := false
		logger.Infof(ctx, "Applying image config: %v", imgConfig.Regex)
		p.Spec.Containers, imgConfigApplied, err = m.MutateContainers(ctx, imgConfig, p.Spec.Containers)
		if err != nil {
			logger.Errorf(ctx, "Error mutating containers: %v", err)
			return nil, false, err
		}

		p.Spec.InitContainers, containerChanged, err = m.MutateContainers(ctx, imgConfig, p.Spec.InitContainers)
		if err != nil {
			logger.Errorf(ctx, "Error mutating init containers: %v", err)
			return nil, false, err
		}

		imgConfigApplied = imgConfigApplied || containerChanged
		if imgConfigApplied {
			storage.MergeMaps(nodeSelectors, imgConfig.NodeSelector)
			tolerations = append(tolerations, imgConfig.Tolerations...)
		}

		changed = changed || imgConfigApplied
	}

	if changed {
		storage.MergeMaps(nodeSelectors, p.Spec.NodeSelector)
		p.Spec.NodeSelector = nodeSelectors
		p.Spec.Tolerations = append(p.Spec.Tolerations, tolerations...)
	}

	return p, changed, nil
}

func (m ManagedImageMutator) MutateContainers(ctx context.Context, imgConfig ManagedImageConfig, containers []v1.Container) (
	newContainers []v1.Container, changed bool, responseErr *admission.Response) {

	containerChanged := false
	for i := range containers {
		logger.Infof(ctx, "Applying image config to container: %v", containers[i].Name)
		containers[i], containerChanged, responseErr = m.MutateContainer(ctx, imgConfig, containers[i])
		if responseErr != nil {
			logger.Errorf(ctx, "Error mutating container: %v", responseErr)
			return nil, false, responseErr
		}

		changed = changed || containerChanged
	}

	return containers, changed, nil
}

func (m ManagedImageMutator) MutateContainer(ctx context.Context, imgConfig ManagedImageConfig, container v1.Container) (newContainer v1.Container, changed bool, responseErr *admission.Response) {
	r, err := imgConfig.CompiledRegex()
	if err != nil {
		logger.Errorf(ctx, "Error compiling regex: %v", err)
		admissionResponse := admission.Errored(http.StatusInternalServerError, err)
		return container, false, &admissionResponse
	}

	if !r.MatchString(container.Image) {
		logger.Infof(ctx, "Image [%v] does not match regex [%v]", container.Image, imgConfig.Regex)
		return container, false, nil
	}

	logger.Infof(ctx, "Updating image for container [%v]: %v", container.Name, container.Image)
	newImg := m.MutateImage(ctx, imgConfig, container.Image)
	if container.Image != newImg {
		container.Image = newImg
		logger.Infof(ctx, "Updated image to: %v", container.Image)
		changed = true
	}

	return container, changed, nil
}

func (m ManagedImageMutator) MutateImage(ctx context.Context, imgConfig ManagedImageConfig, image string) (newImage string) {
	// TODO(haytham): support some template things
	newImage = imgConfig.ReplacementPattern

	return newImage
}

func (m ManagedImageMutator) LabelSelector() *metav1.LabelSelector {
	return &m.cfg.LabelSelector
}

func GetManagedImagesConfig() *ManagedImagesConfig {
	return managedImageConfigSection.GetConfig().(*ManagedImagesConfig)
}

func NewManagedImageMutator(cfg *ManagedImagesConfig) (ManagedImageMutator, error) {
	if cfg != nil {
		for _, img := range cfg.Images {
			_, err := img.CompiledRegex()
			if err != nil {
				return ManagedImageMutator{}, fmt.Errorf("failed compiling regex [%v]. Error: %w", img.Regex, err)
			}
		}
	}

	return ManagedImageMutator{
		cfg: cfg,
	}, nil
}
