package driver

import (
	"context"
	"fmt"

	// "github.com/tkm/tkmgo"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BytesInGigabyte describes how many bytes are in a gigabyte
const BytesInGigabyte int64 = 1024 * 1024 * 1024

// TkmVolumeAvailableRetries is the number of times we will retry to check if a volume is available
const TkmVolumeAvailableRetries int = 20

var supportedAccessModes = map[csi.VolumeCapability_AccessMode_Mode]struct{}{
	csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER:      {},
	csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY: {},
}

// CreateVolume is the first step when a PVC tries to create a dynamic volume
func (d *Driver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	d.log.Info("Request: CreateVolume")

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume Name must be provided")
	}

	if len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume Volume capabilities must be provided")
	}

	d.log.Info("Creating volume", "name", req.Name, "capabilities", req.VolumeCapabilities)

	// Check capabilities
	for _, cap := range req.VolumeCapabilities {
		if _, ok := supportedAccessModes[cap.GetAccessMode().GetMode()]; !ok {
			return nil, status.Error(codes.InvalidArgument, "CreateVolume access mode isn't supported")
		}
		if _, ok := cap.GetAccessType().(*csi.VolumeCapability_Block); ok {
			return nil, status.Error(codes.InvalidArgument, "CreateVolume block types aren't supported, only mount types")
		}
	}

	// Determine required size
	bytes, err := getVolSizeInBytes(req.GetCapacityRange())
	if err != nil {
		return nil, err
	}

	desiredSize := bytes / BytesInGigabyte
	if (bytes % BytesInGigabyte) != 0 {
		desiredSize++
	}

	d.log.V(1).Info("Volume size determined", "size_gb", desiredSize)

	d.log.V(1).Info("Listing current volumes in TKM API")
	/*
		volumes, err := d.TkmClient.ListVolumes()
		if err != nil {
			d.log.Error(err, "Unable to list volumes in TKM API")
			return nil, err
		}
		for _, v := range volumes {
			if v.Name == req.Name {
				d.log.V(1).Info("Volume already exists", "volume_id", v.ID)
				if v.SizeGigabytes != int(desiredSize) {
					return nil, status.Error(codes.AlreadyExists, "Volume already exists with a differnt size")
				}

				available, err := d.waitForVolumeStatus(&v, "available", TkmVolumeAvailableRetries)
				if err != nil {
					d.log.Error(err, "Unable to wait for volume availability in TKM API")
					return nil, err
				}

				if available {
					return &csi.CreateVolumeResponse{
						Volume: &csi.Volume{
							VolumeId:      v.ID,
							CapacityBytes: int64(v.SizeGigabytes) * BytesInGigabyte,
						},
					}, nil
				}

				d.log.Error(fmt.Errorf("TKM Volume is not 'available'"), "status", v.Status)
				return nil, status.Errorf(codes.Unavailable, "Volume isn't available to be attached, state is currently %s", v.Status)
			}
		}
	*/

	// TODO: Uncomment after client implementation is complete.
	// snapshotID := ""
	// if volSource := req.GetVolumeContentSource(); volSource != nil {
	// 	if _, ok := volSource.GetType().(*csi.VolumeContentSource_Snapshot); !ok {
	// 		return nil, status.Error(codes.InvalidArgument, "Unsupported volumeContentSource type")
	// 	}
	// 	snapshot := volSource.GetSnapshot()
	// 	if snapshot == nil {
	// 		return nil, status.Error(codes.InvalidArgument, "Volume content source type is set to Snapshot, but the Snapshot is not provided")
	// 	}
	// 	snapshotID = snapshot.GetSnapshotId()
	// 	if snapshotID == "" {
	// 		return nil, status.Error(codes.InvalidArgument, "Volume content source type is set to Snapshot, but the SnapshotID is not provided")
	// 	}
	// }

	d.log.V(1).Info("Volume doesn't currently exist, will need creating")

	d.log.V(1).Info("Requesting available capacity in client's quota from the TKM API")
	/*
		quota, err := d.TkmClient.GetQuota()
		if err != nil {
			d.log.Error(err, "Unable to get quota from TKM API")
			return nil, err
		}
		availableSize := int64(quota.DiskGigabytesLimit - quota.DiskGigabytesUsage)
		if availableSize < desiredSize {
			d.log.Error(fmt.Errorf("Requested volume would exceed storage quota available"))
			return nil, status.Errorf(codes.OutOfRange, "Requested volume would exceed volume space quota by %d GB", desiredSize-availableSize)
		} else if quota.DiskVolumeCountUsage >= quota.DiskVolumeCountLimit {
			d.log.Error(fmt.Errorf("Requested volume would exceed volume quota available"))
			return nil, status.Errorf(codes.OutOfRange, "Requested volume would exceed volume count limit quota of %d", quota.DiskVolumeCountLimit)
		}

		d.log.V(1).Info("Quota has sufficient capacity remaining",
			"disk_gb_limit", quota.DiskGigabytesLimit, "disk_gb_usage", quota.DiskGigabytesUsage)
	*/

	/*
		v := &tkmgo.VolumeConfig{
			Name:          req.Name,
			Region:        d.Region,
			Namespace:     d.Namespace,
			ClusterID:     d.ClusterID,
			SizeGigabytes: int(desiredSize),
			// SnapshotID: snapshotID, // TODO: Uncomment after client implementation is complete.
		}
	*/
	d.log.V(1).Info("Creating volume in TKM API")
	/*
		result, err := d.TkmClient.NewVolume(v)
		if err != nil {
			d.log.Error(err, "Unable to create volume in TKM API")
			return nil, err
		}
	*/

	d.log.Info("Volume created in TKM API", "volume_id", 14567 /*result.ID*/)

	/*
		volume, err := d.TkmClient.GetVolume(result.ID)
		if err != nil {
			d.log.Error(err, "Unable to get volume updates in TKM API")
			return nil, err
		}

		d.log.V(1).Info("Waiting for volume to become available in TKM API", "volume_id", result.ID)
		available, err := d.waitForVolumeStatus(volume, "available", TkmVolumeAvailableRetries)
		if err != nil {
			d.log.Error(err, "Volume availability never completed successfully in TKM API")
			return nil, err
		}
	*/
	available := true

	if available {
		return &csi.CreateVolumeResponse{
			Volume: &csi.Volume{
				// VolumeId:      volume.ID,
				VolumeId:      "14567",
				CapacityBytes: int64(1 /*v.SizeGigabytes*/) * BytesInGigabyte,
			},
		}, nil
	}

	d.log.Error(err, "TKM Volume is not 'available'")
	return nil, status.Errorf(codes.Unavailable, "TKM Volume %q is not \"available\", state currently is %q", "14567" /*volume.ID*/, "error" /*volume.Status*/)
}

