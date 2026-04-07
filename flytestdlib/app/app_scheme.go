package app

import (
	"k8s.io/client-go/kubernetes/scheme"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// InitAppScheme registers Knative Serving types (Service, Route, Configuration, Revision)
// into the client-go scheme so that the K8s client can manage KService CRDs.
// Must be called before creating any K8s clients that interact with apps.
func InitAppScheme() error {
	return servingv1.AddToScheme(scheme.Scheme)
}
