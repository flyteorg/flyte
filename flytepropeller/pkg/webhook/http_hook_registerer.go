package webhook

import (
	"net/http"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

//go:generate mockery --output=./mocks --case=underscore -name=HTTPHookRegistererIface

type HTTPHookRegistererIface interface {
	Register(path string, hook http.Handler)
}

type K8sRuntimeHTTPHookRegisterer struct {
	mgr manager.Manager
}

func (k K8sRuntimeHTTPHookRegisterer) Register(path string, hook http.Handler) {
	k.mgr.GetWebhookServer().Register(path, hook)
}
