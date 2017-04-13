package main

import (
	"flag"
	"fmt"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig = flag.String("kubeconfig", "/home/awh/.kube/config", "path to kubeconfig file")
)

func main() {
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	s, err := clientSet.CoreV1().Services("default").Get("kubernetes")
	if err != nil {
		log.Fatal(err)
	}

	if s.ObjectMeta.Annotations == nil {
		s.ObjectMeta.Annotations = make(map[string]string)
	}
	s.ObjectMeta.ResourceVersion = "11000"
	s.ObjectMeta.Annotations["lock"] = "foo"

	ns, err := clientSet.CoreV1().Services("default").Update(s)
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok {
			if statusErr.ErrStatus.Reason == unversioned.StatusReasonConflict {
				fmt.Printf("conflict detected: %#v", ns)
			}
		}
	}
}
