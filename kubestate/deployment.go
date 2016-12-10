package kubestate

import (
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

type deploymentCollector struct {
}

const (
	minDeploymentNamespaceSize = 7
)

func (*deploymentCollector) Collect(mts []plugin.Metric, deployment v1beta1.Deployment) ([]plugin.Metric, error) {
	metrics := make([]plugin.Metric, 0)

	for _, mt := range mts {
		ns := mt.Namespace.Strings()

		if len(ns) < minDeploymentNamespaceSize {
			continue
		}
		if !isValidNamespace(ns) {
			continue
		}

		if ns[5] == "metadata" && ns[6] == "generation" {
			metric := createDeploymentMetric(mt, ns, deployment, deployment.Generation)
			metrics = append(metrics, metric)
		} else if ns[5] == "status" && ns[6] == "observedgeneration" {
			metric := createDeploymentMetric(mt, ns, deployment, deployment.Status.ObservedGeneration)
			metrics = append(metrics, metric)
		} else if ns[5] == "status" && ns[6] == "targetedreplicas" {
			metric := createDeploymentMetric(mt, ns, deployment, deployment.Status.Replicas)
			metrics = append(metrics, metric)
		} else if ns[5] == "status" && ns[6] == "availablereplicas" {
			metric := createDeploymentMetric(mt, ns, deployment, deployment.Status.AvailableReplicas)
			metrics = append(metrics, metric)
		} else if ns[5] == "status" && ns[6] == "unavailablereplicas" {
			metric := createDeploymentMetric(mt, ns, deployment, deployment.Status.UnavailableReplicas)
			metrics = append(metrics, metric)
		} else if ns[5] == "status" && ns[6] == "updatedreplicas" {
			metric := createDeploymentMetric(mt, ns, deployment, deployment.Status.UpdatedReplicas)
			metrics = append(metrics, metric)
		} else if ns[5] == "spec" && ns[6] == "desiredreplicas" && deployment.Spec.Replicas != nil {
			metric := createDeploymentMetric(mt, ns, deployment, *deployment.Spec.Replicas)
			metrics = append(metrics, metric)
		} else if ns[5] == "spec" && ns[6] == "paused" {
			metric := createDeploymentMetric(mt, ns, deployment, boolInt(deployment.Spec.Paused))
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
