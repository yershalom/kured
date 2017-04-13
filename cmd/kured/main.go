package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	period         int
	lockAnnotation string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "anr",
		Short: "Kubernetes Node Reboot Daemon",
		Run:   root}

	rootCmd.PersistentFlags().IntVar(&period, "period", 60, "reboot check period in minutes")
	rootCmd.PersistentFlags().StringVar(&lockAnnotation, "lock-annotation", "works.weave/kured-node-lock",
		"annotation in which to record locking node")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func root(cmd *cobra.Command, args []string) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	nodeID := "foo"
	lock := NewClusterLock(client, nodeID, lockAnnotation)

	holdingLock, err := lock.Test()
	if err != nil {
		log.Fatal(err)
	}

	if holdingLock {
		if err := lock.Release(); err != nil {
			log.Fatal(err)
		}
	}

	ticker := time.NewTicker(time.Minute * time.Duration(period))

	for _ = range ticker.C {
		holdingLock, holder, err := lock.Acquire()
		if err != nil {
			log.Errorf("Unable to acquire lock: %v", err)
		}
		if !holdingLock {
			log.Infof("Lock already held: %v", holder)
			continue
		}
	}
}
