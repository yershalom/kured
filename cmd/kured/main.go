package main

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/weaveworks/kured"
)

var (
	period         int
	dsNamespace    string
	dsName         string
	lockAnnotation string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "kured",
		Short: "Kubernetes Reboot Daemon",
		Run:   root}

	rootCmd.PersistentFlags().IntVar(&period, "period", 60,
		"reboot check period in minutes")
	rootCmd.PersistentFlags().StringVar(&dsNamespace, "ds-name", "kube-system",
		"namespace containing daemonset on which to place lock")
	rootCmd.PersistentFlags().StringVar(&dsName, "ds-namespace", "kured",
		"name of daemonset on which to place lock")
	rootCmd.PersistentFlags().StringVar(&lockAnnotation, "lock-annotation", "works.weave/kured-node-lock",
		"annotation in which to record locking node")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func root(cmd *cobra.Command, args []string) {
	nodeID := os.Getenv("KURED_NODE_ID")
	if nodeID == "" {
		log.Fatal("KURED_NODE_ID environment variable required")
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	lock := kured.NewDaemonSetLock(client, nodeID, dsNamespace, dsName, lockAnnotation)

	holding, err := lock.Test()
	if err != nil {
		log.Fatal(err)
	}

	if holding {
		// TBD: Uncordon

		if err := lock.Release(); err != nil {
			log.Fatal(err)
		}
	}

	ticker := time.NewTicker(time.Minute * time.Duration(period))
	for _ = range ticker.C {
		holding, holder, err := lock.Acquire()
		if err != nil {
			log.Fatalf("Unable to acquire lock: %v", err)
		}
		if !holding {
			log.Infof("Lock already held: %v", holder)
			continue
		}

		// TBD: Drain & reboot
	}
}
