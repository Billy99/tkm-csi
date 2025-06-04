package contants

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
	TcvBinary = "tcv"
)