// waitForVolumeAvailable will just sleep/loop waiting for TKM's API to report it's available, or hit a defined
// number of retries
/*
func (d *Driver) waitForVolumeStatus(vol *tkmgo.Volume, desiredStatus string, retries int) (bool, error) {
	d.log.Info("Waiting for Volume to entered desired state", "volume_id", vol.ID, "desired_state", desiredStatus)
	var v *tkmgo.Volume
	var err error

	if d.TestMode {
		return true, nil
	}

	for i := 0; i < retries; i++ {
		time.Sleep(5 * time.Second)

		v, err = d.TkmClient.GetVolume(vol.ID)
		if err != nil {
			d.log.Error(err, "Unable to get volume updates in TKM API")
			return false, err
		}

		if v.Status == desiredStatus {
			return true, nil
		}
	}
	return false, fmt.Errorf("volume isn't %s, state is currently %s", desiredStatus, v.Status)
}
*/

// DeleteVolume is used once a volume is unused and therefore unmounted, to stop the resources being used and subsequent billing
func (d *Driver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	d.log.Info("Request: DeleteVolume", "volume_id", req.VolumeId)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to DeleteVolume")
	}

	d.log.V(1).Info("Deleting volume in TKM API")
	/*
		_, err := d.TkmClient.DeleteVolume(req.VolumeId)
		if err != nil {
			if strings.Contains(err.Error(), "DatabaseVolumeNotFoundError") {
				d.log.Info("Volume already deleted from TKM API", "volume_id", req.VolumeId)
				return &csi.DeleteVolumeResponse{}, nil
			}

			d.log.Error(err, "Unable to delete volume in TKM API")
			return nil, err
		}
	*/

	d.log.Info("Volume deleted from TKM API", "volume_id", req.VolumeId)

	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume is used to mount an underlying volume to required k3s node
