package client

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	bridge "github.com/zmalik/kudo-bridge/bridge-controller/pkg/generated/clientset/versioned"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client provides access different K8S clients
type Client struct {
	KubeClient kubernetes.Interface
	Discovery  discovery.DiscoveryInterface
	Bridge     *bridge.Clientset
}

func buildKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		client, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("error creating kubernetes client from %s: %v", kubeconfig, err)
		}
		return client, err
	}
	log.Infof("kubeconfig file: using InClusterConfig.")
	return rest.InClusterConfig()
}

func GetKubeConfig() (*rest.Config, error) {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	return buildKubeConfig(kubeConfigPath)
}

func GetKubeClient() (*Client, error) {
	config, err := GetKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kube config: %v", err)
	}
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	bridge, err := bridge.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	discovery, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return &Client{client, discovery, bridge}, nil
}
