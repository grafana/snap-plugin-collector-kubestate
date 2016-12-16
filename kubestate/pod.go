package kubestate

import (
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"k8s.io/client-go/pkg/api/v1"
)

type podCollector struct {
}

const (
	minPodNamespaceSize          = 8
	metricTypeNsPart             = 2
	namespaceNsPart              = metricTypeNsPart + 1
	nodeNsPart                   = metricTypeNsPart + 2
	podNameNsPart                = metricTypeNsPart + 3
	podStatusNsPart              = metricTypeNsPart + 4
	podStatusTypeNsPart          = metricTypeNsPart + 5
	podStatusValueNsPart         = metricTypeNsPart + 6
	containerStatusNsPart        = metricTypeNsPart + 5
	containerStatusValueNsPart   = metricTypeNsPart + 6
	containerResourceNsPart      = metricTypeNsPart + 5
	containerResourceValueNsPart = metricTypeNsPart + 6
)

func (*podCollector) Collect(mts []plugin.Metric, pod v1.Pod) ([]plugin.Metric, error) {
	metrics := make([]plugin.Metric, 0)

	for _, mt := range mts {
		ns := mt.Namespace.Strings()

		if len(ns) < minPodNamespaceSize {
			continue
		}
		if !isValidNamespace(ns) {
			continue
		}

		if ns[metricTypeNsPart] == "pod" && ns[podStatusNsPart] == "status" {
			if ns[podStatusTypeNsPart] == "phase" {
				ns[namespaceNsPart] = pod.Namespace
				ns[nodeNsPart] = slugify(pod.Spec.NodeName)
				ns[podNameNsPart] = pod.Name

				mt.Namespace = plugin.NewNamespace(ns...)

				if ns[podStatusValueNsPart] == string(pod.Status.Phase) {
					mt.Data = 1
				} else {
					mt.Data = 0
				}

				mt.Timestamp = time.Now()
				metrics = append(metrics, mt)
			} else if ns[podStatusTypeNsPart] == "condition" {
				ns[namespaceNsPart] = pod.Namespace
				ns[nodeNsPart] = slugify(pod.Spec.NodeName)
				ns[podNameNsPart] = pod.Name
				mt.Namespace = plugin.NewNamespace(ns...)

				if ns[podStatusValueNsPart] == "ready" {
					mt.Data = boolInt(getPodCondition(pod.Status.Conditions, v1.PodReady))
				}
				if ns[podStatusValueNsPart] == "scheduled" {
					mt.Data = boolInt(getPodCondition(pod.Status.Conditions, v1.PodScheduled))
				}

				mt.Timestamp = time.Now()
				metrics = append(metrics, mt)
			}
		} else if ns[metricTypeNsPart] == "container" {
			if ns[containerStatusNsPart] == "status" {
				for _, cs := range pod.Status.ContainerStatuses {
					switch ns[containerStatusValueNsPart] {
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
			}

			nodeName := pod.Spec.NodeName
			for _, c := range pod.Spec.Containers {
				if ns[containerResourceNsPart] == "requested" && ns[containerResourceValueNsPart] == "cpu" {
					req := c.Resources.Requests

					if cpu, ok := req[v1.ResourceCPU]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(cpu.MilliValue())/1000)
						metrics = append(metrics, metric)
					}
				} else if ns[containerResourceNsPart] == "requested" && ns[containerResourceValueNsPart] == "memory" {
					req := c.Resources.Requests

					if mem, ok := req[v1.ResourceMemory]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(mem.Value()))
						metrics = append(metrics, metric)
					}
				} else if ns[containerResourceNsPart] == "limits" && ns[containerResourceValueNsPart] == "cpu" {
					limits := c.Resources.Limits

					if cpu, ok := limits[v1.ResourceCPU]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(cpu.MilliValue())/1000)
						metrics = append(metrics, metric)
					} else if cpu, ok := limits[v1.ResourceLimitsCPU]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(cpu.MilliValue())/1000)
						metrics = append(metrics, metric)
					}
				} else if ns[containerResourceNsPart] == "limits" && ns[containerResourceValueNsPart] == "memory" {
					limits := c.Resources.Limits

					if mem, ok := limits[v1.ResourceMemory]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(mem.Value()))
						metrics = append(metrics, metric)
					} else if mem, ok := limits[v1.ResourceLimitsMemory]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(mem.Value()))
						metrics = append(metrics, metric)
					}
				}
			}
		}
	}

	return metrics, nil
}

func createContainerStatusMetric(mt plugin.Metric, ns []string, pod v1.Pod, cs v1.ContainerStatus, value interface{}) plugin.Metric {
	ns[namespaceNsPart] = pod.Namespace
	ns[nodeNsPart] = slugify(pod.Spec.NodeName)
	ns[nodeNsPart+1] = pod.Name
	ns[nodeNsPart+2] = cs.Name
	mt.Namespace = plugin.NewNamespace(ns...)

	mt.Data = value

	mt.Timestamp = time.Now()
	return mt
}

func createContainerResourcesMetric(mt plugin.Metric, ns []string, pod v1.Pod, c v1.Container, nodeName string, value interface{}) plugin.Metric {
	ns[namespaceNsPart] = pod.Namespace
	ns[namespaceNsPart+1] = slugify(nodeName)
	ns[namespaceNsPart+2] = pod.Name
	ns[namespaceNsPart+3] = c.Name
	mt.Namespace = plugin.NewNamespace(ns...)

	mt.Data = value

	mt.Timestamp = time.Now()
	return mt
}

func getPodCondition(conditions []v1.PodCondition, t v1.PodConditionType) bool {
	for _, c := range conditions {
		if c.Type == t && c.Status == v1.ConditionTrue {
			return true
		}
	}

	return false
}
