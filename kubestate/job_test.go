package kubestate

import (
	"testing"

	"k8s.io/client-go/pkg/api/v1"
	v1batch "k8s.io/client-go/pkg/apis/batch/v1"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	mockJobs = []v1batch.Job{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      "job1",
				Namespace: "default",
			},
			Spec: v1batch.JobSpec{},
			Status: v1batch.JobStatus{
				Active:    1,
				Failed:    2,
				Succeeded: 3,
			},
		},
	}
	jobCases = []struct {
		job      v1batch.Job
		metrics  []plugin.Metric
		expected []string
	}{
		{
			job:     mockJobs[0],
			metrics: getJobMetricTypes(),
			expected: []string{
				"grafanalabs.kubestate.job.default.job1.status.active 1",
				"grafanalabs.kubestate.job.default.job1.status.succeeded 3",
				"grafanalabs.kubestate.job.default.job1.status.failed 2",
			},
		},
		{
			job:      mockJobs[0],
			metrics:  malformedMetricTypes,
			expected: []string{},
		},
	}
)

func TestJobCollector(t *testing.T) {
	Convey("When collecting metrics for jobs", t, func() {
		for _, c := range jobCases {
			metrics, err := new(jobCollector).Collect(c.metrics, c.job)

			So(err, ShouldBeNil)

			So(metrics, ShouldNotBeNil)
			So(len(metrics), ShouldEqual, len(c.expected))

			for i, metric := range metrics {
				So(format(&metric), ShouldEqual, c.expected[i])
			}
		}
	})
}
