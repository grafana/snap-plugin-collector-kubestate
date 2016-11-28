package kubestate

import (
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
		podMetrics, _ := podCollector.CollectPod(mts, p)
		metrics = append(metrics, podMetrics...)
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

func (n *Kubestate) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
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
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddDynamicElement("node", "node name").
			AddStaticElements("requested", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddDynamicElement("node", "node name").
			AddStaticElements("requested", "memory", "bytes"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddDynamicElement("node", "node name").
			AddStaticElements("limits", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddDynamicElement("node", "node name").
			AddStaticElements("limits", "memory", "bytes"),
		Version: 1,
	})

	return mts, nil
}

func (f *Kubestate) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewBoolRule([]string{"grafanalabs", "kubestate"}, "incluster", false)
	policy.AddNewStringRule([]string{"grafanalabs", "kubestate"}, "kubeconfigpath", false)
	return *policy, nil
}