func (d *Driver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	d.log.Info("Request: ControllerPublishVolume", "volume_id", req.VolumeId, "node_id", req.NodeId)

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeCapability to ControllerPublishVolume")
	}

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to ControllerPublishVolume")
	}

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "must provide a NodeId to ControllerPublishVolume")
	}

	d.log.V(1).Info("Check if Node exits")
	/*
		cluster, err := d.TkmClient.GetKubernetesCluster(d.ClusterID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "unable to connect to TKM Api. error: %s", err)
		}
		found := false
		for _, instance := range cluster.Instances {
			if instance.ID == req.NodeId {
				found = true
				break
			}
		}
		if !found {
			return nil, status.Error(codes.NotFound, "Unable to find instance to attach volume to")
		}

		d.log.V(1).Info("Finding volume in TKM API")
		volume, err := d.TkmClient.GetVolume(req.VolumeId)
		if err != nil {
			d.log.Error(err, "Unable to find volume for publishing in TKM API")
			return nil, err
		}
		d.log.V(1).Info("Volume found for publishing in TKM API", "volume_id", volume.ID)

		// Check if the volume is already attached to the requested node
		if volume.InstanceID == req.NodeId && volume.Status == "attached" {
			d.log.Info("Volume is already attached to the requested instance",
				"volume_id", volume.ID,
				"instance_id", req.NodeId)
			return &csi.ControllerPublishVolumeResponse{}, nil
		}

		// if the volume is not available, we can't attach it, so error out
		if volume.Status != "available" && volume.InstanceID != req.NodeId {
			d.log.Error(fmt.Errorf("Volume is not available to be attached"),
				"volume_id", volume.ID,
				"status", volume.Status,
				"requested_instance_id", req.NodeId,
				"current_instance_id", volume.InstanceID,
				"Volume is not available to be attached")
			return nil, status.Errorf(codes.Unavailable, "Volume %q is not available to be attached, state is currently %q", volume.ID, volume.Status)
		}

		// Check if the volume is attaching to this node
		if volume.InstanceID == req.NodeId && volume.Status != "attaching" {
			// Do nothing, the volume is already attaching
			d.log.V(1).Info("Volume is already attaching", "volume_id", volume.ID, "status", volume.Status)
		} else {
			// Call the TKM API to attach it to a node/instance
			d.log.V(1).Info("Requesting volume to be attached in TKN API",
				"volume_id", volume.ID,
				"volume_status", volume.Status,
				"reqested_instance_id", req.NodeId)

			volConfig := tkmgo.VolumeAttachConfig{
				InstanceID: req.NodeId,
				Region:     d.Region,
			}

			_, err = d.TkmClient.AttachVolume(req.VolumeId, volConfig)
			if err != nil {
				d.log.Error(err, "Unable to attach volume in TKM API")
				return nil, err
			}
			d.log.Info("Volume successfully requested to be attached in TKM API",
				"volume_id", volume.ID,
				"instance_id", req.NodeId)
		}

		time.Sleep(5 * time.Second)
		// refetch the volume
		d.log.Info("Fetching volume again to check status after attaching", "volume_id", volume.ID)
		volume, err = d.TkmClient.GetVolume(req.VolumeId)
		if err != nil {
			d.log.Error(err, "Unable to fetch volume from TKM API")
			return nil, err
		}
		if volume.Status != "attached" {
			d.log.Error(fmt.Errorf("Volume is not in the attached state"), "volume_id", volume.ID, "status", volume.Status)
			return nil, status.Errorf(codes.Unavailable, "Volume %q is not attached to the requested instance, state is currently %q", volume.ID, volume.Status)
		}

		if volume.InstanceID != req.NodeId {
			d.log.Error(fmt.Errorf("Volume is not attached to the requested instance"), "volume_id", volume.ID, "instance_id", req.NodeId)
			return nil, status.Errorf(codes.Unavailable, "Volume %q is not attached to the requested instance %q, instance id is currently %q", volume.ID, req.NodeId, volume.InstanceID)
		}
	*/

	d.log.V(1).Info("Volume successfully attached in TKM API", "volume_id", req.VolumeId /*volume.ID*/)
	return &csi.ControllerPublishVolumeResponse{}, nil
}

