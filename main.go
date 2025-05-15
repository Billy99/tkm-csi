package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/billy99/tkm-csi/pkg/driver"
	"go.uber.org/zap/zapcore"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var versionInfo = flag.Bool("version", false, "Print the driver version")

func main() {
	var opts zap.Options

	// Get the Log level for bpfman deployment where this pod is running
	logLevel := os.Getenv("GO_LOG")
	switch logLevel {
	case "info":
		opts = zap.Options{
			Development: false,
		}
	case "debug":
		opts = zap.Options{
			Development: true,
		}
	case "trace":
		opts = zap.Options{
			Development: true,
			Level:       zapcore.Level(-2),
		}
	default:
		opts = zap.Options{
			Development: false,
		}
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	log := ctrl.Log.WithName("tkm-csi")

	flag.Parse()
	if *versionInfo {
		log.Info("CSI Driver Version %s", driver.Version)
		return
	}

	apiURL := strings.TrimSpace(os.Getenv("TKM"))
	apiKey := strings.TrimSpace(os.Getenv("TKM_API_KEY"))
	region := strings.TrimSpace(os.Getenv("TKM_REGION"))
	ns := strings.TrimSpace(os.Getenv("TKM_NAMESPACE"))
	clusterID := strings.TrimSpace(os.Getenv("TKM_CLUSTER_ID"))

	d, err := driver.NewDriver(log, apiURL, apiKey, region, ns, clusterID)
	if err != nil {
		log.Error(err, "Failed to create new Driver object")
		return
	}

	log.Info("Created a new driver: %d", d)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Info("Running until SIGINT/SIGTERM received")
		sig := <-c
		log.Info("Received signal: %v", sig)
		cancel()
	}()

	log.Info("Running the driver")

	if err := d.Run(ctx); err != nil {
		log.Error(err, "Run failure")
		return
	}
}
