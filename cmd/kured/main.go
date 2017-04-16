package main

import (
	"math/rand"
	"os"
	"os/exec"
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

func rebootRequired() bool {
	_, err := os.Stat("/var/run/reboot-required")
	switch {
	case err == nil:
		return true
	case os.IsNotExist(err):
		return false
	default:
		log.Fatalf("Unable to determine if reboot required: %v", err)
		return false // unreachable; prevents compilation error
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
		uncordonCmd := exec.Command("/usr/local/bin/kubectl", "uncordon", nodeID)
		if err := uncordonCmd.Run(); err != nil {
			log.Fatalf("Error invoking uncordon command: %v", err)
		}

		if err := lock.Release(); err != nil {
			log.Fatal(err)
		}
	}

	source := rand.NewSource(time.Now().UnixNano())
	ticker := kured.NewDelayTick(source, time.Minute*time.Duration(period))
	for _ = range ticker {
		if !rebootRequired() {
			continue
		}

		holding, holder, err := lock.Acquire()
		switch {
		case err != nil:
			log.Fatalf("Error during lock acquisition: %v", err)
		case !holding:
			log.Warnf("Lock already held: %v", holder)
			continue
		}

		drainCmd := exec.Command("/usr/local/bin/kubectl", "drain",
			"--ignore-daemonsets", "--delete-local-data", "--force", nodeID)
		if err := drainCmd.Run(); err != nil {
			log.Fatalf("Error invoking drain command: %v", err)
		}

		// Relies on /var/run/dbus/system_bus_socket bind mount
		rebootCmd := exec.Command("/bin/systemctl", "reboot")
		if err := rebootCmd.Run(); err != nil {
			log.Fatalf("Error invoking reboot command: %v", err)
		}

		break
	}

	// Wait indefinitely for reboot to occur
	for {
		log.Infof("Waiting for reboot")
		time.Sleep(time.Minute)
	}
}
