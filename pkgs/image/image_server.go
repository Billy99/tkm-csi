package image

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/billy99/tkm-csi/pkgs/constants"
	pb "github.com/billy99/tkm-csi/proto"
)

type ImageServer struct {
	grpcServer *grpc.Server
	log        logr.Logger

	NodeName  string
	Namespace string
	ImagePort string
	TestMode  bool
	noGpu     bool

	pb.UnimplementedTkmCsiServiceServer
}

func (s *ImageServer) LoadKernelImage(ctx context.Context, req *pb.LoadKernelImageRequest) (*pb.LoadKernelImageResponse, error) {
	var namespace string

	if req.Namespace != nil {
		namespace = *req.Namespace
	} else {
		namespace = constants.ClusterScopedSubDir
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
	nodeName, namespace, imagePort string,
	noGpu bool) (*ImageServer, error) {

	if !fileExists(constants.TcvBinary) {
		return nil, fmt.Errorf("TCV must be installed", "location", constants.TcvBinary)
	}

	return &ImageServer{
		grpcServer: grpc.NewServer(),
		NodeName:   nodeName,
		Namespace:  namespace,
		ImagePort:  imagePort,
		noGpu:      noGpu,
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
