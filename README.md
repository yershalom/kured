WIP Kubernetes Reboot Daemon

Outstanding work:

* Testing

# kubectl annoyance

$ kubectl-1.4.8 --kubeconfig=ansible/dev/kubeconfig drain --delete-local-data --force --ignore-daemonsets ip-172-20-1-135.ec2.internal
Error from server: Operation cannot be fulfilled on nodes "ip-172-20-1-135.ec2.internal": the object has been modified; please apply your changes to the latest version and try again
2017-04-27 12:25:06 [2636] awh@terraqueous:~/workspace/service-conf [master] 
$ echo $?
1
2017-04-27 12:25:17 [2637] awh@terraqueous:~/workspace/service-conf [master] 
$ kubectl-1.4.8 --kubeconfig=ansible/dev/kubeconfig drain --delete-local-data --force --ignore-daemonsets ip-172-20-1-135.ec2.internal
node "ip-172-20-1-135.ec2.internal" cordoned
WARNING: Deleting pods not managed by ReplicationController, ReplicaSet, Job, or DaemonSet: kube-proxy-ip-172-20-1-135.ec2.internal; Ignoring DaemonSet-managed pods: kured-qc4md, scope-probe-master-ddfw7, fluentd-loggly-vrblq, prom-node-exporter-b7woj, reboot-required-i8xwz; Deleting pods with local storage: terradiff-792849672-69ukn
pod "consul-3919604694-9j0iy" deleted
pod "ingester-2514985883-ten3a" deleted
pod "memcached-2323247251-txjqd" deleted
pod "users-db-exporter-1123349103-kbl92" deleted
pod "demo-3909231545-pgod1" deleted
pod "launch-generator-3829894089-vu8du" deleted
pod "fluxsvc-3308921893-xxnqp" deleted
pod "memcached-1940254774-3b0av" deleted
pod "kube-dns-2757152549-iwuy2" deleted
pod "scope-app-1386610216-qstg6" deleted
pod "compare-revisions-3366969505-bdl37" deleted
pod "metrics-873123495-nffld" deleted
pod "terradiff-792849672-69ukn" deleted
pod "collection-1189917692-flv0s" deleted
pod "query-2969484575-1ipnl" deleted
pod "demo-3909231545-pgod1" deleted
pod "launch-generator-3829894089-vu8du" deleted
pod "memcached-1940254774-3b0av" deleted
pod "memcached-2323247251-txjqd" deleted
pod "users-db-exporter-1123349103-kbl92" deleted
pod "scope-app-1386610216-qstg6" deleted
pod "metrics-873123495-nffld" deleted
pod "query-2969484575-1ipnl" deleted
pod "terradiff-792849672-69ukn" deleted
pod "collection-1189917692-flv0s" deleted
pod "consul-3919604694-9j0iy" deleted
pod "fluxsvc-3308921893-xxnqp" deleted
pod "kube-dns-2757152549-iwuy2" deleted
pod "compare-revisions-3366969505-bdl37" deleted
pod "ingester-2514985883-ten3a" deleted
node "ip-172-20-1-135.ec2.internal" drained
