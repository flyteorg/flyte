package webhook

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
)

const (
	testNamespace = "test-namespace"
)

var (
	expectedSecretsLabelSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			secrets.PodLabel: secrets.PodLabelValue,
		},
	}
	expectedImageBuilderLabelSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"test-arbitrary-label": "test-arbitrary-value",
		},
	}
	testImageBuilderConfig = config.ImageBuilderConfig{
		Enabled: true,
		HostnameReplacement: config.HostnameReplacement{
			Existing:    "test.existing.hostname",
			Replacement: "test.hostname",
		},
		LabelSelector: expectedImageBuilderLabelSelector,
	}
	secretManagerConfig = config.EmbeddedSecretManagerConfig{
		FileMountInitContainer: config.FileMountInitContainerConfig{
			ContainerName: config.EmbeddedSecretsFileMountInitContainerName,
		},
	}
)