// ControllerUnpublishVolume detaches the volume from the k3s node it was connected
func (d *Driver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	d.log.Info("Request: ControllerUnpublishVolume", "volume_id", req.VolumeId)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to ControllerUnpublishVolume")
	}

	/*
		d.log.V(1).Info("Finding volume in TKM API")
		volume, err := d.TkmClient.GetVolume(req.VolumeId)
		if err != nil {
			if strings.Contains(err.Error(), "DatabaseVolumeNotFoundError") || strings.Contains(err.Error(), "ZeroMatchesError") {
				d.log.Info("Volume already deleted from TKM API, pretend it's unmounted", "volume_id", req.VolumeId)
				return &csi.ControllerUnpublishVolumeResponse{}, nil
			}
			d.log.V(1).Info("message", err.Error()).Msg("Error didn't match DatabaseVolumeNotFoundError")

			d.log.Error(err, "Unable to find volume for unpublishing in TKM API")
			return nil, err
		}

		d.log.V(1).Info("Volume found for unpublishing in TKM API", "volume_id", volume.ID)

		// If the volume is currently available, it's not attached to anything to return success
		if volume.Status == "available" {
			d.log.Info("Volume is already available, no need to unpublish", "volume_id", volume.ID)
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}

		// If requeseted node doesn't match the current volume instance, return success
		if volume.InstanceID != req.NodeId {
			d.log.Info("Volume is not attached to the requested instance",
				"volume_id", volume.ID,
				"instance_id", volume.InstanceID,
				"requested_instance_id", req.NodeId)
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}

		if volume.Status != "detaching" {
			// The volume is either attached to the requested node or the requested node is empty
			// and the volume is attached, so we need to detach the volume
			d.log.Info("Requesting volume to be detached",
				"volume_id", volume.ID,
				"current_instance_id", volume.InstanceID,
				"requested_instance_id", req.NodeId,
				"status", volume.Status)

			_, err = d.TkmClient.DetachVolume(req.VolumeId)
			if err != nil {
				d.log.Error(err, "Unable to detach volume in TKM API")
				return nil, err
			}

			d.log.Info("Volume sucessfully requested to be detached in TKM API", "volume_id", volume.ID)
		}

		// Fetch the new state after 5 seconds
		time.Sleep(5 * time.Second)
		volume, err = d.TkmClient.GetVolume(req.VolumeId)
		if err != nil {
			d.log.Error(err, "Unable to find volume for unpublishing in TKM API")
			return nil, err
		}
	*/

	/*
		if volume.Status == "available" {
	*/
	d.log.V(1).Info("Volume is now available again", "volume_id", req.VolumeId /*volume.ID*/)
	return &csi.ControllerUnpublishVolumeResponse{}, nil
	/*
		}
	*/

	/*
		// err that the the volume is not available
		d.log.Error(fmt.Errorf("TKM Volume did not go back to 'available' status"))
		return nil, status.Errorf(codes.Unavailable, "TKM Volume %q did not go back to \"available\", state is currently %q", req.VolumeId, volume.Status)
	*/
}

// ValidateVolumeCapabilities returns the features of the volume, e.g. RW, RO, RWX
func (d *Driver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	d.log.Info("Request: ValidateVolumeCapabilities", "volume_id", req.VolumeId)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to ValidateVolumeCapabilities")
	}

	if req.VolumeCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "must provide VolumeCapabilities to ValidateVolumeCapabilities")
	}

	/*
		_, err := d.TkmClient.GetVolume(req.VolumeId)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "Unable to fetch volume from TKM API: %s", err)
		}

		accessModeSupported := false
		for _, cap := range req.VolumeCapabilities {
			if _, ok := supportedAccessModes[cap.GetAccessMode().GetMode()]; ok {
				accessModeSupported = true
				break
			}
		}

		if !accessModeSupported {
			return nil, status.Errorf(codes.NotFound, "%v not supported", req.GetVolumeCapabilities())
		}
	*/

	resp := &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: req.VolumeCapabilities,
		},
	}

	return resp, nil
}

