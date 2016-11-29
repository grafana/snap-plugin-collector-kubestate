package kubestate

import (
	"testing"

	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

var mockNodes = []v1.Node{
	{
		ObjectMeta: v1.ObjectMeta{
			Name:      "127.0.0.1",
			Namespace: "default",
		},
		Spec: v1.NodeSpec{
			Unschedulable: true,
		},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4.3"),
				v1.ResourceMemory: resource.MustParse("2G"),
				v1.ResourcePods:   resource.MustParse("1000"),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("3"),
				v1.ResourceMemory: resource.MustParse("1G"),
				v1.ResourcePods:   resource.MustParse("555"),
			},
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeOutOfDisk, Status: v1.ConditionFalse},
			},
		},
	},
}

var nodeCases = []struct {
	node     v1.Node
	metrics  []plugin.Metric
	expected []string
}{
	{
		node:    mockNodes[0],
		metrics: getNodeMetricTypes(),
		expected: []string{
			"grafanalabs.kubestate.node.127_0_0_1.spec.unschedulable 1",
			"grafanalabs.kubestate.node.127_0_0_1.status.outofdisk 0",
			"grafanalabs.kubestate.node.127_0_0_1.status.capacity.cpu.cores 4.3",
			"grafanalabs.kubestate.node.127_0_0_1.status.capacity.memory.bytes 2e+09",
			"grafanalabs.kubestate.node.127_0_0_1.status.capacity.pods 1000",
			"grafanalabs.kubestate.node.127_0_0_1.status.allocatable.cpu.cores 3",
			"grafanalabs.kubestate.node.127_0_0_1.status.allocatable.memory.bytes 1e+09",
			"grafanalabs.kubestate.node.127_0_0_1.status.allocatable.pods 555",
		},
	},
}

func TestNodeCollector(t *testing.T) {
	Convey("When collecting metrics for nodes", t, func() {
		for _, c := range nodeCases {
			metrics, err := new(nodeCollector).Collect(c.metrics, c.node)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeNil)

			So(len(metrics), ShouldEqual, len(c.expected))

			for i, metric := range metrics {
				So(format(&metric), ShouldEqual, c.expected[i])
			}
		}
	})
}
