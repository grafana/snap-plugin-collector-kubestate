package kubestate

import (
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"k8s.io/client-go/pkg/api/v1"
)

type podCollector struct {
}

func (*podCollector) Collect(mts []plugin.Metric, pod v1.Pod) ([]plugin.Metric, error) {
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
				if ns[8] == "requested" && ns[9] == "cpu" {
					req := c.Resources.Requests

					if cpu, ok := req[v1.ResourceCPU]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(cpu.MilliValue())/1000)
						metrics = append(metrics, metric)
					}
				} else if ns[8] == "requested" && ns[9] == "memory" {
					req := c.Resources.Requests

					if mem, ok := req[v1.ResourceMemory]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(mem.Value()))
						metrics = append(metrics, metric)
					}
				} else if ns[8] == "limits" && ns[9] == "cpu" {
					limits := c.Resources.Limits

					if cpu, ok := limits[v1.ResourceCPU]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(cpu.MilliValue())/1000)
						metrics = append(metrics, metric)
					} else if cpu, ok := limits[v1.ResourceLimitsCPU]; ok {
						metric := createContainerResourcesMetric(mt, ns, pod, c, nodeName, float64(cpu.MilliValue())/1000)
						metrics = append(metrics, metric)
					}
				} else if ns[8] == "limits" && ns[9] == "memory" {
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
	ns[5] = slugify(nodeName)
	ns[6] = pod.Name
	ns[7] = c.Name
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
