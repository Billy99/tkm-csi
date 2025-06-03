package image

import (
	"context"
	"fmt"
	"net"
	"os"

	pb "github.com/billy99/tkm-csi/proto"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ImageServer struct {
	grpcServer *grpc.Server
	log        logr.Logger

	NodeName            string
	Namespace           string
	ImagePort           string
	cacheDir            string
	clusterScopedSubDir string
	tcvBinary           string
	TestMode            bool
	noGpu               bool

	pb.UnimplementedTkmCsiServiceServer
}

func (s *ImageServer) LoadKernelImage(ctx context.Context, req *pb.LoadKernelImageRequest) (*pb.LoadKernelImageResponse, error) {
	var namespace string

	if req.Namespace != nil {
		namespace = *req.Namespace
	} else {
		namespace = s.clusterScopedSubDir
	}

	s.log.Info("Received Load Kernel Image Request",
		"CRD Name", req.Name,
		"Image", req.Image.Url,
		"Namespace", namespace)

	err := s.ExtractImage(ctx, req.Image.Url, namespace, req.Name)

	if err != nil {
		return &pb.LoadKernelImageResponse{Message: "Load Image Request Failed"}, err
	}
	return &pb.LoadKernelImageResponse{Message: "Load Image Request Succeeded"}, nil
}

func (s *ImageServer) UnloadKernelImage(ctx context.Context, req *pb.UnloadKernelImageRequest) (*pb.UnloadKernelImageResponse, error) {
	s.log.Error(fmt.Errorf("failed to"), "Received Unload Kernel Image Request",
		"CRD Name", req.Name,
		"Namespace", req.Namespace)
	return &pb.UnloadKernelImageResponse{Message: "Unload Image Request Received"}, nil
}

// NewImageServer returns an ImageServer instance that implements gRPC endpoints
// for TKM to manage Triton Kernel Caches that are loaded via OCI Images.
func NewImageServer(
	nodeName, namespace, imagePort, cacheDir, clusterScopedSubDir, tcvBinary string,
	noGpu bool) (*ImageServer, error) {

	if !fileExists(tcvBinary) {
		return nil, fmt.Errorf("TCV must be installed", "location", tcvBinary)
	}

	return &ImageServer{
		grpcServer:          grpc.NewServer(),
		NodeName:            nodeName,
		Namespace:           namespace,
		ImagePort:           imagePort,
		cacheDir:            cacheDir,
		clusterScopedSubDir: clusterScopedSubDir,
		tcvBinary:           tcvBinary,
		noGpu:               noGpu,
	}, nil
}

func (s *ImageServer) Run(ctx context.Context) error {
	s.log = ctrl.Log.WithName("tkm-image-server")

	s.initializeFilesystem()

	lis, err := net.Listen("tcp", s.ImagePort)
	if err != nil {
		s.log.Error(err, "failed to listen", "ImagePort", s.ImagePort)
		return err
	}

	pb.RegisterTkmCsiServiceServer(s.grpcServer, s)
	s.log.Info("gRPC server", "listening at", lis.Addr())
	if err := s.grpcServer.Serve(lis); err != nil {
		s.log.Error(err, "failed to serve")
		return err
	}

	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
