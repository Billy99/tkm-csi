package driver

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"

	// "github.com/tkm/tkmgo"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/go-logr/logr"

	// "github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// Name is the name of the driver
const Name string = "TKM CSI Driver"

// Version is the current version of the driver to set in the User-Agent header
var Version string = "0.0.1"

// DefaultVolumeSizeGB is the default size in Gigabytes of an unspecified volume
const DefaultVolumeSizeGB int = 10

// DefaultSocketFilename is the location of the Unix domain socket for this driver
const DefaultSocketFilename string = "unix:///var/lib/kubelet/plugins/tkm-csi/csi.sock"

// Driver implement the CSI endpoints for Identity, Node and Controller
type Driver struct {
	// TkmClient     tkmgo.Clienter
	// DiskHotPlugger DiskHotPlugger
	controller     bool
	SocketFilename string
	NodeInstanceID string
	Region         string
	Namespace      string
	ClusterID      string
	TestMode       bool
	grpcServer     *grpc.Server
	log            logr.Logger
}

// NewDriver returns a CSI driver that implements gRPC endpoints for CSI
func NewDriver(log logr.Logger, apiURL, apiKey, region, namespace, clusterID string) (*Driver, error) {
	// var client *tkmgo.Client
	var err error

	/*
		if apiKey != "" {
			client, err = tkmgo.NewClientWithURL(apiKey, apiURL, region)
			if err != nil {
				return nil, err
			}
		}

		userAgent := &tkmgo.Component{
			ID:      clusterID,
			Name:    "tkm-csi",
			Version: Version,
		}

		client.SetUserAgent(userAgent)
	*/

	socketFilename := os.Getenv("CSI_ENDPOINT")
	if socketFilename == "" {
		socketFilename = DefaultSocketFilename
	}

	log.Info("Created a new driver - api_url %s  region %v  namespace %s  cluster_id %s  socketFilename %s",
		apiURL, region, namespace, clusterID, socketFilename)

	return &Driver{
		// TkmClient:     client,
		Region:    region,
		Namespace: namespace,
		ClusterID: clusterID,
		// DiskHotPlugger: &RealDiskHotPlugger{},
		controller:     (apiKey != ""),
		SocketFilename: socketFilename,
		grpcServer:     &grpc.Server{},
		log:            log,
	}, nil
}

// NewTestDriver returns a new TKM CSI driver specifically setup to call a fake TKM API
/*
func NewTestDriver(fc *tkmgo.FakeClient) (*Driver, error) {
	d, err := NewDriver("https://tkm-api.example.com", "NO_API_KEY_NEEDED", "TEST1", "default", "12345678")
	d.SocketFilename = "unix:///tmp/tkm-csi.sock"
	if fc != nil {
		d.TkmClient = fc
	} else {
		d.TkmClient, _ = tkmgo.NewFakeClient()
	}

	d.DiskHotPlugger = &FakeDiskHotPlugger{}
	d.TestMode = true // Just stops so much logging out of failures, as they are often expected during the tests

	zerolog.SetGlobalLevel(zerolog.PanicLevel)

	return d, err
}
*/

// Run the driver's gRPC server
func (d *Driver) Run(ctx context.Context) error {
	d.log.Info("Parsing the socket filename to make a gRPC server - socketFilename %s", d.SocketFilename)
	urlParts, _ := url.Parse(d.SocketFilename)
	d.log.Info("Parsed socket filename")

	grpcAddress := path.Join(urlParts.Host, filepath.FromSlash(urlParts.Path))
	if urlParts.Host == "" {
		grpcAddress = filepath.FromSlash(urlParts.Path)
	}
	d.log.Info("Generated gRPC address")

	// remove any existing left-over socket
	if err := os.Remove(grpcAddress); err != nil && !os.IsNotExist(err) {
		d.log.Error("failed to remove unix domain socket file %s, error: %s", grpcAddress, err)
		return fmt.Errorf("failed to remove unix domain socket file %s, error: %s", grpcAddress, err)
	}
	d.log.Info("Removed any exsting old socket")

	grpcListener, err := net.Listen(urlParts.Scheme, grpcAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	d.log.Info("Created gRPC listener")

	// log gRPC response errors for better observability
	errHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			d.log.Info("method failed - method %v", info.FullMethod)
		}
		return resp, err
	}

	if d.TestMode {
		d.grpcServer = grpc.NewServer()
	} else {
		d.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(errHandler))
	}
	d.log.Info("Created new RPC server")

	csi.RegisterIdentityServer(d.grpcServer, d)
	d.log.Info("Registered Identity server")
	csi.RegisterControllerServer(d.grpcServer, d)
	d.log.Info("Registered Controller server")
	csi.RegisterNodeServer(d.grpcServer, d)
	d.log.Info("Registered Node server")

	d.log.Info("Starting gRPC server - grpc_address %s", grpcAddress)

	var eg errgroup.Group

	eg.Go(func() error {
		go func() {
			<-ctx.Done()
			d.log.Info("Stopping gRPC because the context was cancelled")
			d.grpcServer.GracefulStop()
		}()
		d.log.Info("Awaiting gRPC requests")
		return d.grpcServer.Serve(grpcListener)
	})

	d.log.Info("Running gRPC server, waiting for a signal to quit the process... - grpc_address %s", grpcAddress)

	return eg.Wait()
}
