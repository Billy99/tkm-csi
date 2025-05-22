package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/billy99/tkm-csi/pkgs/driver"
	"go.uber.org/zap/zapcore"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var versionInfo = flag.Bool("version", false, "Print the driver version")

func main() {
	var opts zap.Options

	nodeName := strings.TrimSpace(os.Getenv("KUBE_NODE_NAME"))
	ns := strings.TrimSpace(os.Getenv("TKM_NAMESPACE"))
	logLevel := strings.TrimSpace(os.Getenv("GO_LOG"))

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
		log.Info("CSI Driver", "Version", driver.Version)
		return
	}

	d, err := driver.NewDriver(log, nodeName, ns)
	if err != nil {
		log.Error(err, "Failed to create new Driver object")
		return
	}

	log.Info("Created a new driver:", "driver", d)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log = ctrl.Log.WithName("tkm-csi-main")
		log.Info("Running until SIGINT/SIGTERM received")
		sig := <-c
		log.Info("Received signal:", "sig", sig)
		cancel()
	}()

	log.Info("Running the driver")

	if err := d.Run(ctx); err != nil {
		log.Error(err, "Run failure")
		return
	}
}
