# fs-exporter

> **Note**
>
> The overall structure of the code refers to the [node_exporter], where the cli part refers to [k8s] and the log part refers to [etcd]. 

1. file system exporter for Prometheus
  - GlusterFS
  - ZFS

2. The fs-exporter listens on HTTP port 9097 by default. See the --help output for more options.

# References
- [node_exporter]
- [gluster-prometheus]
- [gluster_exporter]

[node_exporter]:https://github.com/prometheus/node_exporter
[k8s]:https://github.com/kubernetes/kubernetes
[etcd]:https://github.com/etcd-io/etcd
[gluster-prometheus]:https://github.com/gluster/gluster-prometheus
[gluster_exporter]:https://github.com/ofesseler/gluster_exporter

