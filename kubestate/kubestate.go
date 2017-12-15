package kubestate

import (
	"strings"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/intelsdi-x/snap-plugin-utilities/logger"
	. "github.com/intelsdi-x/snap-plugin-utilities/ns"
)

type Kubestate struct {
}

// CollectMetrics collects metrics for testing
func (n *Kubestate) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	LogDebug("request to collect metrics", "metric_count", len(mts))

	incluster, err := mts[0].Config.GetBool("incluster")
	if err != nil {
		LogError("failed to fetch config value incluster.", "error", err)
		incluster = true
	}

	kubeconfigpath := ""
	if !incluster {
		kubeconfigpath, err = mts[0].Config.GetString("kubeconfigpath")
		if err != nil {
			LogError("failed to fetch config value kubeconfigpath.", "error", err)
			return nil, err
		}
	}

	client, err := newClient(incluster, kubeconfigpath)
	if err != nil {
		LogError("failed to create Kubernetes api client.", "error", err)
		return nil, err
	}

	metrics, err := collect(client, mts)
	if err != nil {
		return nil, err
	}

	LogDebug("collecting metrics completed", "metric_count", len(metrics))
	return metrics, nil
}

type MetricScope struct {
	Resource  string
	Node      string
	Namespace string
}

var collect = func(client *Client, mts []plugin.Metric) ([]plugin.Metric, error) {
	metrics := make([]plugin.Metric, 0)

	// group the request metrics by resource, namespace and node
	//
	//       grafanalabs.kubestate.pod.<namespace>.<node>
	//       grafanalabs.kubestate.container.<namespace>.<node>
	//       grafanalabs.kubestate.deployment.<namespace>
	//       grafanalabs.kubestate.job.<namespace>
	//       grafanalabs.kubestate.node.<node>
	//
	groupedMts := make(map[MetricScope][]plugin.Metric)
	for _, mt := range mts {
		ns := mt.Namespace.Strings()
		if len(ns) < 3 {
			continue
		}
		switch ns[2] {
		case "pod":
			namespace := "*"
			node := "*"
			if len(ns) >= 4 {
				namespace = ns[4]
			}
			if len(ns) >= 5 {
				node = ns[5]
			}
			scope := MetricScope{
				Resource:  "pod/container",
				Namespace: namespace,
				Node:      node,
			}
			groupedMts[scope] = append(groupedMts[scope], mt)
		case "container":
			namespace := "*"
			node := "*"
			if len(ns) >= 4 {
				namespace = ns[4]
			}
			if len(ns) >= 5 {
				node = ns[5]
			}
			scope := MetricScope{
				Resource:  "pod/container",
				Namespace: namespace,
				Node:      node,
			}
			groupedMts[scope] = append(groupedMts[scope], mt)
		case "deployment":
			namespace := "*"
			if len(ns) >= 4 {
				namespace = ns[4]
			}
			scope := MetricScope{
				Resource:  "deployment",
				Namespace: namespace,
			}
			groupedMts[scope] = append(groupedMts[scope], mt)
		case "job":
			namespace := "*"
			if len(ns) >= 4 {
				namespace = ns[4]
			}
			scope := MetricScope{
				Resource:  "job",
				Namespace: namespace,
			}
			groupedMts[scope] = append(groupedMts[scope], mt)
		case "node":
			node := "*"
			if len(ns) >= 4 {
				node = ns[4]
			}
			scope := MetricScope{
				Resource: "node",
				Node:     node,
			}
			groupedMts[scope] = append(groupedMts[scope], mt)
		}
	}

	for scope, scopedMts := range groupedMts {
		switch scope.Resource {
		case "pod/container":
			pods, err := client.GetPods(scope.Namespace, scope.Node)
			if err != nil {
				return nil, err
			}

			podCollector := new(podCollector)
			for _, p := range pods.Items {
				podMetrics, _ := podCollector.Collect(scopedMts, p)
				metrics = append(metrics, podMetrics...)
			}
		case "deployment":
			deployments, err := client.GetDeployments(scope.Namespace)
			if err != nil {
				return nil, err
			}

			deploymentCollector := new(deploymentCollector)
			for _, d := range deployments.Items {
				deploymentMetrics, _ := deploymentCollector.Collect(scopedMts, d)
				metrics = append(metrics, deploymentMetrics...)
			}
		case "job":
			jobs, err := client.GetJobs(scope.Namespace)
			if err != nil {
				return nil, err
			}

			jobCollector := new(jobCollector)
			for _, d := range jobs.Items {
				jobMetrics, _ := jobCollector.Collect(scopedMts, d)
				metrics = append(metrics, jobMetrics...)
			}
		case "node":
			nodes, err := client.GetNodes(scope.Node)
			if err != nil {
				return nil, err
			}

			nodeCollector := new(nodeCollector)
			for _, n := range nodes.Items {
				nodeMetrics, _ := nodeCollector.Collect(scopedMts, n)
				metrics = append(metrics, nodeMetrics...)
			}
		}
	}

	return metrics, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}

	return 0
}