// ListVolumes returns the existing TKM volumes for this customer
func (d *Driver) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	if req.StartingToken != "" {
		return &csi.ListVolumesResponse{}, status.Errorf(codes.Aborted, "%v not supported", "starting-token")
	}

	d.log.Info("Request: ListVolumes")

	d.log.V(1).Info("Listing all volume in TKM API")
	/*
		volumes, err := d.TkmClient.ListVolumes()
		if err != nil {
			d.log.Error(err, "Unable to list volumes in TKM API")
			return nil, err
		}
	*/
	d.log.V(1).Info("Successfully retrieved all volumes from the TKM API")

	resp := &csi.ListVolumesResponse{
		Entries: []*csi.ListVolumesResponse_Entry{},
	}

	/*
		for _, v := range volumes {
	*/
	resp.Entries = append(resp.Entries, &csi.ListVolumesResponse_Entry{
		Volume: &csi.Volume{
			CapacityBytes: int64(1 /*v.SizeGigabytes*/) * BytesInGigabyte,
			VolumeId:      "14567", /*v.ID*/
			ContentSource: &csi.VolumeContentSource{
				Type: &csi.VolumeContentSource_Volume{},
			},
		},
		Status: &csi.ListVolumesResponse_VolumeStatus{},
	})
	/*
		}
	*/

	return resp, nil
}

// GetCapacity calls the TKM API to determine the user's available quota
func (d *Driver) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	d.log.Info("Request: GetCapacity")

	d.log.V(1).Info("Requesting available capacity in client's quota from the TKM API")
	/*
		quota, err := d.TkmClient.GetQuota()
		if err != nil {
			d.log.Error(err, "Unable to get quota in TKM API")
			return nil, err
		}
	*/
	d.log.V(1).Info("Successfully retrieved quota from the TKM API")

	availableBytes := int64(1 /*quota.DiskGigabytesLimit-quota.DiskGigabytesUsage*/) * BytesInGigabyte
	d.log.V(1).Info("Available capacity determined", "available_gb", availableBytes/BytesInGigabyte)
	if availableBytes < BytesInGigabyte {
		d.log.Error(fmt.Errorf("available capacity is less than 1GB, volumes can't be launched"),
			"available_bytes", availableBytes)
	}

	/*
		if quota.DiskVolumeCountUsage >= quota.DiskVolumeCountLimit {
			d.log.Error("Number of volumes is at the quota limit, no capacity left")
			availableBytes = 0
		}
	*/

	resp := &csi.GetCapacityResponse{
		AvailableCapacity: availableBytes,
	}

	return resp, nil
}

// ControllerGetCapabilities returns the capabilities of the controller, what features it implements
func (d *Driver) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	d.log.Info("Request: ControllerGetCapabilities")

	rawCapabilities := []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
		csi.ControllerServiceCapability_RPC_GET_CAPACITY,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		// csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT, TODO: Uncomment after client implementation is complete.
		// csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS, TODO: Uncomment after client implementation is complete.
	}

	var csc []*csi.ControllerServiceCapability

	for _, cap := range rawCapabilities {
		csc = append(csc, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		})
	}

	d.log.V(1).Info("Capabilities for controller requested", "capabilities", csc)

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: csc,
	}

	return resp, nil
}

