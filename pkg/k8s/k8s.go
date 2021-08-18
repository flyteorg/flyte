package k8s

import (
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type K8s interface {
	CoreV1() corev1.CoreV1Interface
}

var Client K8s

// GetK8sClient return the k8s client from sandbox kubeconfig
func GetK8sClient(cfg, master string) (K8s, error) {
	kubeConfigPath := os.ExpandEnv(cfg)
	kubecfg, err := clientcmd.BuildConfigFromFlags(master, kubeConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error building kubeconfig")
	}
	if Client == nil {
		kubeClient, err := kubernetes.NewForConfig(kubecfg)
		if err != nil {
			return nil, errors.Wrapf(err, "Error building kubernetes clientset")
		}
		return kubeClient, nil
	}
	return Client, nil
}