func slugify(name string) string {
	return strings.Replace(name, ".", "_", -1)
}

func isValidNamespace(ns []string) bool {
	for i := range ns {
		err := ValidateMetricNamespacePart(ns[i])
		if err != nil {
			// Logger.
			return false
		}
	}
	return true
}

func getPodMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}
	// Pod metrics
	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase", "Pending"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase", "Running"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase", "Succeeded"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase", "Failed"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElements("status", "phase", "Unknown"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElement("status").
			AddStaticElements("condition", "ready"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "pod").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddStaticElement("status").
			AddStaticElements("condition", "scheduled"),
		Version: 1,
	})

	return mts
}

func getJobMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}
	// Job metrics
	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "job").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("job", "job name").
			AddStaticElements("status", "active"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "job").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("job", "job name").
			AddStaticElements("status", "succeeded"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "job").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("job", "job name").
			AddStaticElements("status", "failed"),
		Version: 1,
	})

	return mts
}

func getPodContainerMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "restarts"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "ready"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "waiting"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "running"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("status", "terminated"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("requested", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("requested", "memory", "bytes"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("limits", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "container").
			AddDynamicElement("namespace", "kubernetes namespace").
			AddDynamicElement("node", "node name").
			AddDynamicElement("pod", "pod name").
			AddDynamicElement("container", "container name").
			AddStaticElements("limits", "memory", "bytes"),
		Version: 1,
	})

	return mts
}

func getNodeMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}

	//Node metrics
	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("spec", "unschedulable"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "outofdisk"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "capacity", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "capacity", "memory", "bytes"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "capacity", "pods"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "allocatable", "cpu", "cores"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "allocatable", "memory", "bytes"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "node").
			AddDynamicElement("node", "node name").
			AddStaticElements("status", "allocatable", "pods"),
		Version: 1,
	})

	return mts
}

func getDeploymentMetricTypes() []plugin.Metric {
	mts := []plugin.Metric{}

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("metadata", "generation"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "observedgeneration"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "targetedreplicas"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "availablereplicas"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "unavailablereplicas"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "updatedreplicas"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("spec", "desiredreplicas"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("spec", "paused"),
		Version: 1,
	})

	mts = append(mts, plugin.Metric{
		Namespace: plugin.NewNamespace("grafanalabs", "kubestate", "deployment").
			AddDynamicElement("namespace", "Kubernetes namespace").
			AddDynamicElement("deployment", "deployment name").
			AddStaticElements("status", "deploynotfinished"),
		Version: 1,
	})

	return mts
}

func (n *Kubestate) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	mts := []plugin.Metric{}

	mts = append(mts, getPodMetricTypes()...)
	mts = append(mts, getPodContainerMetricTypes()...)
	mts = append(mts, getNodeMetricTypes()...)
	mts = append(mts, getDeploymentMetricTypes()...)
	mts = append(mts, getJobMetricTypes()...)

	return mts, nil
}

func (f *Kubestate) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewBoolRule([]string{"grafanalabs", "kubestate"}, "incluster", false, plugin.SetDefaultBool(true))
	policy.AddNewStringRule([]string{"grafanalabs", "kubestate"}, "kubeconfigpath", false)
	return *policy, nil
}
