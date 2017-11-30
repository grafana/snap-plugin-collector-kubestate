package kubestate

import (
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	v1batch "k8s.io/client-go/pkg/apis/batch/v1"
)

type jobCollector struct {
}

const (
	minJobNamespaceSize = 7
	metricJobPart       = 4
	metricJobTypePart   = 6
)

func (*jobCollector) Collect(mts []plugin.Metric, job v1batch.Job) ([]plugin.Metric, error) {
	metrics := make([]plugin.Metric, 0)
	for _, mt := range mts {
		ns := mt.Namespace.Strings()

		if len(ns) < minJobNamespaceSize {
			continue
		}

		if ns[metricTypeNsPart] == "job" {
			switch ns[metricJobTypePart] {
			case "active":
				metrics = append(metrics, createJobMetric(mt, job, job.Status.Active))
			case "succeeded":
				metrics = append(metrics, createJobMetric(mt, job, job.Status.Succeeded))
			case "failed":
				metrics = append(metrics, createJobMetric(mt, job, job.Status.Failed))
			}
		}
	}

	return metrics, nil
}

func createJobMetric(mt plugin.Metric, job v1batch.Job, value interface{}) plugin.Metric {
	ns := plugin.CopyNamespace(mt.Namespace)
	ns[namespaceNsPart].Value = job.Namespace
	ns[metricJobPart].Value = job.Name
	return plugin.Metric{
		Namespace: ns,
		Data:      value,
		Timestamp: time.Now(),
	}
}
