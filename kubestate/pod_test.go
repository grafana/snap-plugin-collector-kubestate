package kubestate

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

var mockPods = []v1.Pod{
	{
		ObjectMeta: v1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
		},
		Status: v1.PodStatus{
			Phase: "Running",
			Conditions: []v1.PodCondition{
				v1.PodCondition{
					Type:   v1.PodReady,
					Status: v1.ConditionTrue,
				},
				v1.PodCondition{
					Type:   v1.PodScheduled,
					Status: v1.ConditionTrue,
				},
			},
			ContainerStatuses: []v1.ContainerStatus{
				v1.ContainerStatus{
					Name:  "container1",
					Ready: true,
					State: v1.ContainerState{
						Running: &v1.ContainerStateRunning{},
					},
					RestartCount: 3,
				},
			},
		},
		Spec: v1.PodSpec{
			NodeName: "127.0.0.1",
			Containers: []v1.Container{
				v1.Container{
					Name: "container1",
					Resources: v1.ResourceRequirements{
						Requests: map[v1.ResourceName]resource.Quantity{
							v1.ResourceCPU:    resource.MustParse("100m"),
							v1.ResourceMemory: resource.MustParse("100M"),
						},
						Limits: map[v1.ResourceName]resource.Quantity{
							v1.ResourceCPU:    resource.MustParse("200m"),
							v1.ResourceMemory: resource.MustParse("200M"),
						},
					},
				},
			},
		},
	},
	{
		ObjectMeta: v1.ObjectMeta{
			Name:      "pod2",
			Namespace: "kube-system",
		},
		Status: v1.PodStatus{
			Phase: "Pending",
			ContainerStatuses: []v1.ContainerStatus{
				v1.ContainerStatus{
					Name:  "container1",
					Ready: true,
					State: v1.ContainerState{
						Waiting: &v1.ContainerStateWaiting{},
					},
					RestartCount: 3,
				},
				v1.ContainerStatus{
					Name:  "container2",
					Ready: false,
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{},
					},
					RestartCount: 5,
				},
			},
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
			Containers: []v1.Container{
				v1.Container{
					Name: "container1",
					Resources: v1.ResourceRequirements{
						Limits: map[v1.ResourceName]resource.Quantity{
							v1.ResourceLimitsCPU:    resource.MustParse("200m"),
							v1.ResourceLimitsMemory: resource.MustParse("200M"),
						},
					},
				},
				v1.Container{
					Name: "container2",
					Resources: v1.ResourceRequirements{
						Limits: map[v1.ResourceName]resource.Quantity{
							v1.ResourceLimitsCPU:    resource.MustParse("200m"),
							v1.ResourceLimitsMemory: resource.MustParse("200M"),
						},
					},
				},
			},
		},
	},
}

var malformedMetricTypes = []plugin.Metric{
	{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate"),
		Version:   1,
	},
	{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container", "test", "test", "status", "metric^1"),
		Version:   1,
	},
}

var cases = []struct {
	pod      v1.Pod
	metrics  []plugin.Metric
	expected []string
}{
	{
		pod:     mockPods[0],
		metrics: getPodMetricTypes(),
		expected: []string{
			"grafanalabs.kubestate.pod.default.pod1.status.phase.Running.value 1",
			"grafanalabs.kubestate.pod.default.pod1.status.condition.ready 1",
			"grafanalabs.kubestate.pod.default.pod1.status.condition.scheduled 1",
		},
	},
	{
		pod:     mockPods[0],
		metrics: getPodContainerMetricTypes(),
		expected: []string{
			"grafanalabs.kubestate.container.default.pod1.container1.status.restarts 3",
			"grafanalabs.kubestate.container.default.pod1.container1.status.ready 1",
			"grafanalabs.kubestate.container.default.pod1.container1.status.waiting 0",
			"grafanalabs.kubestate.container.default.pod1.container1.status.running 1",
			"grafanalabs.kubestate.container.default.pod1.container1.status.terminated 0",
			"grafanalabs.kubestate.container.default.127_0_0_1.pod1.container1.requested.cpu.cores 0.1",
			"grafanalabs.kubestate.container.default.127_0_0_1.pod1.container1.requested.memory.bytes 1e+08",
			"grafanalabs.kubestate.container.default.127_0_0_1.pod1.container1.limits.cpu.cores 0.2",
			"grafanalabs.kubestate.container.default.127_0_0_1.pod1.container1.limits.memory.bytes 2e+08",
		},
	},
	{
		pod:     mockPods[1],
		metrics: getPodContainerMetricTypes(),
		expected: []string{
			"grafanalabs.kubestate.container.kube-system.pod2.container1.status.restarts 3",
			"grafanalabs.kubestate.container.kube-system.pod2.container2.status.restarts 5",
			"grafanalabs.kubestate.container.kube-system.pod2.container1.status.ready 1",
			"grafanalabs.kubestate.container.kube-system.pod2.container2.status.ready 0",
			"grafanalabs.kubestate.container.kube-system.pod2.container1.status.waiting 1",
			"grafanalabs.kubestate.container.kube-system.pod2.container2.status.waiting 0",
			"grafanalabs.kubestate.container.kube-system.pod2.container1.status.running 0",
			"grafanalabs.kubestate.container.kube-system.pod2.container2.status.running 0",
			"grafanalabs.kubestate.container.kube-system.pod2.container1.status.terminated 0",
			"grafanalabs.kubestate.container.kube-system.pod2.container2.status.terminated 1",
			"grafanalabs.kubestate.container.kube-system.node1.pod2.container1.limits.cpu.cores 0.2",
			"grafanalabs.kubestate.container.kube-system.node1.pod2.container2.limits.cpu.cores 0.2",
			"grafanalabs.kubestate.container.kube-system.node1.pod2.container1.limits.memory.bytes 2e+08",
			"grafanalabs.kubestate.container.kube-system.node1.pod2.container2.limits.memory.bytes 2e+08",
		},
	},
	{
		pod:      mockPods[1],
		metrics:  malformedMetricTypes,
		expected: []string{},
	},
}

func TestPodCollector(t *testing.T) {
	Convey("When collecting metrics for pods", t, func() {
		for _, c := range cases {
			metrics, err := new(podCollector).Collect(c.metrics, c.pod)

			So(err, ShouldBeNil)

			So(metrics, ShouldNotBeNil)
			So(len(metrics), ShouldEqual, len(c.expected))

			for i, metric := range metrics {
				So(format(&metric), ShouldEqual, c.expected[i])
			}
		}
	})
}

func format(m *plugin.Metric) string {
	return fmt.Sprintf("%v %v", strings.Join(m.Namespace.Strings(), "."), m.Data)
}