// CreateSnapshot is part of implementing Snapshot & Restore functionality, but we don't support that
func (d *Driver) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
	// TODO: Uncomment after client implementation is complete.
	// snapshotName := req.GetName()
	// sourceVolID := req.GetSourceVolumeId()
	//
	// d.log.Info("Request: CreateSnapshot", "snapshot_name", snapshotName, "source_volume_id", sourceVolID)
	//
	// if len(snapshotName) == 0 {
	// 	return nil, status.Error(codes.InvalidArgument, "Snapshot name is required")
	// }
	// if len(sourceVolID) == 0 {
	// 	return nil, status.Error(codes.InvalidArgument, "SourceVolumeId is required")
	// }
	//
	// d.log.V(1).Info("Finding current snapshot in TKM API",
	// 	"source_volume_id", sourceVolID)
	//
	// snapshots, err := d.TkmClient.ListVolumeSnapshotsByVolumeID(sourceVolID)
	// if err != nil {
	// 	d.log.Error(err, "Unable to list snapshot in TKM API", "source_volume_id", sourceVolID)
	// 	return nil, status.Errorf(codes.Internal, "failed to list snapshots by %q: %s", sourceVolID, err)
	// }
	//
	// // Check for an existing snapshot with the specified name.
	// for _, snapshot := range snapshots {
	// 	if snapshot.Name != snapshotName {
	// 		continue
	// 	}
	// 	if snapshot.VolumeID == sourceVolID {
	// 		return &csi.CreateSnapshotResponse{
	// 			Snapshot: &csi.Snapshot{
	// 				SnapshotId:     snapshot.SnapshotID,
	// 				SourceVolumeId: snapshot.VolumeID,
	// 				CreationTime:   snapshot.CreationTime,
	// 				SizeBytes:      snapshot.RestoreSize,
	// 				ReadyToUse:     true,
	// 			},
	// 		}, nil
	// 	}
	// 	d.log.Error(err,
	//		"Snapshot with the same name but with different SourceVolumeId already exist",
	// 		"snapshot_name", snapshotName,
	// 		"requested_source_volume_id", sourceVolID,
	// 		"actual_source_volume_id", snapshot.VolumeID)
	// 	return nil, status.Errorf(codes.AlreadyExists, "snapshot with the same name %q but with different SourceVolumeId already exist", snapshotName)
	// }
	//
	// d.log.V(1).Info("Create volume snapshot in TKM API",
	// 	"snapshot_name", snapshotName,
	// 	"source_volume_id", sourceVolID)
	//
	// result, err := d.TkmClient.CreateVolumeSnapshot(sourceVolID, &tkmgo.VolumeSnapshotConfig{
	// 	Name: snapshotName,
	// })
	// if err != nil {
	// 	if strings.Contains(err.Error(), "DatabaseVolumeSnapshotLimitExceededError") {
	// 		d.log.Error(err, "Requested volume snapshot would exceed volume quota available")
	// 		return nil, status.Errorf(codes.ResourceExhausted, "failed to create volume snapshot due to over quota: %s", err)
	// 	}
	// 	d.log.Error(err, "Unable to create snapshot in TKM API")
	// 	return nil, status.Errorf(codes.Internal, "failed to create volume snapshot: %s", err)
	// }
	//
	// d.log.Info("Snapshot created in TKM API", "snapshot_id", result.SnapshotID)
	//
	// // NOTE: Add waitFor logic if creation takes long time.
	// snapshot, err := d.TkmClient.GetVolumeSnapshot(result.SnapshotID)
	// if err != nil {
	// 	d.log.Error(err, "Unsable to get snapshot updates from TKM API", "snapshot_id", result.SnapshotID)
	// 	return nil, status.Errorf(codes.Internal, "failed to get snapshot by %q: %s", result.SnapshotID, err)
	// }
	// return &csi.CreateSnapshotResponse{
	// 	Snapshot: &csi.Snapshot{
	// 		SnapshotId:     snapshot.SnapshotID,
	// 		SourceVolumeId: snapshot.VolumeID,
	// 		CreationTime:   snapshot.CreationTime,
	// 		SizeBytes:      snapshot.RestoreSize,
	// 		ReadyToUse:     true,
	// 	},
	// }, nil
}

// DeleteSnapshot is part of implementing Snapshot & Restore functionality, and it will be supported in the future.
func (d *Driver) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	d.log.Info("Request: DeleteSnapshot", "snapshot_id", req.GetSnapshotId())

	snapshotID := req.GetSnapshotId()
	if snapshotID == "" {
		return nil, status.Error(codes.InvalidArgument, "must provide SnapshotId to DeleteSnapshot")
	}

	d.log.V(1).Info("Deleting snapshot in TKM API", "snapshot_id", snapshotID)

	// TODO: Uncomment after client implementation is complete.
	// _, err := d.TkmClient.DeleteVolumeSnapshot(snapshotID)
	// if err != nil {
	// 	if strings.Contains(err.Error(), "DatabaseVolumeSnapshotNotFoundError") {
	// 		d.log.Info("Snapshot already deleted from TKM API", "volume_id", snapshotID)
	// 		return &csi.DeleteSnapshotResponse{}, nil
	// 	} else if strings.Contains(err.Error(), "DatabaseSnapshotCannotDeleteInUseError") {
	// 		return nil, status.Errorf(codes.FailedPrecondition, "failed to delete snapshot %q, it is currently in use, err: %s", snapshotID, err)
	// 	}
	// 	return nil, status.Errorf(codes.Internal, "failed to delete snapshot %q, err: %s", snapshotID, err)
	// }
	// return &csi.DeleteSnapshotResponse{}, nil
	return nil, status.Error(codes.Unimplemented, "")
}

