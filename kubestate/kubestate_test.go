package kubestate

import (
	"fmt"
	"strings"
	"testing"

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
					Type: v1.PodReady,
				},
				v1.PodCondition{
					Type: v1.PodScheduled,
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
	},
}

var podStatusMts = []plugin.Metric{
	plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase").
			AddDynamicElement("phase", "current phase").
			AddStaticElement("value"),
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddStaticElement("status").
			AddStaticElement("condition").
			AddStaticElement("ready"),
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddStaticElement("status").
			AddStaticElement("condition").
			AddStaticElement("scheduled"),
	},
}

var podContainerMts = []plugin.Metric{
	plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElement("status").
			AddStaticElement("restarts"),
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElement("status").
			AddStaticElement("ready"),
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElement("status").
			AddStaticElement("waiting"),
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElement("status").
			AddStaticElement("running"),
	},
	plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElement("status").
			AddStaticElement("terminated"),
	},
}

var cases = []struct {
	pod      v1.Pod
	metrics  []plugin.Metric
	expected []string
}{
	{
		pod:     mockPods[0],
		metrics: podStatusMts,
		expected: []string{
			"grafanalabs.kubestate.pod.default.pod1.status.phase.Running.value 1",
			"grafanalabs.kubestate.pod.default.pod1.status.condition.ready 1",
			"grafanalabs.kubestate.pod.default.pod1.status.condition.scheduled 1",
		},
	},
	{
		pod:     mockPods[0],
		metrics: podContainerMts,
		expected: []string{
			"grafanalabs.kubestate.pod.container.default.pod1.container1.status.restarts 3",
			"grafanalabs.kubestate.pod.container.default.pod1.container1.status.ready 1",
			"grafanalabs.kubestate.pod.container.default.pod1.container1.status.waiting 0",
			"grafanalabs.kubestate.pod.container.default.pod1.container1.status.running 1",
			"grafanalabs.kubestate.pod.container.default.pod1.container1.status.terminated 0",
		},
	},
	{
		pod:     mockPods[1],
		metrics: podContainerMts,
		expected: []string{
			"grafanalabs.kubestate.pod.container.kube-system.pod2.container1.status.restarts 3",
			"grafanalabs.kubestate.pod.container.kube-system.pod2.container2.status.restarts 5",
			"grafanalabs.kubestate.pod.container.kube-system.pod2.container1.status.ready 1",
			"grafanalabs.kubestate.pod.container.kube-system.pod2.container2.status.ready 0",
			"grafanalabs.kubestate.pod.container.kube-system.pod2.container1.status.waiting 1",
			"grafanalabs.kubestate.pod.container.kube-system.pod2.container2.status.waiting 0",
			"grafanalabs.kubestate.pod.container.kube-system.pod2.container1.status.running 0",
			"grafanalabs.kubestate.pod.container.kube-system.pod2.container2.status.running 0",
			"grafanalabs.kubestate.pod.container.kube-system.pod2.container1.status.terminated 0",
			"grafanalabs.kubestate.pod.container.kube-system.pod2.container2.status.terminated 1",
		},
	},
}

func TestPodCollector(t *testing.T) {
	Convey("When collecting metrics for pods", t, func() {
		for _, c := range cases {
			metrics, err := new(podCollector).CollectPod(c.metrics, c.pod)

			So(err, ShouldBeNil)

			So(metrics, ShouldNotBeNil)
			So(len(metrics), ShouldEqual, len(c.expected))

			for i, metric := range metrics {
				So(format(&metric), ShouldEqual, c.expected[i])
			}
		}
	})
}

// func TestKubestate(t *testing.T) {
// 	Convey("When getting meta data", t, func() {

// 	}
// }

func format(m *plugin.Metric) string {
	return fmt.Sprintf("%v %v", strings.Join(m.Namespace.Strings(), "."), m.Data)
}
