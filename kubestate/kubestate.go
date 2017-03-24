package kubestate

import (
	"strings"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/intelsdi-x/snap-plugin-utilities/logger"
	. "github.com/intelsdi-x/snap-plugin-utilities/ns"
	// "k8s.io/client-go/rest"
)

type Kubestate struct {
}

// CollectMetrics collects metrics for testing
func (n *Kubestate) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	LogDebug("request to collect metrics", "metric_count", len(mts))

	incluster, err := mts[0].Config.GetBool("incluster")
	if err != nil {
		LogError("failed to fetch config value incluster.", "error", err)
		incluster = true
	}

	kubeconfigpath := ""
	if !incluster {
		kubeconfigpath, err = mts[0].Config.GetString("kubeconfigpath")
		if err != nil {
			LogError("failed to fetch config value kubeconfigpath.", "error", err)
			return nil, err
		}
	}

	client, err := newClient(incluster, kubeconfigpath)
	if err != nil {
		LogError("failed to create Kubernetes api client.", "error", err)
		return nil, err
	}

	metrics, err := collect(client, mts)
	if err != nil {
		return nil, err
	}

	LogDebug("collecting metrics completed", "metric_count", len(metrics))
	return metrics, nil
}

var collect = func(client *Client, mts []plugin.Metric) ([]plugin.Metric, error) {
	metrics := make([]plugin.Metric, 0)

	if shouldCollectMetricsFor("pod", mts) || shouldCollectMetricsFor("container", mts) {
		pods, err := client.GetPods()
		if err != nil {
			return nil, err
		}

		podCollector := new(podCollector)
		for _, p := range pods.Items {
			podMetrics, _ := podCollector.Collect(mts, p)
			metrics = append(metrics, podMetrics...)
		}
	}

	if shouldCollectMetricsFor("node", mts) {
		nodes, err := client.GetNodes()
		if err != nil {
			return nil, err
		}

		nodeCollector := new(nodeCollector)
		for _, n := range nodes.Items {
			nodeMetrics, _ := nodeCollector.Collect(mts, n)
			metrics = append(metrics, nodeMetrics...)
		}
	}

	if shouldCollectMetricsFor("deployment", mts) {
		deployments, err := client.GetDeployments()
		if err != nil {
			return nil, err
		}

		deploymentCollector := new(deploymentCollector)
		for _, d := range deployments.Items {
			deploymentMetrics, _ := deploymentCollector.Collect(mts, d)
			metrics = append(metrics, deploymentMetrics...)
		}
	}

	return metrics, nil
}

func shouldCollectMetricsFor(metricType string, mts []plugin.Metric) bool {
	for _, mt := range mts {
		ns := mt.Namespace.Strings()
		if len(ns) < 3 {
			continue
		}
		if ns[2] == metricType {
			return true
		}
	}
	return false
}

func boolInt(b bool) int {
	if b {
		return 1
	}

	return 0
}

func slugify(name string) string {
	return strings.Replace(name, ".", "_", -1)
}

func isValidNamespace(ns []string) bool {
	for i := range ns {
		err := ValidateMetricNamespacePart(ns[i])
		if err != nil {
			// Logger.
			return false
		}
	}
	return true
}

func getPodMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}
	// Pod metrics
	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase", "Pending"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase", "Running"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase", "Succeeded"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase", "Failed"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase", "Unknown"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElement("status").
			AddStaticElements("condition", "ready"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElement("status").
			AddStaticElements("condition", "scheduled"),
		Version: 1,
	})

	return mts
}

func getPodContainerMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "restarts"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "ready"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "waiting"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "running"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "terminated"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("requested", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("requested", "memory", "bytes"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("limits", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("limits", "memory", "bytes"),
		Version: 1,
	})

	return mts
}

func getNodeMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}

	//Node metrics
	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("spec", "unschedulable"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "outofdisk"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "capacity", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "capacity", "memory", "bytes"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "capacity", "pods"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "allocatable", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "allocatable", "memory", "bytes"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "allocatable", "pods"),
		Version: 1,
	})

	return mts
}

func getDeploymentMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("metadata", "generation"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "observedgeneration"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "targetedreplicas"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "availablereplicas"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "unavailablereplicas"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "updatedreplicas"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("spec", "desiredreplicas"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("spec", "paused"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "deploynotfinished"),
		Version: 1,
	})

	return mts
}

func (n *Kubestate) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	mts := []plugin.Metric{}

	mts = append(mts, getPodMetricTypes()...)
	mts = append(mts, getPodContainerMetricTypes()...)
	mts = append(mts, getNodeMetricTypes()...)
	mts = append(mts, getDeploymentMetricTypes()...)

	return mts, nil
}

func (f *Kubestate) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewBoolRule([]string{"grafanalabs", "kubestate"}, "incluster", false, plugin.SetDefaultBool(true))
	policy.AddNewStringRule([]string{"grafanalabs", "kubestate"}, "kubeconfigpath", false)
	return *policy, nil
}
