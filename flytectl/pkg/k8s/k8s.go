package k8s

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
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

// CopyKubeContext copies context fromContext part of fromConfigAccess to toContext part of toConfigAccess.
func CopyKubeContext(fromConfigAccess, toConfigAccess clientcmd.ConfigAccess, fromContext, toContext string) error {
	_, err := toConfigAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	fromStartingConfig, err := fromConfigAccess.GetStartingConfig()
	if err != nil {
		return err
	}
	_, exists := fromStartingConfig.Contexts[fromContext]
	if !exists {
		return fmt.Errorf("context %v doesn't exist", fromContext)
	}

	toStartingConfig, err := toConfigAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	_, exists = toStartingConfig.Contexts[toContext]
	if exists {
		fmt.Printf("context %v already exist. Overwriting it\n", toContext)
	} else {
		toStartingConfig.Contexts[toContext] = clientcmdapi.NewContext()
	}

	toStartingConfig.Clusters[toContext] = fromStartingConfig.Clusters[fromContext]
	toStartingConfig.Clusters[toContext].LocationOfOrigin = toConfigAccess.GetDefaultFilename()
	toStartingConfig.AuthInfos[toContext] = fromStartingConfig.AuthInfos[fromContext]
	toStartingConfig.AuthInfos[toContext].LocationOfOrigin = toConfigAccess.GetDefaultFilename()
	toStartingConfig.Contexts[toContext].Cluster = toContext
	toStartingConfig.Contexts[toContext].AuthInfo = toContext
	toStartingConfig.CurrentContext = toContext

	if err := clientcmd.ModifyConfig(toConfigAccess, *toStartingConfig, true); err != nil {
		return err
	}

	fmt.Printf("context modified for %q and switched over to it.\n", toContext)
	return nil
}

// RemoveKubeContext removes the contextToRemove from the kubeContext pointed to be fromConfigAccess
func RemoveKubeContext(fromConfigAccess clientcmd.ConfigAccess, contextToRemove string) error {
	fromStartingConfig, err := fromConfigAccess.GetStartingConfig()
	if err != nil {
		return err
	}
	_, exists := fromStartingConfig.Contexts[contextToRemove]
	if !exists {
		return fmt.Errorf("context %v doesn't exist", contextToRemove)
	}

	delete(fromStartingConfig.Clusters, contextToRemove)
	delete(fromStartingConfig.AuthInfos, contextToRemove)
	delete(fromStartingConfig.Contexts, contextToRemove)
	fromStartingConfig.CurrentContext = ""

	if err := clientcmd.ModifyConfig(fromConfigAccess, *fromStartingConfig, true); err != nil {
		return err
	}

	fmt.Printf("context removed for %q.\n", contextToRemove)
	return nil
}
