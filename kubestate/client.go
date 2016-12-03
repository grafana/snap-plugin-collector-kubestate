package kubestate

import (
	"flag"

	. "github.com/intelsdi-x/snap-plugin-utilities/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *kubernetes.Clientset
}

func NewClient(incluster bool, kubeconfigpath string) (*Client, error) {
	var config *rest.Config
	var err error

	if incluster {
		config, err = rest.InClusterConfig()
		if err != nil {
			LogError("failed to read the Kubernetes api client config.", "error", err)
			return nil, err
		}
	} else {
		flag.Parse()
		// uses the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigpath)

		if err != nil {
			LogError("failed to read the Kubernetes api client config.", "error", err)
			return nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		LogError("failed to create Kubernetes api client.", "error", err)
		return nil, err
	}

	c := &Client{
		clientset: clientset,
	}

	return c, nil
}

func (c *Client) GetPods() (*v1.PodList, error) {
	return c.clientset.Core().Pods("").List(v1.ListOptions{})
}

func (c *Client) GetNodes() (*v1.NodeList, error) {
	return c.clientset.Core().Nodes().List(v1.ListOptions{})
}

func (c *Client) GetDeployments() (*v1beta1.DeploymentList, error) {
	return c.clientset.Extensions().Deployments("").List(v1.ListOptions{})
}