// ListSnapshots retrieves a list of existing snapshots as part of the Snapshot & Restore functionality.
func (d *Driver) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	d.log.Info("Request: ListSnapshots")

	snapshotID := req.GetSnapshotId()
	sourceVolumeID := req.GetSourceVolumeId()

	if req.GetStartingToken() != "" {
		d.log.Error(
			fmt.Errorf("ListSnapshots RPC received a Starting token, but pagination is not supported. Ensure the request does not include a starting token"),
			"Invalid Input")
		return nil, status.Error(codes.Aborted, "starting-token not supported")
	}

	// case 1: SnapshotId is not empty, return snapshots that match the snapshot id
	if len(snapshotID) != 0 {
		d.log.V(1).Info("Fetching snapshot", "snapshot_id", snapshotID)

		// Retrieve a specific snapshot by ID
		// Todo: GetSnapshot to be implemented in tkmgo
		// Todo: Un-comment post client implementation
		// snapshot, err := d.TkmClient.GetSnapshot(snapshotID)
		// if err != nil {
		// 	// Todo: DatabaseSnapshotNotFoundError & DiskSnapshotNotFoundError are placeholders, it's still not clear what error will be returned by API (awaiting implementation - WIP)
		// 	if strings.Contains(err.Error(), "DatabaseSnapshotNotFoundError") ||
		// 		strings.Contains(err.Error(), "DiskSnapshotNotFoundError") {
		// 		d.log.Info("ListSnapshots: no snapshot found, returning with success", "snapshot_id", snapshotID)
		// 		return &csi.ListSnapshotsResponse{}, nil
		// 	}
		// 	d.log.Error(err, "Failed to list snapshot from TKM API" , "snapshot_id", snapshotID)
		// 	return nil, status.Errorf(codes.Internal, "failed to list snapshot %q: %v", snapshotID, err)
		// }
		// return &csi.ListSnapshotsResponse{
		// 	Entries: []*csi.ListSnapshotsResponse_Entry{convertSnapshot(snapshot)},
		// }, nil
	}

	// case 2: Retrieve snapshots by source volume ID
	if len(sourceVolumeID) != 0 {
		d.log.V(1).Info("Fetching volume snapshots",
			"operation", "list_snapshots",
			"source_volume_id", sourceVolumeID)

		// snapshots, err := d.TkmClient.ListSnapshots() // Todo: ListSnapshots to be implemented in tkmgo
		// if err != nil {
		// 	d.log.Error(err, "Failed to list snapshots for volume", "source_volume_id", sourceVolumeID)
		// 	return nil, status.Errorf(codes.Internal, "failed to list snapshots for volume %q: %v", sourceVolumeID, err)
		// }
		//
		// entries := []*csi.ListSnapshotsResponse_Entry{}
		// for _, snapshot := range snapshots {
		// 	if snapshot.VolID == sourceVolumeID {
		// 		entries = append(entries, convertSnapshot(snapshot))
		// 	}
		// }
		//
		// return &csi.ListSnapshotsResponse{
		// 	Entries: entries,
		// }, nil
	}

	d.log.V(1).Info("Fetching all snapshots")

	// case 3: Retrieve all snapshots if no filters are provided
	// Todo: un-comment post client(tkmgo) implementation
	// snapshots, err := d.TkmClient.ListSnapshots() // Todo: ListSnapshots to be implemented in tkmgo
	// if err != nil {
	// 	d.log.Error(err, "Failed to list snapshots from TKM API")
	// 	return nil, status.Errorf(codes.Internal, "failed to list snapshots from TKM API: %v", err)
	// }
	//
	// sort.Slice(snapshots, func(i, j int) bool {
	// 	return snapshots[i].Id < snapshots[j].Id
	// })
	//
	// entries := []*csi.ListSnapshotsResponse_Entry{}
	// for _, snap := range snapshots {
	// 	entries = append(entries, convertSnapshot(snap))
	// }
	//
	// d.log.Info("Snapshots listed successfully", "total_snapshots", len(entries))
	//
	// return &csi.ListSnapshotsResponse{
	// 	Entries: entries,
	// }, nil
	return nil, status.Error(codes.Unimplemented, "")
}

