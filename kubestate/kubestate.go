package kubestate

import (
	"strings"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/intelsdi-x/snap-plugin-utilities/logger"
	// "k8s.io/client-go/rest"
)

type Kubestate struct {
}

// CollectMetrics collects metrics for testing
func (n *Kubestate) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	LogDebug("request to collect metrics", "metric_count", len(mts))
	metrics := make([]plugin.Metric, 0)

	incluster := true
	kubeconfigpath := ""

	// incluster, err := mts[0].Config.GetBool("incluster")
	// if err != nil {
	// 	LogError("failed to fetch config value incluster.", "error", err)
	// 	incluster = true
	// }

	// kubeconfigpath, err := mts[0].Config.GetString("kubeconfigpath")
	// if err != nil {
	// 	LogError("failed to fetch config value kubeconfigpath.", "error", err)
	// 	return nil, err
	// }

	client, err := NewClient(incluster, kubeconfigpath)
	if err != nil {
		LogError("failed to create Kubernetes api client.", "error", err)
		return nil, err
	}

	pods, err := client.GetPods()
	podCollector := new(podCollector)
	for _, p := range pods.Items {
		podMetrics, _ := podCollector.Collect(mts, p)
		metrics = append(metrics, podMetrics...)
	}

	nodes, err := client.GetNodes()
	nodeCollector := new(nodeCollector)
	for _, n := range nodes.Items {
		nodeMetrics, _ := nodeCollector.Collect(mts, n)
		metrics = append(metrics, nodeMetrics...)
	}

	LogDebug("collecting metrics completed", "metric_count", len(metrics))
	return metrics, nil
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

func getPodMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}
	// Pod metrics
	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase").
			AddDynamicElement("phase", "current phase").
			AddStaticElement("value"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddStaticElement("status").
			AddStaticElements("condition", "ready"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddStaticElement("status").
			AddStaticElements("condition", "scheduled"),
		Version: 1,
	})

	return mts
}

func getPodContainerMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}
	// Pod Container metrics
	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "restarts"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "ready"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "waiting"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "running"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "terminated"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("requested", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("requested", "memory", "bytes"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("limits", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
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
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddStaticElements("spec", "unschedulable"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "outofdisk"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "capacity", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "capacity", "memory", "bytes"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "capacity", "pods"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "allocatable", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "allocatable", "memory", "bytes"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "allocatable", "pods"),
		Version: 1,
	})

	return mts
}

func (n *Kubestate) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	mts := []plugin.Metric{}

	mts = append(mts, getPodMetricTypes()...)
	mts = append(mts, getPodContainerMetricTypes()...)
	mts = append(mts, getNodeMetricTypes()...)

	return mts, nil
}

func (f *Kubestate) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewBoolRule([]string{"grafanalabs", "kubestate"}, "incluster", false)
	policy.AddNewStringRule([]string{"grafanalabs", "kubestate"}, "kubeconfigpath", false)
	return *policy, nil
}
