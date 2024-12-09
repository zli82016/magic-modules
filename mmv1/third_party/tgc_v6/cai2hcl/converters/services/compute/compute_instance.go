package compute

import (
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl/converters/interfaces"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl/converters/resource"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl/hcl"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/caiasset"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-google-beta/google-beta/tpgresource"
	compute "google.golang.org/api/compute/v0.beta"
)

// ComputeInstanceAssetType is the CAI asset type name for compute instance.
const ComputeInstanceAssetType string = "compute.googleapis.com/Instance"

// ComputeInstanceSchemaName is the TF resource schema name for compute instance.
const ComputeInstanceSchemaName string = "google_compute_instance"

// ComputeInstanceConverter for compute instance resource.
type ComputeInstanceConverter struct {
	name   string
	schema map[string]*schema.Schema
}

// NewComputeInstanceConverter returns an HCL converter for compute instance.
func NewComputeInstanceConverter(provider *schema.Provider) interfaces.Converter {
	schema := provider.ResourcesMap[ComputeInstanceSchemaName].Schema

	return &ComputeInstanceConverter{
		name:   ComputeInstanceSchemaName,
		schema: schema,
	}
}

// Convert converts asset to HCL resource blocks.
func (c *ComputeInstanceConverter) Convert(asset *caiasset.Asset) ([]*hcl.HCLResourceBlock, error) {
	if asset == nil || asset.Resource == nil && asset.Resource.Data == nil {
		return nil, nil
	}
	var blocks []*hcl.HCLResourceBlock
	block, err := c.convertResourceData(asset)
	if err != nil {
		return nil, err
	}
	blocks = append(blocks, block)
	return blocks, nil
}

