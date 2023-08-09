# custom-autoscaler-operator

The CustomScaler controller provides custom scaling for Kubernetes Deployments based on metric thresholds. e.g Number of queued jobs, number of messages in a queue, etc.

## Description

The custom-autoscaler-operator is a specialized Kubernetes operator designed to offer flexible and user-defined autoscaling capabilities. Unlike traditional Kubernetes autoscalers that rely mainly on CPU and memory metrics, the CustomScaler controller within this operator allows for scaling decisions based on a wider variety of metrics. This could include application-specific indicators such as the number of queued jobs in a processing queue, the volume of messages in a message broker, or any other custom metric that can provide insights into the load or demand on your applications.



## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

```sh
make deploy IMG=rscheele3214/custom-autoscaler-operator:latest
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing

// TODO

### How it works

Upon detecting that a custom metric has crossed a threshold, the CustomScaler controller will adjust the replica count of the target deployment either up or down, depending on the metric value. Once a scaling action is taken, the controller enters a cooldown period during which no further scaling actions will be performed.


### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## TODO

### fetchMetricValue Function:

- This function is meant to fetch the metric value from a given metric source, potentially from monitoring systems like Prometheus.
- The actual logic to fetch this metric is still pending.

### isMetricSourceValid Function:

- This function serves to validate the given metric source, perhaps by making a test connection or through other validation mechanisms.
- The actual validation logic is still pending.


## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

