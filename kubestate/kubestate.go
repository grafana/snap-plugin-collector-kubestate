package kubestate

import (
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/intelsdi-x/snap-plugin-utilities/logger"
	"k8s.io/client-go/pkg/api/v1"
	// "k8s.io/client-go/rest"
)

type Kubestate struct {
}

type podCollector struct {
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

func (*podCollector) CollectPod(mts []plugin.Metric, pod v1.Pod) ([]plugin.Metric, error) {
	metrics := make([]plugin.Metric, 0)

	for _, mt := range mts {
		ns := mt.Namespace.Strings()

		if ns[2] == "pod" && ns[5] == "status" {
			if ns[6] == "phase" {
				ns[3] = pod.Namespace
				ns[4] = pod.Name
				ns[7] = string(pod.Status.Phase)
				mt.Namespace = plugin.NewNamespace(ns...)

				mt.Data = 1

				mt.Timestamp = time.Now()
				metrics = append(metrics, mt)
			} else if ns[6] == "condition" {
				ns[3] = pod.Namespace
				ns[4] = pod.Name
				mt.Namespace = plugin.NewNamespace(ns...)

				if ns[7] == "ready" {
					mt.Data = boolInt(getPodCondition(pod.Status.Conditions, v1.PodReady))
				}
				if ns[7] == "scheduled" {
					mt.Data = boolInt(getPodCondition(pod.Status.Conditions, v1.PodScheduled))
				}

				mt.Timestamp = time.Now()
				metrics = append(metrics, mt)
			}
		} else if ns[3] == "container" {
			for _, cs := range pod.Status.ContainerStatuses {
				switch ns[8] {
				case "restarts":
					metric := createContainerStatusMetric(mt, ns, pod, cs, cs.RestartCount)
					metrics = append(metrics, metric)

				case "ready":
					metric := createContainerStatusMetric(mt, ns, pod, cs, boolInt(cs.Ready))
					metrics = append(metrics, metric)

				case "waiting":
					metric := createContainerStatusMetric(mt, ns, pod, cs, boolInt(cs.State.Waiting != nil))
					metrics = append(metrics, metric)

				case "running":
					metric := createContainerStatusMetric(mt, ns, pod, cs, boolInt(cs.State.Running != nil))
					metrics = append(metrics, metric)

				case "terminated":
					metric := createContainerStatusMetric(mt, ns, pod, cs, boolInt(cs.State.Terminated != nil))
					metrics = append(metrics, metric)
				}
			}

			nodeName := pod.Spec.NodeName
			for _, c := range pod.Spec.Containers {
				if ns[8] == "requested" {
					req := c.Resources.Requests

					if cpu, ok := req[v1.ResourceCPU]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(cpu.MilliValue())/1000)
						metrics = append(metrics, metric)
					}
				} else if ns[8] == "limits" {
					limits := c.Resources.Limits

					if cpu, ok := limits[v1.ResourceCPU]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(cpu.MilliValue())/1000)
						metrics = append(metrics, metric)
					}
				}
			}
		}
	}

	return metrics, nil
}

func createContainerStatusMetric(mt plugin.Metric, ns []string, pod v1.Pod, cs v1.ContainerStatus, value interface{}) plugin.Metric {
	ns[4] = pod.Namespace
	ns[5] = pod.Name
	ns[6] = cs.Name
	mt.Namespace = plugin.NewNamespace(ns...)

	mt.Data = value

	mt.Timestamp = time.Now()
	return mt
}

func createContainerResourcesMetric(mt plugin.Metric, ns []string, pod v1.Pod, c v1.Container, nodeName string, value interface{}) plugin.Metric {
	ns[4] = pod.Namespace
	ns[5] = pod.Name
	ns[6] = c.Name
	ns[7] = nodeName
	mt.Namespace = plugin.NewNamespace(ns...)

	mt.Data = value

	mt.Timestamp = time.Now()
	return mt
}

func getPodCondition(conditions []v1.PodCondition, t v1.PodConditionType) bool {
	for _, c := range conditions {
		if c.Type == t {
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
			AddStaticElements("limits", "cpu", "cores"),
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
