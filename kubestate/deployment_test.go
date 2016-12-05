package kubestate

import (
	"testing"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"

	. "github.com/smartystreets/goconvey/convey"
)

var desiredReplicas int32 = 16

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
		Spec: v1beta1.DeploymentSpec{
			Replicas: &desiredReplicas,
		},
	},
	{
		ObjectMeta: v1.ObjectMeta{
			Name:      "NoDesiredReplicas",
			Namespace: "default",
		},
		Spec: v1beta1.DeploymentSpec{},
	},
	{
		ObjectMeta: v1.ObjectMeta{
			Name:       "PausedDeploy",
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
		Spec: v1beta1.DeploymentSpec{
			Replicas: &desiredReplicas,
			Paused:   true,
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
			"grafanalabs.kubestate.deployment.default.BeingDeployed.status.targetedreplicas 15",
			"grafanalabs.kubestate.deployment.default.BeingDeployed.status.availablereplicas 10",
			"grafanalabs.kubestate.deployment.default.BeingDeployed.status.unavailablereplicas 5",
			"grafanalabs.kubestate.deployment.default.BeingDeployed.status.updatedreplicas 2",
			"grafanalabs.kubestate.deployment.default.BeingDeployed.spec.desiredreplicas 16",
			"grafanalabs.kubestate.deployment.default.BeingDeployed.spec.paused 0",
		},
	},
	{
		deployment: mockDeployments[1],
		metrics:    getDeploymentMetricTypes(),
		expected: []string{
			"grafanalabs.kubestate.deployment.default.NoDesiredReplicas.metadata.generation 0",
			"grafanalabs.kubestate.deployment.default.NoDesiredReplicas.status.observedgeneration 0",
			"grafanalabs.kubestate.deployment.default.NoDesiredReplicas.status.targetedreplicas 0",
			"grafanalabs.kubestate.deployment.default.NoDesiredReplicas.status.availablereplicas 0",
			"grafanalabs.kubestate.deployment.default.NoDesiredReplicas.status.unavailablereplicas 0",
			"grafanalabs.kubestate.deployment.default.NoDesiredReplicas.status.updatedreplicas 0",
			"grafanalabs.kubestate.deployment.default.NoDesiredReplicas.spec.paused 0",
		},
	},
	{
		deployment: mockDeployments[2],
		metrics:    getDeploymentMetricTypes(),
		expected: []string{
			"grafanalabs.kubestate.deployment.default.PausedDeploy.metadata.generation 2",
			"grafanalabs.kubestate.deployment.default.PausedDeploy.status.observedgeneration 1",
			"grafanalabs.kubestate.deployment.default.PausedDeploy.status.targetedreplicas 15",
			"grafanalabs.kubestate.deployment.default.PausedDeploy.status.availablereplicas 10",
			"grafanalabs.kubestate.deployment.default.PausedDeploy.status.unavailablereplicas 5",
			"grafanalabs.kubestate.deployment.default.PausedDeploy.status.updatedreplicas 2",
			"grafanalabs.kubestate.deployment.default.PausedDeploy.spec.desiredreplicas 16",
			"grafanalabs.kubestate.deployment.default.PausedDeploy.spec.paused 1",
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
