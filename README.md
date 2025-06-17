# Kubernetes Cluster Autoscaler Cluster API Monitor

An application to help monitor the Cluster Autoscaler Cluster API provider.

How to build:

`go build`

How to run:

`./cas-capi-monitor --kubeconfig /path/to/clusterapi/kubeconfig --workload-kubeconfig /path/to/node/kubeconfig --workload-namespace default --management-namespace default`

`--kubeconfig`, or in-cluster settings, should point to the management cluster with the Machine resources.

`--workload-kubeconfig`, should point to the workload cluster with the Node and Event resources.

`--workload-namespace`, should be the namespace where the cluster autoscaler will record Events.

`--management-namespace`, should be the namespace where the cluster api Machines are stored.

There are also shorthand options, see `cas-capi-monitor --help` for more details.

if it works, you will see this kind of output:

```
{"level":"info","ts":"2025-06-13T21:08:29Z","logger":"machine-reconciler","msg":"observed Machine with deletion annotation","name":"km-wl-kubemark-md-0-xmz26-xc7dw","value":"yes"}
{"level":"info","ts":"2025-06-13T21:09:54Z","logger":"node-reconciler","msg":"observed Node with ToBeDeleted taint","name":"km-wl-kubemark-md-0-xmz26-xc7dw"}
{"level":"info","ts":"2025-06-13T21:10:07Z","logger":"node-reconciler","msg":"observed Node with ToBeDeleted taint","name":"km-wl-kubemark-md-0-xmz26-xc7dw"}
{"level":"info","ts":"2025-06-16T23:02:29Z","logger":"event-reconciler","msg":"observed cluster-autoscaler generated Event","reason":"ScaleDown","message":"marked the node as toBeDeleted/unschedulable"}
{"level":"info","ts":"2025-06-16T23:02:29Z","logger":"event-reconciler","msg":"observed cluster-autoscaler generated Event","reason":"ScaledUpGroup","message":"Scale-up: setting group MachineDeployment/default/km-wl-kubemark-md-0 size to 5 instead of 1 (max: 5)"}
```

Permissions:

For the kubeconfig to the management cluster, the following is required:
```
rules:
- apiGroups:
  - cluster.x-k8s.io
  resources:
  - machines
  verbs:
  - get
  - list
  - watch
```

For the kubeconfig to the workload cluster, the following is required:
```
rules:
- apiGroups:
  - ""
  resources:
  - events
  - nodes
  verbs:
  - get
  - list
  - watch
```