func (c *ComputeInstanceConverter) convertResourceData(asset *caiasset.Asset) (*hcl.HCLResourceBlock, error) {
	if asset == nil || asset.Resource == nil || asset.Resource.Data == nil {
		return nil, fmt.Errorf("asset resource data is nil")
	}

	var instance *compute.Instance
	if err := resource.DecodeJSON(asset.Resource.Data, &instance); err != nil {
		return nil, err
	}

	hclData := make(map[string]interface{})

	// bootDisks, scratchDisks, attachedDisks := convertDisks(instance.Disks)

	// hclData["metadata"] = convertMetadata(instance.Metadata)
	// hclData["partner_metadata"] = flattenPartnerMetadata(instance.PartnerMetadata)
	// hclData["can_ip_forward"] = instance.CanIpForward
	// hclData["machine_type"] = resource.ParseFieldValue(instance.MachineType, "machineTypes")
	// hclData["network_performance_config"] = flattenNetworkPerformanceConfig(instance.NetworkPerformanceConfig)
	// hclData["network_interface"] = flattenNetworkInterfaces(instance.NetworkInterfaces)
	// hclData["tags"] = instance.Tags.Items
	// hclData["labels"] = instance.Labels
	// hclData["boot_disk"] = bootDisks
	// hclData["resource_policies"] = instance.ResourcePolicies
	// hclData["service_account"] = flattenServiceAccounts(instance.ServiceAccounts)
	// hclData["attached_disk"] = attachedDisks
	// hclData["scratch_disk"] = scratchDisks
	// hclData["scheduling"] = convertScheduling(instance.Scheduling)
	// hclData["guest_accelerator"] = flattenGuestAccelerators(instance.GuestAccelerators)
	// hclData["shielded_instance_config"] = flattenShieldedVmConfig(instance.ShieldedInstanceConfig)
	// hclData["enable_display"] = flattenEnableDisplay(instance.DisplayDevice)
	// hclData["min_cpu_platform"] = instance.MinCpuPlatform
	// hclData["deletion_protection"] = instance.DeletionProtection
	// if instance.Zone == "" {
	// 	hclData["zone"] = resource.ParseFieldValue(asset.Name, "zones")
	// } else {
	// 	hclData["zone"] = resource.ParseFieldValue(instance.Zone, "zones")
	// }
	// hclData["name"] = instance.Name
	// hclData["description"] = instance.Description
	// hclData["hostname"] = instance.Hostname
	// hclData["confidential_instance_config"] = flattenConfidentialInstanceConfig(instance.ConfidentialInstanceConfig)
	// hclData["advanced_machine_features"] = flattenAdvancedMachineFeatures(instance.AdvancedMachineFeatures)
	// hclData["desired_status"] = instance.Status
	// hclData["reservation_affinity"] = flattenReservationAffinity(instance.ReservationAffinity)
	// hclData["key_revocation_action_type"] = instance.KeyRevocationActionType

	// hclData["metadata"] = flattenMetadataBeta(instance.Metadata)

	// If the existing state contains "metadata_startup_script" instead of "metadata.startup-script",
	// we should move the remote metadata.startup-script to metadata_startup_script to avoid
	// specifying it in two places.
	// if _, ok := d.GetOk("metadata_startup_script"); ok {
	// 	if err := d.Set("metadata_startup_script", md["startup-script"]); err != nil {
	// 		return fmt.Errorf("Error setting metadata_startup_script: %s", err)
	// 	}

	// 	delete(md, "startup-script")
	// }

	// if instance.PartnerMetadata != nil {
	// 	partnerMetadata, err := flattenPartnerMetadata(instance.PartnerMetadata)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("Error parsing partner metadata: %s", err)
	// 	}
	// 	hclData["partner_metadata"] = partnerMetadata
	// }

	hclData["can_ip_forward"] = instance.CanIpForward
	hclData["machine_type"] = tpgresource.GetResourceNameFromSelfLink(instance.MachineType)
	hclData["network_performance_config"] = flattenNetworkPerformanceConfig(instance.NetworkPerformanceConfig)
	// Set the networks
	// Use the first external IP found for the default connection info.
	networkInterfaces, _, _, _, err := flattenNetworkInterfaces(instance.NetworkInterfaces)
	if err != nil {
		return nil, err
	}
	hclData["network_interface"] = networkInterfaces

	// Fall back on internal ip if there is no external ip.  This makes sense in the situation where
	// terraform is being used on a cloud instance and can therefore access the instances it creates
	// via their internal ips.
	// sshIP := externalIP
	// if sshIP == "" {
	// 	sshIP = internalIP
	// }

	// Initialize the connection info
	// d.SetConnInfo(map[string]string{
	// 	"type": "ssh",
	// 	"host": sshIP,
	// })

	// Set the tags fingerprint if there is one.
	if instance.Tags != nil {
		hclData["tags"] = tpgresource.ConvertStringArrToInterface(instance.Tags.Items)
	}

	hclData["labels"] = instance.Labels

	attachedDisks := make([]map[string]interface{}, 0)
	scratchDisks := []map[string]interface{}{}
	for _, disk := range instance.Disks {
		if disk.Boot {
			hclData["boot_disk"] = flattenBootDisk(disk)
		} else if disk.Type == "SCRATCH" {
			scratchDisks = append(scratchDisks, flattenScratchDisk(disk))
		} else {
			di := map[string]interface{}{
				"source":      tpgresource.ConvertSelfLinkToV1(disk.Source),
				"device_name": disk.DeviceName,
				"mode":        disk.Mode,
			}
			if key := disk.DiskEncryptionKey; key != nil {
				if key.KmsKeyName != "" {
					// The response for crypto keys often includes the version of the key which needs to be removed
					// format: projects/<project>/locations/<region>/keyRings/<keyring>/cryptoKeys/<key>/cryptoKeyVersions/1
					di["kms_key_self_link"] = strings.Split(disk.DiskEncryptionKey.KmsKeyName, "/cryptoKeyVersions")[0]
				}
			}
			attachedDisks = append(attachedDisks, di)
		}
	}

	hclData["resource_policies"] = instance.ResourcePolicies

	// Remove nils from map in case there were disks in the config that were not present on read;
	// i.e. a disk was detached out of band
	ads := []map[string]interface{}{}
	for _, d := range attachedDisks {
		if d != nil {
			ads = append(ads, d)
		}
	}

	zone := tpgresource.GetResourceNameFromSelfLink(instance.Zone)

	hclData["service_account"] = flattenServiceAccounts(instance.ServiceAccounts)
	hclData["attached_disk"] = ads
	hclData["scratch_disk"] = scratchDisks
	hclData["scheduling"] = flattenScheduling(instance.Scheduling)
	hclData["guest_accelerator"] = flattenGuestAccelerators(instance.GuestAccelerators)
	hclData["shielded_instance_config"] = flattenShieldedVmConfig(instance.ShieldedInstanceConfig)
	hclData["enable_display"] = flattenEnableDisplay(instance.DisplayDevice)
	hclData["cpu_platform"] = instance.CpuPlatform
	hclData["min_cpu_platform"] = instance.MinCpuPlatform
	hclData["deletion_protection"] = instance.DeletionProtection
	hclData["self_link"] = tpgresource.ConvertSelfLinkToV1(instance.SelfLink)
	// hclData["project"] = project
	hclData["zone"] = zone
	hclData["name"] = instance.Name
	hclData["description"] = instance.Description
	hclData["hostname"] = instance.Hostname
	hclData["confidential_instance_config"] = flattenConfidentialInstanceConfig(instance.ConfidentialInstanceConfig)
	hclData["advanced_machine_features"] = flattenAdvancedMachineFeatures(instance.AdvancedMachineFeatures)
	hclData["desired_status"] = instance.Status
	hclData["reservation_affinity"] = flattenReservationAffinity(instance.ReservationAffinity)
	hclData["key_revocation_action_type"] = instance.KeyRevocationActionType

	ctyVal, err := resource.MapToCtyValWithSchema(hclData, c.schema)
	if err != nil {
		return nil, err
	}
	return &hcl.HCLResourceBlock{
		Labels: []string{c.name, instance.Name},
		Value:  ctyVal,
	}, nil

}

