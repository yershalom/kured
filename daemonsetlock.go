package kured

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
)

type dsl struct {
	client     *kubernetes.Clientset
	nodeID     string
	annotation string
}

func NewDaemonSetLock(client *kubernetes.Clientset, nodeID string, annotation string) *dsl {
	return &dsl{client, nodeID, annotation}
}

func (dsl *dsl) Acquire() (bool, string, error) {
	for {
		// We should infer our daemonset from kubernetes.io/created-by eventually
		ds, err := dsl.client.ExtensionsV1beta1().DaemonSets("kube-system").Get("kured")
		if err != nil {
			return false, "", err
		}

		holder, exists := ds.ObjectMeta.Annotations[dsl.annotation]
		if exists {
			return holder == dsl.nodeID, holder, nil
		}

		if ds.ObjectMeta.Annotations == nil {
			ds.ObjectMeta.Annotations = make(map[string]string)
		}
		ds.ObjectMeta.Annotations[dsl.annotation] = dsl.nodeID

		_, err = dsl.client.ExtensionsV1beta1().DaemonSets("kube-system").Update(ds)
		if err != nil {
			if se, ok := err.(*errors.StatusError); ok && se.ErrStatus.Reason == unversioned.StatusReasonConflict {
				// Something else updated the resource between us reading and writing - try again soon
				time.Sleep(time.Second)
				continue
			} else {
				return false, "", err
			}
		}
		return true, dsl.nodeID, nil
	}
}

func (dsl *dsl) Test() (bool, error) {
	return false, nil
}

func (dsl *dsl) Release() error {
	return nil
}
