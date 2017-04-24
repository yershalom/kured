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

	"github.com/weaveworks/kured/pkg/alerts"
	"github.com/weaveworks/kured/pkg/daemonsetlock"
	"github.com/weaveworks/kured/pkg/delaytick"
)

var (
	version        = "unreleased"
	period         int
	dsNamespace    string
	dsName         string
	lockAnnotation string
	prometheusURL  string
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
	rootCmd.PersistentFlags().StringVar(&prometheusURL, "prometheus-url", "",
		"Prometheus instance to probe for active alarms")

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

func rebootBlocked() bool {
	if prometheusURL != "" {
		count, err := alerts.PrometheusCountActive(prometheusURL)
		if err != nil {
			log.Warnf("Error probing Prometheus for active alarms: %v", err)
			return true
		}
		if count > 0 {
			log.Warnf("Reboot blocked: %d active alerts", count)
			return true
		}
	}
	return false
}

func holding(lock *daemonsetlock.DaemonSetLock) bool {
	holding, err := lock.Test()
	if err != nil {
		log.Fatalf("Error testing lock: %v", err)
	}
	return holding
}

func acquire(lock *daemonsetlock.DaemonSetLock) bool {
	holding, holder, err := lock.Acquire()
	switch {
	case err != nil:
		log.Fatalf("Error acquiring lock: %v", err)
		return false
	case !holding:
		log.Warnf("Lock already held: %v", holder)
		return false
	default:
		log.Infof("Acquired reboot lock")
		return true
	}
}

func release(lock *daemonsetlock.DaemonSetLock) {
	if err := lock.Release(); err != nil {
		log.Fatalf("Error releasing lock: %v", err)
	}
}

func drain(nodeID string) {
	drainCmd := exec.Command("/usr/local/bin/kubectl", "drain",
		"--ignore-daemonsets", "--delete-local-data", "--force", nodeID)
	if err := drainCmd.Run(); err != nil {
		log.Fatalf("Error invoking drain command: %v", err)
	}
}

func uncordon(nodeID string) {
	uncordonCmd := exec.Command("/usr/local/bin/kubectl", "uncordon", nodeID)
	if err := uncordonCmd.Run(); err != nil {
		log.Fatalf("Error invoking uncordon command: %v", err)
	}
}

func reboot() {
	// Relies on /var/run/dbus/system_bus_socket bind mount to talk to systemd
	rebootCmd := exec.Command("/bin/systemctl", "reboot")
	if err := rebootCmd.Run(); err != nil {
		log.Fatalf("Error invoking reboot command: %v", err)
	}
}

func waitForReboot() {
	for {
		log.Infof("Waiting for reboot")
		time.Sleep(time.Minute)
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

	lock := daemonsetlock.New(client, nodeID, dsNamespace, dsName, lockAnnotation)

	if holding(lock) {
		uncordon(nodeID)
		release(lock)
	}

	source := rand.NewSource(time.Now().UnixNano())
	tick := delaytick.New(source, time.Minute*time.Duration(period))
	for _ = range tick {
		if rebootRequired() && !rebootBlocked() && acquire(lock) {
			drain(nodeID)
			reboot()
			break
		}
	}

	waitForReboot()
}