func flattenBootDisk(disk *compute.AttachedDisk) []map[string]interface{} {
	result := map[string]interface{}{
		"auto_delete": disk.AutoDelete,
		"device_name": disk.DeviceName,
		"mode":        disk.Mode,
		"source":      tpgresource.ConvertSelfLinkToV1(disk.Source),
		"interface":   disk.Interface,
	}

	// result["initialize_params"] = []map[string]interface{}{{
	// 	"type": tpgresource.GetResourceNameFromSelfLink(diskDetails.Type),
	// 	// If the config specifies a family name that doesn't match the image name, then
	// 	// the diff won't be properly suppressed. See DiffSuppressFunc for this field.
	// 	"image":                       diskDetails.SourceImage,
	// 	"size":                        diskDetails.SizeGb,
	// 	"labels":                      diskDetails.Labels,
	// 	"resource_manager_tags":       d.Get("boot_disk.0.initialize_params.0.resource_manager_tags"),
	// 	"resource_policies":           diskDetails.ResourcePolicies,
	// 	"provisioned_iops":            diskDetails.ProvisionedIops,
	// 	"provisioned_throughput":      diskDetails.ProvisionedThroughput,
	// 	"enable_confidential_compute": diskDetails.EnableConfidentialCompute,
	// 	"storage_pool":                tpgresource.GetResourceNameFromSelfLink(diskDetails.StoragePool),
	// }}

	if disk.DiskEncryptionKey != nil {
		if disk.DiskEncryptionKey.KmsKeyName != "" {
			// The response for crypto keys often includes the version of the key which needs to be removed
			// format: projects/<project>/locations/<region>/keyRings/<keyring>/cryptoKeys/<key>/cryptoKeyVersions/1
			result["kms_key_self_link"] = strings.Split(disk.DiskEncryptionKey.KmsKeyName, "/cryptoKeyVersions")[0]
		}
	}

	return []map[string]interface{}{result}
}

func flattenScratchDisk(disk *compute.AttachedDisk) map[string]interface{} {
	result := map[string]interface{}{
		"device_name": disk.DeviceName,
		"interface":   disk.Interface,
		"size":        disk.DiskSizeGb,
	}
	return result
}
