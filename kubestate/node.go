package kubestate

import (
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"k8s.io/client-go/pkg/api/v1"
)

type nodeCollector struct {
}

func (*nodeCollector) Collect(mts []plugin.Metric, node v1.Node) ([]plugin.Metric, error) {
	metrics := make([]plugin.Metric, 0)

	for _, mt := range mts {
		ns := mt.Namespace.Strings()

		if ns[5] == "spec" && ns[6] == "unschedulable" {
			metric := createNodeMetric(mt, ns, node, boolInt(node.Spec.Unschedulable))
			metrics = append(metrics, metric)
		} else if ns[6] == "outofdisk" {
			metric := createNodeMetric(mt, ns, node, boolInt(getNodeCondition(node.Status.Conditions, v1.NodeOutOfDisk)))
			metrics = append(metrics, metric)
		} else if ns[6] == "capacity" && ns[7] == "cpu" {
			if cpu, ok := node.Status.Capacity[v1.ResourceCPU]; ok {
				metric := createNodeMetric(mt, ns, node, float64(cpu.MilliValue())/1000)
				metrics = append(metrics, metric)
			}
		} else if ns[6] == "capacity" && ns[7] == "memory" {
			if memory, ok := node.Status.Capacity[v1.ResourceMemory]; ok {
				metric := createNodeMetric(mt, ns, node, float64(memory.Value()))
				metrics = append(metrics, metric)
			}
		} else if ns[6] == "capacity" && ns[7] == "pods" {
			if pods, ok := node.Status.Capacity[v1.ResourcePods]; ok {
				metric := createNodeMetric(mt, ns, node, float64(pods.Value()))
				metrics = append(metrics, metric)
			}
		} else if ns[6] == "allocatable" && ns[7] == "cpu" {
			if cpu, ok := node.Status.Allocatable[v1.ResourceCPU]; ok {
				metric := createNodeMetric(mt, ns, node, float64(cpu.MilliValue())/1000)
				metrics = append(metrics, metric)
			}
		} else if ns[6] == "allocatable" && ns[7] == "memory" {
			if memory, ok := node.Status.Allocatable[v1.ResourceMemory]; ok {
				metric := createNodeMetric(mt, ns, node, float64(memory.Value()))
				metrics = append(metrics, metric)
			}
		} else if ns[6] == "allocatable" && ns[7] == "pods" {
			if pods, ok := node.Status.Allocatable[v1.ResourcePods]; ok {
				metric := createNodeMetric(mt, ns, node, float64(pods.Value()))
				metrics = append(metrics, metric)
			}
		}
	}

	return metrics, nil
}

func createNodeMetric(mt plugin.Metric, ns []string, node v1.Node, value interface{}) plugin.Metric {
	ns[3] = node.Namespace
	ns[4] = slugify(node.Name)
	mt.Namespace = plugin.NewNamespace(ns...)

	mt.Data = value

	mt.Timestamp = time.Now()
	return mt
}

func getNodeCondition(conditions []v1.NodeCondition, t v1.NodeConditionType) bool {
	for _, c := range conditions {
		if c.Type == t && c.Status == v1.ConditionTrue {
			return true
		}
	}

	return false
}
