package kubestate

import (
	"testing"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

var mockDeployments = []v1beta1.Deployment{
	{
		ObjectMeta: v1.ObjectMeta{
			Name:       "BeingDeployed",
			Namespace:  "default",
			Generation: 2,
		},
		Status: v1beta1.DeploymentStatus{
			Replicas:            15,
			AvailableReplicas:   10,
			UnavailableReplicas: 5,
			UpdatedReplicas:     2,
			ObservedGeneration:  1,
		},
	},
}

var deploymentCases = []struct {
	deployment v1beta1.Deployment
	metrics    []plugin.Metric
	expected   []string
}{
	{
		deployment: mockDeployments[0],
		metrics:    getDeploymentMetricTypes(),
		expected: []string{
			"grafanalabs.kubestate.deployment.default.BeingDeployed.metadata.generation 2",
			"grafanalabs.kubestate.deployment.default.BeingDeployed.status.observedgeneration 1",
		},
	},
}

func TestDeploymentCollector(t *testing.T) {
	Convey("When collecting metrics for deployments", t, func() {
		for _, c := range deploymentCases {
			metrics, err := new(deploymentCollector).Collect(c.metrics, c.deployment)
			So(err, ShouldBeNil)
			So(metrics, ShouldNotBeNil)

			So(len(metrics), ShouldEqual, len(c.expected))

			for i, metric := range metrics {
				So(format(&metric), ShouldEqual, c.expected[i])
			}
		}
	})
}