func getVolSizeInBytes(capRange *csi.CapacityRange) (int64, error) {
	if capRange == nil {
		return int64(DefaultVolumeSizeGB) * BytesInGigabyte, nil
	}

	// Volumes can be of a flexible size, but they must specify one of the fields, so we'll use that
	bytes := capRange.GetRequiredBytes()
	if bytes == 0 {
		bytes = capRange.GetLimitBytes()
	}

	return bytes, nil
}

// Todo: Un-comment post client implementation is complete
// Todo: Snapshot to be defined in tkmgo
// convertSnapshot function converts a tkmgo.Snapshot object(API response) into a CSI ListSnapshotsResponse_Entry
// func convertSnapshot(snap *tkmgo.Snapshot) *csi.ListSnapshotsResponse_Entry {
// 	return &csi.ListSnapshotsResponse_Entry{
// 		Snapshot: &csi.Snapshot{
// 			SnapshotId:     snap.Id,
// 			SourceVolumeId: snap.VolID,
// 			CreationTime:   snap.CreationTime,
// 			SizeBytes:      snap.SizeBytes,
// 			ReadyToUse:     snap.ReadyToUse,
// 		},
// 	}
// }

// ControllerExpandVolume allows for offline expansion of Volumes
func (d *Driver) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	volID := req.GetVolumeId()

	d.log.Info("Request: ControllerExpandVolume", "volume_id", volID)

	if volID == "" {
		return nil, status.Error(codes.InvalidArgument, "must provide a VolumeId to ControllerExpandVolume")
	}

	/*
		// Get the volume from the TKM API
		volume, err := d.TkmClient.GetVolume(volID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "ControllerExpandVolume could not retrieve existing volume: %v", err)
		}

		if req.CapacityRange == nil {
			return nil, status.Error(codes.InvalidArgument, "must provide a capacity range to ControllerExpandVolume")
		}
		bytes, err := getVolSizeInBytes(req.GetCapacityRange())
		if err != nil {
			return nil, err
		}
		desiredSize := bytes / BytesInGigabyte
		if (bytes % BytesInGigabyte) != 0 {
			desiredSize++
		}
		d.log.V(1).Info("Volume found",
			"current_size", volume.SizeGigabytes,
			"desired_size", desiredSize,
			"state", volume.Status)

		if volume.Status == "resizing" {
			return nil, status.Error(codes.Aborted, "volume is already being resized")
		}

		if desiredSize <= int64(volume.SizeGigabytes) {
			d.log.Info("Volume is currently larger that desired Size", "volume_id", volID)
			return &csi.ControllerExpandVolumeResponse{CapacityBytes: int64(volume.SizeGigabytes) * BytesInGigabyte, NodeExpansionRequired: true}, nil
		}

		if volume.Status != "available" {
			return nil, status.Error(codes.FailedPrecondition, "volume is not in an available state for OFFLINE expansion")
		}

		d.log.Info("Volume resize request sent", "size_gb", desiredSize, "volume_id", volID)
		_, err = d.TkmClient.ResizeVolume(volID, int(desiredSize))
		// Handles unexpected errors (e.g., API retry error or other upstream errors).
		if err != nil {
			d.log.Error(err, "Failed to resize volume in TKM API", "VolumeID", volID)
			return nil, status.Errorf(codes.Internal, "cannot resize volume %s: %s", volID, err.Error())
		}

		// Resizes can take a while, double the number of normal retries
		available, err := d.waitForVolumeStatus(volume, "available", TkmVolumeAvailableRetries*2)
		if err != nil {
			d.log.Error(err, "Unable to wait for volume availability in TKM API")
			return nil, err
		}

		if !available {
			return nil, status.Error(codes.Internal, "failed to wait for volume to be in an available state")
		}

		volume, _ = d.TkmClient.GetVolume(volID)
	*/
	d.log.Info("Volume successfully resized", "size_gb", 1 /*volume.SizeGigabytes*/, "volume_id", volID)
	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         int64(1 /*volume.SizeGigabytes*/) * BytesInGigabyte,
		NodeExpansionRequired: true,
	}, nil
}

// ControllerGetVolume is for optional Kubernetes health checking of volumes and we don't support it yet
func (d *Driver) ControllerGetVolume(context.Context, *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerModifyVolume modify volume
func (d *Driver) ControllerModifyVolume(_ context.Context, _ *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// func (d *Driver) mustEmbedUnimplementedControllerServer() {}
