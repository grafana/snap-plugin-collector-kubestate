# snap collector plugin - kube state

This plugin collects metrics from Kubernetes about the state of pods, nodes and deployments.

It's used in the [snap framework](http://github.com:intelsdi-x/snap).

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Contributing](#contributing)
4. [License](#license-and-authors)
5. [Acknowledgements](#acknowledgements)

## Getting Started

### System Requirements

* [golang 1.5+](https://golang.org/dl/) (needed only for building)

### Operating systems

All OSs currently supported by snap:
* Linux/amd64
* Darwin/amd64

### Installation

This plugin monitors Kubernetes so you need a Kubernetes cluster to monitor. A quick way to get a local test cluster installed is to use [Minikube](https://github.com/kubernetes/minikube).

This plugin is included in the [Raintank Docker image for Kubernetes](https://github.com/raintank/snap_k8s) if you want to see an example of how to easily deploy it as pod to Kubernetes.

This plugin and the above Docker image are used to collect metrics for the [Grafana Kubernetes App](https://github.com/raintank/kubernetes-app).

#### Download kubestate plugin binary:

You can get the pre-built binaries at [GitHub Releases](https://github.com/raintank/snap-plugin-collector-kubestate/releases) page.

#### To build the plugin binary:

Fork https://github.com/raintank/snap-plugin-collector-kubestate
Clone repo into `$GOPATH/src/github.com/raintank/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-kubestate.git
```

Build the plugin by running make within the cloned repo:
```
$ ./build.sh
```
This builds the plugin binary in `/build/`

This plugin uses govendor to manage dependencies. If you want to add a dependency, then:

1. Install govendor with: `go get -u github.com/kardianos/govendor`
2. `govendor fetch <dependency path>` e.g. `govendor fetch k8s.io/client-go/tools/clientcmd/...@v2.0.0-alpha.0` The `...` means install all sub dependencies.
3. `govendor install` to update the vendor.json file.
4. Check in the new dependency that will be in the vendor directory.

### Configuration and Usage

* Set up the [snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)
* Ensure `$SNAP_PATH` is exported
`export SNAP_PATH=$GOPATH/src/github.com/intelsdi-x/snap/build`
* If running the task outside of a kubernetes cluster rather than in a pod, then the following two config variables must be set:
  - `incluster` expects a boolean, default is true.
  - `kubeconfigpath` expects a string path to the Kubernetes config file, default is empty string.

  Example of how to configure it in a json task manifest:
  ```json
  "workflow": {
    "collect": {
      "metrics": {
        "/grafanalabs/kubestate/*":{}
      },
      "config": {
        "/grafanalabs/kubestate": {
          "incluster": false,
          "kubeconfigpath": "/home/user/.kube/config"
        }
      },
  ```

## Documentation

There are a number of other resources you can review to learn to use this plugin:

- [The Kubernetes API spec](http://kubernetes.io/docs/api-reference/v1/definitions/). All the metrics are fetched via the API.

### Collected Metrics

This plugin has the ability to gather the following metrics:

#### Pods

Namespace | Description (optional)
----------|-----------------------
/grafanalabs/kubestate/pod/[NAMESPACE]/[POD]/status/condition/ready | specifies if the pod is ready to serve requests
/grafanalabs/kubestate/pod/[NAMESPACE]/[POD]/status/condition/scheduled | status of the scheduling process for the pod
/grafanalabs/kubestate/pod/[NAMESPACE]/[POD]/status/phase/[PHASE]/value | Phase can be Pending, Running, Succeeded, Failed, Unknown
/grafanalabs/kubestate/container/[NAMESPACE]/[NODE]/[POD]/[CONTAINER]/limits/cpu/cores | The limit on cpu cores to be used by a container.
/grafanalabs/kubestate/container/[NAMESPACE]/[NODE]/[POD]/[CONTAINER]/limits/memory/bytes | The limit on memory to be used by a container in bytes.
/grafanalabs/kubestate/container/[NAMESPACE]/[NODE]/[POD]/[CONTAINER]/requested/cpu/cores | The number of requested cpu cores by a container.
/grafanalabs/kubestate/container/[NAMESPACE]/[NODE]/[POD]/[CONTAINER]/requested/memory/bytes | The number of requested memory bytes by a container.
/grafanalabs/kubestate/container/[NAMESPACE]/[POD]/[CONTAINER]/status/ready | specifies whether the container has passed its readiness probe
/grafanalabs/kubestate/container/[NAMESPACE]/[POD]/[CONTAINER]/status/restarts | number of times the container has been restarted
/grafanalabs/kubestate/container/[NAMESPACE]/[POD]/[CONTAINER]/status/running | value 1 if container is running else value 0
/grafanalabs/kubestate/container/[NAMESPACE]/[POD]/[CONTAINER]/status/terminated | value 1 if container is terminated else value 0
/grafanalabs/kubestate/container/[NAMESPACE]/[POD]/[CONTAINER]/status/waiting | value 1 if container is waiting else value 0

#### Nodes

Namespace | Description (optional)
----------|-----------------------
/grafanalabs/kubestate/node/[NODE]/spec/unschedulable | Whether a node can schedule new pods.
/grafanalabs/kubestate/node/[NODE]/status/allocatable/cpu/cores | The CPU resources of a node that are available for scheduling.
/grafanalabs/kubestate/node/[NODE]/status/allocatable/memory/bytes | The memory resources of a node that are available for scheduling.
/grafanalabs/kubestate/node/[NODE]/status/allocatable/pods | The pod resources of a node that are available for scheduling.
/grafanalabs/kubestate/node/[NODE]/status/capacity/cpu/cores | The total CPU resources of the node.
/grafanalabs/kubestate/node/[NODE]/status/capacity/memory/bytes | The total memory resources of the node.
/grafanalabs/kubestate/node/[NODE]/status/capacity/pods | The total pod resources of the node.

#### Deployments

Namespace | Description (optional)
----------|-----------------------
/grafanalabs/kubestate/deployment/[NAMESPACE]/[DEPLOYMENT]/metadata/generation | The desired generation sequence number for deployment. If a deployment succeeds should be the same as the observed generation.
/grafanalabs/kubestate/deployment/[NAMESPACE]/[DEPLOYMENT]/status/observedgeneration | The generation sequence number after deployment.
/grafanalabs/kubestate/deployment/[NAMESPACE]/[DEPLOYMENT]/status/targetedreplicas | Total number of non-terminated pods targeted by this deployment (their labels match the selector).
/grafanalabs/kubestate/deployment/[NAMESPACE]/[DEPLOYMENT]/status/availablereplicas | Total number of available pods (ready for at least minReadySeconds) targeted by this deployment.
/grafanalabs/kubestate/deployment/[NAMESPACE]/[DEPLOYMENT]/status/unavailablereplicas | Total number of unavailable pods targeted by this deployment.
/grafanalabs/kubestate/deployment/[NAMESPACE]/[DEPLOYMENT]/spec/desiredreplicas | Number of desired pods.

### Examples

### Roadmap

1. Disk capacity for the cluster

## Contributing

See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This plugin is released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [@daniellee](https://github.com/daniellee/)