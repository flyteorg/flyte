package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// EnsureNamespaceExists creates the given namespace if it does not already exist.
func EnsureNamespaceExists(ctx context.Context, k8sClient client.Client, name string) error {
	ns := &corev1.Namespace{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, ns)
	if err == nil {
		logger.Infof(ctx, "Namespace '%s' already exists", name)
		return nil
	}

	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to check namespace: %w", err)
	}

	logger.Infof(ctx, "Creating namespace '%s'", name)
	ns = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	if err := k8sClient.Create(ctx, ns); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	logger.Infof(ctx, "Created namespace '%s'", name)
	return nil
}
