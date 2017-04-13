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
	namespace  string
	name       string
	annotation string
}

func NewDaemonSetLock(client *kubernetes.Clientset, nodeID, namespace, name, annotation string) *dsl {
	return &dsl{client, nodeID, namespace, name, annotation}
}

func (dsl *dsl) Acquire() (bool, string, error) {
	for {
		ds, err := dsl.client.ExtensionsV1beta1().DaemonSets(dsl.namespace).Get(dsl.name)
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

		_, err = dsl.client.ExtensionsV1beta1().DaemonSets(dsl.namespace).Update(ds)
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
