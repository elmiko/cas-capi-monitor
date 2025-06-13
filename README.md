# Kubernetes Cluster Autoscaler Cluster API Monitor

An application to help monitor the Cluster Autoscaler Cluster API provider.

How to build:

`go build`

How to run:

`./cas-capi-monitor --kubeconfig /path/to/clusterapi/kubeconfig --nk /path/to/node/kubeconfig`

if it works, you will see this kind of output:

```
{"level":"info","ts":"2025-06-13T21:08:29Z","logger":"machine-reconciler","msg":"observed Machine with deletion annotation","name":"km-wl-kubemark-md-0-xmz26-xc7dw","value":"yes"}
{"level":"info","ts":"2025-06-13T21:09:54Z","logger":"node-reconciler","msg":"observed Node with ToBeDeleted taint","name":"km-wl-kubemark-md-0-xmz26-xc7dw"}
{"level":"info","ts":"2025-06-13T21:10:07Z","logger":"node-reconciler","msg":"observed Node with ToBeDeleted taint","name":"km-wl-kubemark-md-0-xmz26-xc7dw"}
```
