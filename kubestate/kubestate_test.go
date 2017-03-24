package kubestate

import (
	"testing"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKubestate(t *testing.T) {
	Convey("When collect metrics is called", t, func() {
		var inclusterSpy bool
		var kubeConfigSpy string

		newClient = func(incluster bool, kubeconfigpath string) (*Client, error) {
			c := &Client{}
			inclusterSpy = incluster
			kubeConfigSpy = kubeconfigpath
			return c, nil
		}

		collect = func(client *Client, mts []plugin.Metric) ([]plugin.Metric, error) {
			return make([]plugin.Metric, 0), nil
		}

		ks := new(Kubestate)

		metrics, err := ks.CollectMetrics(metricWithInclusterConfig)
		So(err, ShouldBeNil)
		So(metrics, ShouldNotBeNil)
		So(inclusterSpy, ShouldBeTrue)
		So(kubeConfigSpy, ShouldEqual, "")

		metrics, err = ks.CollectMetrics(metricWithKubeConfig)
		So(err, ShouldBeNil)
		So(metrics, ShouldNotBeNil)
		So(inclusterSpy, ShouldBeFalse)
		So(kubeConfigSpy, ShouldEqual, "/home/user/.kube/config")

		metrics, err = ks.CollectMetrics(metricWithNoConfig)
		So(err, ShouldBeNil)
		So(metrics, ShouldNotBeNil)
		So(inclusterSpy, ShouldBeTrue)
		So(kubeConfigSpy, ShouldEqual, "")
	})

	Convey("When checking if metrics contain pod metrics", t, func() {
		shouldCollect := shouldCollectMetricsFor("pod", getDeploymentMetricTypes())
		So(shouldCollect, ShouldBeFalse)

		shouldCollect = shouldCollectMetricsFor("pod", getPodMetricTypes())
		So(shouldCollect, ShouldBeTrue)

		shouldCollect = shouldCollectMetricsFor("container", getPodContainerMetricTypes())
		So(shouldCollect, ShouldBeTrue)

		shouldCollect = shouldCollectMetricsFor("node", getNodeMetricTypes())
		So(shouldCollect, ShouldBeTrue)

		shouldCollect = shouldCollectMetricsFor("deployment", getDeploymentMetricTypes())
		So(shouldCollect, ShouldBeTrue)
	})
}

var metricWithInclusterConfig = []plugin.Metric{
	{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate"),
		Version:   1,
		Config: plugin.Config{
			"incluster": true,
		},
	},
}

var metricWithKubeConfig = []plugin.Metric{
	{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate"),
		Version:   1,
		Config: plugin.Config{
			"incluster":      false,
			"kubeconfigpath": "/home/user/.kube/config",
		},
	},
}

var metricWithNoConfig = []plugin.Metric{
	{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate"),
		Version:   1,
	},
}
