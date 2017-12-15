package kubestate

import (
	"flag"

	. "github.com/intelsdi-x/snap-plugin-utilities/logger"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	v1batch "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *kubernetes.Clientset
}

var newClient = func(incluster bool, kubeconfigpath string) (*Client, error) {
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

func (c *Client) GetPods(namespace, node string) (*v1.PodList, error) {
	opts := v1.ListOptions{}
	if node != "*" {
		opts.FieldSelector = fields.OneTermEqualSelector("spec.nodeName", node).String()
	}
	if namespace == "*" {
		namespace = ""
	}
	return c.clientset.Core().Pods(namespace).List(opts)
}

func (c *Client) GetNodes(node string) (*v1.NodeList, error) {
	opts := v1.ListOptions{}
	if node != "*" {
		opts.FieldSelector = fields.OneTermEqualSelector("metadata.name", node).String()
	}
	return c.clientset.Core().Nodes().List(v1.ListOptions{})
}

func (c *Client) GetDeployments(namespace string) (*v1beta1.DeploymentList, error) {
	if namespace == "*" {
		namespace = ""
	}
	return c.clientset.Extensions().Deployments(namespace).List(v1.ListOptions{})
}

func (c *Client) GetJobs(namespace string) (*v1batch.JobList, error) {
	if namespace == "*" {
		namespace = ""
	}
	return c.clientset.BatchClient.Jobs(namespace).List(v1.ListOptions{})
}
