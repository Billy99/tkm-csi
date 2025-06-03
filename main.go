package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/billy99/tkm-csi/pkgs/driver"
	"github.com/billy99/tkm-csi/pkgs/image"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	// DefaultSocketFilename is the location of the Unix domain socket for this CSI driver
	// for Kubelet to send requests.
	DefaultSocketFilename string = "unix:///var/lib/kubelet/plugins/csi-tkm/csi.sock"

	// DefaultImagePort is the location of port the Image Server will listen on for TKM
	// to send requests.
	DefaultImagePort string = ":50051"

	// DefaultCacheDir is the default root directory to store the expanded the Triton Kernel
	// images.
	DefaultCacheDir = "/run/tkm/caches"

	// DefaultCacheDir is the default root directory to store the expanded the Triton Kernel
	// images.
	ClusterScopedSubDir = "cluster-scoped"

	// TcvBinary is the location on the host of the TCV binary.
	TcvBinary = "/usr/sbin/tcv"
)

var versionInfo = flag.Bool("version", false, "Print the driver version")
var testMode = flag.Bool("test", false, "Flag to indicate in a Test Mode. Creates a stubbed out Kubelet Server")
var noGpu = flag.Bool("nogpu", false, "Flag to indicate in a test scenario and GPU is not present")

func main() {
	// Process input data through environment variables
	nodeName := strings.TrimSpace(os.Getenv("KUBE_NODE_NAME"))
	if nodeName == "" {
		nodeName = "local"
	}
	ns := strings.TrimSpace(os.Getenv("TKM_NAMESPACE"))
	if ns == "" {
		ns = "default"
	}
	logLevel := strings.TrimSpace(os.Getenv("GO_LOG"))
	if logLevel == "" {
		logLevel = "info"
	}
	socketFilename := os.Getenv("CSI_ENDPOINT")
	if socketFilename == "" {
		socketFilename = DefaultSocketFilename
	}
	imagePort := os.Getenv("CSI_IMAGE_SERVER_PORT")
	if imagePort == "" {
		imagePort = DefaultImagePort
	}

	// Parse command line variables
	flag.Parse()

	// Setup logging before anything else so code can log errors.
	log := initializeLogging(logLevel)

	// Process command line variables
	if *versionInfo {
		log.Info("CSI Driver", "Version", driver.Version)
		return
	}

	// Setup CSI Driver, which receives CSI requests from Kubelet
	d, err := driver.NewDriver(log, nodeName, ns, socketFilename, *testMode)
	if err != nil {
		log.Error(err, "Failed to create new Driver object")
		return
	}
	log.Info("Created a new driver:", "driver", d)

	// Setup Image Server, which receives OCI Image management requests from TKM
	s, err := image.NewImageServer(nodeName, ns, imagePort, DefaultCacheDir, ClusterScopedSubDir, TcvBinary, *noGpu)
	if err != nil {
		log.Error(err, "Failed to create new Image Server object")
		return
	}
	log.Info("Created a new Image Server:", "image", s)

	// Create the context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the CSI Driver
	if !(*testMode) {
		go func() {
			log.Info("Running the CSI Driver")

			if err := d.Run(ctx); err != nil {
				log.Error(err, "Driver run failure")
				return
			}
		}()
	}

	// Run the Image Server
	go func() {
		log.Info("Running the Image Server")

		if err := s.Run(ctx); err != nil {
			log.Error(err, "Image Server run failure")
			return
		}
	}()

	// Listen for termination requests
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log = ctrl.Log.WithName("tkm-csi-main")
	log.Info("Running until SIGINT/SIGTERM received")
	sig := <-c
	log.Info("Received signal:", "sig", sig)
	cancel()
}

func initializeLogging(logLevel string) logr.Logger {
	var opts zap.Options

	// Setup logging
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
	return ctrl.Log.WithName("tkm-csi")
}
