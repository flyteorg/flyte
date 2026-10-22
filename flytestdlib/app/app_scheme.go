package app

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// InitAppScheme registers Kubernetes apps/v1 and core/v1 types into the
// client-go scheme so that the K8s client can manage App Deployments and Services.
// Must be called before creating any K8s clients that interact with apps.
func InitAppScheme() error {
	if err := appsv1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}
	return corev1.AddToScheme(scheme.Scheme)
}
