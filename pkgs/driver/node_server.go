package driver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/billy99/tkm-csi/pkgs/utils"
)

// MaxVolumesPerNode is the maximum number of volumes a single node may host
const MaxVolumesPerNode int64 = 1024
const TritonKernelCacheIndex string = "csi.tkm.io/tritonKernelCache"
const TritonKernelCacheNamespaceIndex string = "csi.tkm.io/namespace"

// NodeStageVolume is called after the volume is attached to the instance, so it can be partitioned,
// formatted and mounted to a staging path
func (d *Driver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest,
) (*csi.NodeStageVolumeResponse, error) {
	d.log.Info("Request: NodeStageVolume", "volume_id", req.VolumeId, "staging_target_path", req.StagingTargetPath)

	if req.VolumeId == "" {
		d.log.Error(fmt.Errorf("must provide a VolumeId to NodeStageVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to NodeStageVolume")
	}
	if req.StagingTargetPath == "" {
		d.log.Error(fmt.Errorf("must provide a StagingTargetPath to NodeStageVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a StagingTargetPath to NodeStageVolume")
	}
	if req.VolumeCapability == nil {
		d.log.Error(fmt.Errorf("must provide a VolumeCapability to NodeStageVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeCapability to NodeStageVolume")
	}

	d.log.V(1).Info("Formatting and mounting volume (staging)", "volume_id", req.VolumeId)

	/*
		// Find the disk attachment location
		attachedDiskPath := d.DiskHotPlugger.PathForVolume(req.VolumeId)
		if attachedDiskPath == "" {
			d.log.Error(fmt.Errorf("path to volume (/dev/disk/by-id/VOLUME_ID) not found"),
				"Invalid Input", "volume_id", req.VolumeId)
			return nil, status.Errorf(codes.NotFound, "path to volume (/dev/disk/by-id/%s) not found", req.VolumeId)
		}

		// Format the volume if not already formatted
		formatted, err := d.DiskHotPlugger.IsFormatted(attachedDiskPath)
		if err != nil {
			d.log.Error(err, "Formatted check errored", "path", attachedDiskPath)
			return nil, err
		}
		d.log.V(1).Info("Is currently formatted?", "volume_id", req.VolumeId, "formatted", formatted)

		if !formatted {
			d.DiskHotPlugger.Format(d.DiskHotPlugger.PathForVolume(req.VolumeId), "ext4")
		}

		// Mount the volume if not already mounted
		mounted, err := d.DiskHotPlugger.IsMounted(d.DiskHotPlugger.PathForVolume(req.VolumeId))
		if err != nil {
			d.log.Error(err, "Mounted check errored", "path", attachedDiskPath)
			return nil, err
		}
		d.log.V(1).Info("Is currently mounted?", "volume_id", req.VolumeId, "mounted", formatted)

		if !mounted {
			mount := req.VolumeCapability.GetMount()
			options := []string{}
			if mount != nil {
				options = mount.MountFlags
			}
			d.DiskHotPlugger.Mount(d.DiskHotPlugger.PathForVolume(req.VolumeId), req.StagingTargetPath, "ext4", options...)
		}
	*/

	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume unmounts the volume when it's finished with, ready for deletion
func (d *Driver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest,
) (*csi.NodeUnstageVolumeResponse, error) {
	d.log.Info("Request: NodeUnstageVolume", "volume_id", req.VolumeId, "staging_target_path", req.StagingTargetPath)

	if req.VolumeId == "" {
		d.log.Error(fmt.Errorf("must provide a VolumeId to NodeUnstageVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to NodeUnstageVolume")
	}
	if req.StagingTargetPath == "" {
		d.log.Error(fmt.Errorf("must provide a StagingTargetPath to NodeUnstageVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a StagingTargetPath to NodeUnstageVolume")
	}

	/*
		d.log.V(1).Info("Unmounting volume (unstaging)", "volume_id", req.VolumeId, "path", req.StagingTargetPath)
		path := d.DiskHotPlugger.PathForVolume(req.VolumeId)

		if path == "" && !d.TestMode {
			d.log.Errorfmt.Errorf("path to volume (/dev/disk/by-id/VOLUME_ID) not found"), "volume_id", req.VolumeId)
			return &csi.NodeUnstageVolumeResponse{}, nil
		}

		mounted, err := d.DiskHotPlugger.IsMounted(path)
		if err != nil {
			d.log.Error(err, "Mounted check errored", "path", path)
			return nil, err
		}
		d.log.V(1).Info("Mounted check completed", "volume_id", req.VolumeId, "mounted", mounted)

		if mounted {
			d.log.V(1).Info("Unmounting", "volume_id", req.VolumeId, "mounted", mounted)
			d.DiskHotPlugger.Unmount(path)
		}
	*/

	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodePublishVolume bind mounts the staging path into the container
func (d *Driver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {
	d.log.Info("Request: NodePublishVolume",
		"VolumeId", req.VolumeId,
		"StagingTargetPath", req.StagingTargetPath,
		"TargetPath", req.TargetPath,
		"VolumeCapability", req.VolumeCapability,
		"VolumeContext", req.VolumeContext)

	if req.VolumeId == "" {
		d.log.Error(fmt.Errorf("must provide a VolumeId to NodePublishVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to NodePublishVolume")
	}
	/*
		if req.StagingTargetPath == "" {
			d.log.Error(fmt.Errorf("must provide a StagingTargetPath to NodePublishVolume"), "Invalid Input")
			return nil, status.Error(codes.InvalidArgument, "must provide a StagingTargetPath to NodePublishVolume")
		}
	*/
	if req.TargetPath == "" {
		d.log.Error(fmt.Errorf("must provide a TargetPath to NodePublishVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a TargetPath to NodePublishVolume")
	}
	if req.VolumeCapability == nil {
		d.log.Error(fmt.Errorf("must provide a VolumeCapability to NodePublishVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeCapability to NodePublishVolume")
	}

	clusterScoped := false
	tkcNamespace, ok := req.VolumeContext[TritonKernelCacheNamespaceIndex]
	if !ok {
		// Namespace is not required. If not provided, then assume it is a Cluster scoped
		// TritonKernelCache instance.
		tkcNamespace = utils.ClusterScopedSubDir
		clusterScoped = true
	}

	tkcName, ok := req.VolumeContext[TritonKernelCacheIndex]
	if !ok {
		if clusterScoped {
			d.log.Error(fmt.Errorf("must provide a TritonKernelCacheCluster"), "Invalid Input")
			return nil, status.Error(codes.InvalidArgument, "must provide a TritonKernelCacheCluster NodePublishVolume")
		} else {
			d.log.Error(fmt.Errorf("must provide a TritonKernelCache"), "Invalid Input")
			return nil, status.Error(codes.InvalidArgument, "must provide a TritonKernelCache NodePublishVolume")
		}
	}

	sourcePath := utils.DefaultCacheDir
	sourcePath = filepath.Join(sourcePath, tkcNamespace)
	sourcePath = filepath.Join(sourcePath, tkcName)
	if _, err := os.Stat(sourcePath); err != nil {
		if os.IsNotExist(err) {
			if clusterScoped {
				d.log.Error(fmt.Errorf("TritonKernelCacheCluster has not been created"), "Invalid Input", "name", tkcName)
				return nil, status.Error(codes.InvalidArgument, "TritonKernelCacheCluster has not been created NodePublishVolume")
			} else {
				d.log.Error(fmt.Errorf("TritonKernelCache has not been created"),
					"Invalid Input", "name", tkcName, "namespace", tkcNamespace)
				return nil, status.Error(codes.InvalidArgument, "TritonKernelCache has not been created NodePublishVolume")
			}
		} else {
			d.log.Error(fmt.Errorf("unable to verify sourcePath"), "Invalid Input", "sourcePath", sourcePath)
			return nil, status.Error(codes.InvalidArgument, "unable to verify sourcePath")
		}
	}
	d.log.Info("Found TritonKernelCache CRD", "sourcePath", sourcePath)

	if err := os.MkdirAll(req.TargetPath, 0o750); err != nil {
		d.log.Error(err, "Failed to create target path", "volume_id", req.VolumeId, "targetPath", req.TargetPath)
		return nil, err
	}

	// Check if already mounted
	mounted, err := utils.IsTargetBindMount(req.TargetPath)
	if err != nil {
		d.log.Error(fmt.Errorf("unable to verify if targetPath already mounted"),
			"Invalid Input", "targetPath", req.TargetPath)
		return nil, status.Error(codes.InvalidArgument, "unable to verify if targetPath already mounted")
	}

	// If Already Mounted
	if mounted {
		_, ok := d.volumeIdMapping[req.VolumeId]
		if !ok {
			// Save off the VolumeId mapping to CRD Info
			d.volumeIdMapping[req.VolumeId] = CacheData{
				KernelName:    tkcName,
				Namespace:     tkcNamespace,
				ClusterScoped: clusterScoped,
			}
		}

		d.log.Info("kernel cache already bind mounted", "sourcePath", sourcePath, "targetPath", req.TargetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	d.log.Info("bind mounting kernel cache", "sourcePath", sourcePath, "targetPath", req.TargetPath)

	// Perform the bind mount
	options := []string{"bind"}
	if req.Readonly {
		options = append(options, "ro")
	}

	if err := d.mounter.Mount(sourcePath, req.TargetPath, "", options); err != nil {
		d.log.Error(fmt.Errorf("bind mount failed"), "Invalid Input", "sourcePath", sourcePath, "targetPath", req.TargetPath)
		return nil, status.Error(codes.Internal, "bind mount failed")
	}

	// Save off the VolumeId mapping to CRD Info
	d.volumeIdMapping[req.VolumeId] = CacheData{
		KernelName:    tkcName,
		Namespace:     tkcNamespace,
		ClusterScoped: clusterScoped,
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume removes the bind mount
func (d *Driver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {
	d.log.Info("Request: NodeUnpublishVolume",
		"VolumeId", req.VolumeId,
		"TargetPath", req.TargetPath)

	if req.VolumeId == "" {
		d.log.Error(fmt.Errorf("must provide a VolumeId to NodeUnpublishVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to NodeUnpublishVolume")
	}
	if req.TargetPath == "" {
		d.log.Error(fmt.Errorf("must provide a TargetPath to NodeUnpublishVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a TargetPath to NodeUnpublishVolume")
	}

	/*cacheData*/
	_, ok := d.volumeIdMapping[req.VolumeId]
	if !ok {
		d.log.Error(fmt.Errorf("could not map VolumeId to TritonKernelCache"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "could not map VolumeId to TritonKernelCache NodeUnpublishVolume")
	}

	// Check if already mounted
	// d.mounter.IsLikelyNotMountPoint() doesn't detect bind mounts, so manually search
	// the list of mounts for the Target Path.
	mounted, err := utils.IsTargetBindMount(req.TargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			d.log.Info("targetPath does not exist, just continue", "VolumeId", req.VolumeId, "TargetPath", req.TargetPath,
				"Mount", mounted)
		} else {
			d.log.Error(fmt.Errorf("unable to verify if targetPath already mounted"),
				"Internal Error", "targetPath", req.TargetPath)
			return nil, status.Error(codes.InvalidArgument, "unable to verify if targetPath already mounted")
		}
	}

	// Only attempt to unmount if it's mounted
	if mounted {
		if err := d.mounter.Unmount(req.TargetPath); err != nil {
			d.log.Error(fmt.Errorf("umount failed"), "Invalid Input", "targetPath", req.TargetPath)
			return nil, status.Error(codes.Internal, "umount failed")
		}
	} else {
		d.log.Info("targetPath is not mounted, just continue", "VolumeId", req.VolumeId, "TargetPath", req.TargetPath)
	}

	delete(d.volumeIdMapping, req.VolumeId)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetInfo returns some identifier (ID, name) for the current node
func (d *Driver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	d.log.Info("Request: NodeGetInfo")

	return &csi.NodeGetInfoResponse{
		NodeId:            d.NodeName,
		MaxVolumesPerNode: MaxVolumesPerNode,

		AccessibleTopology: &csi.Topology{
			Segments: map[string]string{
				"region": "unknown",
			},
		},
	}, nil
}

type VolumeStatistics struct {
	AvailableBytes, TotalBytes, UsedBytes    int64
	AvailableInodes, TotalInodes, UsedInodes int64
}

// NodeGetVolumeStats returns the volume capacity statistics available for the the given volume
func (d *Driver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {
	d.log.Info("Request: NodeGetVolumeStats", "volume_id", req.VolumeId)

	if req.VolumeId == "" {
		d.log.Error(fmt.Errorf("must provide a VolumeId to NodeGetVolumeStats"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to NodeGetVolumeStats")
	}

	volumePath := req.VolumePath
	if volumePath == "" {
		d.log.Error(fmt.Errorf("must provide a VolumePath to NodeGetVolumeStats"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumePath to NodeGetVolumeStats")
	}

	/*
		mounted, err := d.DiskHotPlugger.IsMounted(volumePath)
		if err != nil {
			log.Error(err, "Failed to check if volume path is mounted", "volume_id", req.VolumeId, "path", volumePath)
			return nil, status.Errorf(codes.Internal, "failed to check if volume path %q is mounted: %s", volumePath, err)
		}

		if !mounted {
			log.Error(fmt.Errorf("Volume path is not mounted"), "Invalid Input", "volume_id", req.VolumeId, "path", volumePath)
			return nil, status.Errorf(codes.NotFound, "volume path %q is not mounted", volumePath)
		}

		stats, err := d.DiskHotPlugger.GetStatistics(volumePath)
		if err != nil {
			log.Error(err, "Failed to retrieve capacity statistics", "volume_id", req.VolumeId, "path", volumePath)
			return nil,
				status.Errorf(
					codes.Internal,
					"failed to retrieve capacity statistics for volume path %q: %s",
					volumePath,
					err,
				)
		}
	*/

	var stats VolumeStatistics

	d.log.Info("Node capacity statistics retrieved",
		"bytes_available", stats.AvailableBytes,
		"bytes_total", stats.TotalBytes,
		"bytes_used", stats.UsedBytes,
		"inodes_available", stats.AvailableInodes,
		"inodes_total", stats.TotalInodes,
		"inodes_used", stats.UsedInodes)

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Available: stats.AvailableBytes,
				Total:     stats.TotalBytes,
				Used:      stats.UsedBytes,
				Unit:      csi.VolumeUsage_BYTES,
			},
			{
				Available: stats.AvailableInodes,
				Total:     stats.TotalInodes,
				Used:      stats.UsedInodes,
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}, nil
}

// NodeExpandVolume is used to expand the filesystem inside volumes
func (d *Driver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest,
) (*csi.NodeExpandVolumeResponse, error) {
	d.log.Info("Request: NodeExpandVolume", "volume_id", req.VolumeId, "target_path", req.VolumePath)
	if req.VolumeId == "" {
		d.log.Error(fmt.Errorf("must provide a VolumeId to NodeExpandVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to NodeExpandVolume")
	}
	if req.VolumePath == "" {
		d.log.Error(fmt.Errorf("must provide a VolumePath to NodeExpandVolume"), "Invalid Input")
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumePath to NodeExpandVolume")
	}

	/*
		_, err := d.TkmClient.GetVolume(req.VolumeId)
		if err != nil {
			d.log.Error(err, "Failed to find VolumeID to NodeExpandVolume", "volume_id", req.VolumeId)
			return nil, status.Errorf(codes.NotFound, "unable to find VolumeID %q to NodeExpandVolume: %s", req.VolumeId, err)
		}
		// Find the disk attachment location
		attachedDiskPath := d.DiskHotPlugger.PathForVolume(req.VolumeId)
		if attachedDiskPath == "" {
			log.Error(fmt.Errorf("path to volume (/dev/disk/by-id/VOLUME_ID) not found"),
				"Invalid Input", "volume_id", req.VolumeId)
			return nil, status.Errorf(codes.NotFound, "path to volume (/dev/disk/by-id/%s) not found", req.VolumeId)
		}

		log.Info("Expanding Volume", "volume_id", req.VolumeId, "path", attachedDiskPath)
		err = d.DiskHotPlugger.ExpandFilesystem(d.DiskHotPlugger.PathForVolume(req.VolumeId))
		if err != nil {
			log.Error(err, "Failed to expand filesystem", "volume_id", req.VolumeId)
			return nil, status.Errorf(codes.Internal, "failed to expand file system: %s", err)
		}
	*/

	// TODO: Get new size for resposne

	return &csi.NodeExpandVolumeResponse{}, nil
}

// NodeGetCapabilities returns the capabilities that this node and driver support
func (d *Driver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {
	// Intentionally don't return VOLUME_CONDITION and NODE_GET_VOLUME_STATS
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
					},
				},
			},
		},
	}, nil
}
