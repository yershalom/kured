package main

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
)

type clusterLock struct {
	client     *kubernetes.Clientset
	nodeID     string
	annotation string
}

func NewClusterLock(client *kubernetes.Clientset, nodeID string, annotation string) *clusterLock {
	return &clusterLock{client, nodeID, annotation}
}

func (cl *clusterLock) Acquire() (bool, string, error) {
	for {
		// We should infer our daemonset from kubernetes.io/created-by eventually
		ds, err := cl.client.ExtensionsV1beta1().DaemonSets("kube-system").Get("kured")
		if err != nil {
			return false, "", err
		}

		holder, exists := ds.ObjectMeta.Annotations[cl.annotation]
		if exists {
			return holder == cl.nodeID, holder, nil
		}

		if ds.ObjectMeta.Annotations == nil {
			ds.ObjectMeta.Annotations = make(map[string]string)
		}
		ds.ObjectMeta.Annotations[cl.annotation] = cl.nodeID

		_, err = cl.client.ExtensionsV1beta1().DaemonSets("kube-system").Update(ds)
		if err != nil {
			if se, ok := err.(*errors.StatusError); ok && se.ErrStatus.Reason == unversioned.StatusReasonConflict {
				// Something else updated the resource between us reading and writing - try again soon
				time.Sleep(time.Second)
				continue
			} else {
				return false, "", err
			}
		}
		return true, cl.nodeID, nil
	}
}

func (cl *clusterLock) Test() (bool, error) {
	return false, nil
}

func (cl *clusterLock) Release() error {
	return nil
}
