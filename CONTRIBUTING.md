# Contributing guidelines

## Filing issues

File issues using the standard Github issue tracker for the repo.

# Building the code

1. * [golang 1.5+](https://golang.org/dl/)
2. Run ./build.sh
3. To run the tests: `go test -v ./kubestate`

You will need a Kubernetes cluster to fetch metrics from. A quick way to get going is to use [Minikube](https://github.com/kubernetes/minikube).

You can use client_test.go as way to test if you can successfully contact the cluster. Remove `Skip` from `SkipConvey` to enable the tests and then run `go test -v ./kubestate`

To test the snap task locally with the snap server (snapd or snapteld):

1. Load the plugin:
  `snapctl plugin load build/snap-plugin-collector-kubestate` or `snaptel plugin load build/snap-plugin-collector-kubestate` if you have latest version of Snap.
2. If you want to publish the results to a file, you need to download the [file publisher plugin](https://s3-us-west-2.amazonaws.com/snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file) and load that too:
  `snapctl plugin load snap-plugin-publisher-file_linux_x86_64`
2. Create a task using the [example task](examples/task.json) after editing the config to point to your Kubernetes config file:
  `snapctl task create -t examples/task.json`
3. `snapctl task list` to check if it is running and then `snapctl task watch <task id>` to see the value being fetched.