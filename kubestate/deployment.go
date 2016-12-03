package kubestate

import (
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type deploymentCollector struct {
}

func (*deploymentCollector) Collect(mts []plugin.Metric, deployment v1beta1.Deployment) ([]plugin.Metric, error) {
	metrics := make([]plugin.Metric, 0)

	for _, mt := range mts {
		ns := mt.Namespace.Strings()

		if ns[5] == "metadata" && ns[6] == "generation" {
			metric := createDeploymentMetric(mt, ns, deployment, deployment.Generation)
			metrics = append(metrics, metric)
		} else if ns[5] == "status" && ns[6] == "observedgeneration" {
			metric := createDeploymentMetric(mt, ns, deployment, deployment.Status.ObservedGeneration)
			metrics = append(metrics, metric)
		}
	}

	return metrics, nil
}

func createDeploymentMetric(mt plugin.Metric, ns []string, deployment v1beta1.Deployment, value interface{}) plugin.Metric {
	ns[3] = deployment.Namespace
	ns[4] = slugify(deployment.Name)
	mt.Namespace = plugin.NewNamespace(ns...)

	mt.Data = value

	mt.Timestamp = time.Now()
	return mt
}
