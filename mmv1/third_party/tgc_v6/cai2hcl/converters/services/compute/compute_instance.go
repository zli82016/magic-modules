package compute

import (
	"fmt"

	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl/converters/interfaces"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl/converters/resource"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/cai2hcl/hcl"
	"github.com/GoogleCloudPlatform/terraform-google-conversion/v6/pkg/caiasset"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/api/compute/v1"
)

// ComputeInstanceAssetType is the CAI asset type name for compute assetResourceData[
const ComputeInstanceAssetType string = "compute.googleapis.com/Instance"

// ComputeInstanceSchemaName is the TF resource schema name for compute assetResourceData[
const ComputeInstanceSchemaName string = "google_compute_instance"

// ComputeInstanceConverter for compute instance resource.
type ComputeInstanceConverter struct {
	name   string
	schema map[string]*schema.Schema
}

// NewComputeInstanceConverter returns an HCL converter for compute assetResourceData[
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

	assetResourceData := asset.Resource.Data

	bootDisks, scratchDisks, attachedDisks := convertDisks(assetResourceData["disks"])

	hclData := make(map[string]interface{})
	hclData["can_ip_forward"] = assetResourceData["canIpForward"]
	hclData["description"] = assetResourceData["description"]
	hclData["boot_disk"] = bootDisks
	hclData["scratch_disk"] = scratchDisks
	hclData["attached_disk"] = attachedDisks
	hclData["machine_type"] = resource.ParseFieldValue(assetResourceData["machineType"], "machineTypes")
	hclData["name"] = assetResourceData["name"]
	hclData["network_interface"] = flattenNetworkInterfaces(assetResourceData["networkInterfaces"])
	hclData["tags"] = assetResourceData["tags.items"]
	hclData["tags_fingerprint"] = assetResourceData["tags.fingerprint"]
	hclData["labels"] = assetResourceData["labels"]
	hclData["service_account"] = flattenServiceAccounts(assetResourceData["serviceAccounts"])
	hclData["guest_accelerator"] = flattenGuestAccelerators(assetResourceData["guestAccelerators"])
	hclData["min_cpu_platform"] = assetResourceData["minCpuPlatform"]
	hclData["scheduling"] = convertScheduling(assetResourceData["scheduling"])
	hclData["deletion_protection"] = assetResourceData["deletionProtection"]
	hclData["hostname"] = assetResourceData["hostname"]
	hclData["shielded_instance_config"] = flattenShieldedVmConfig(assetResourceData["shieldedInstanceConfig"])
	hclData["enable_display"] = flattenEnableDisplay(assetResourceData["displayDevice"])
	hclData["metadata_fingerprint"] = assetResourceData["metadata.fingerprint"]
	hclData["metadata"] = convertMetadata(assetResourceData["metadata")

	if assetResourceData[Zone == "" {
		hclData["zone"] = resource.ParseFieldValue(asset["name"], "zones")
	} else {
		hclData["zone"] = resource.ParseFieldValue(assetResourceData["zone"], "zones")
	}

	ctyVal, err := resource.MapToCtyValWithSchema(hclData, c.schema)
	if err != nil {
		return nil, err
	}
	return &hcl.HCLResourceBlock{
		Labels: []string{c.name, assetResourceData["name"]},
		Value:  ctyVal,
	}, nil

}

func convertDisks(disks []*compute.AttachedDisk) (bootDisks []map[string]interface{}, scratchDisks []map[string]interface{}, attachedDisks []map[string]interface{}) {
	for _, disk := range disks {
		if disk.Boot {
			bootDisks = append(bootDisks, convertBootDisk(disk))
			continue
		}
		if disk.Type == "SCRATCH" {
			scratchDisks = append(scratchDisks, flattenScratchDisk(disk))
			continue
		}
		attachedDisks = append(attachedDisks, convertAttachedDisk(disk))
	}
	return
}

func convertBootDisk(disk *compute.AttachedDisk) map[string]interface{} {
	data := map[string]interface{}{
		"auto_delete": disk.AutoDelete,
		"device_name": disk.DeviceName,
		"source":      disk.Source,
		"mode":        disk.Mode,
	}
	if disk.InitializeParams != nil {
		data["initialize_params"] = []map[string]interface{}{
			{
				"size":   disk.InitializeParams.DiskSizeGb,
				"type":   resource.ParseFieldValue(disk.InitializeParams.DiskType, "diskTypes"),
				"image":  disk.InitializeParams.SourceImage,
				"labels": disk.InitializeParams.Labels,
			},
		}
	}
	if disk.DiskEncryptionKey != nil {
		if disk.DiskEncryptionKey.RawKey != "" {
			data["disk_encryption_key_raw"] = disk.DiskEncryptionKey.RawKey
		}
		if disk.DiskEncryptionKey.Sha256 != "" {
			data["disk_encryption_key_sha256"] = disk.DiskEncryptionKey.Sha256
		}
		if disk.DiskEncryptionKey.KmsKeyName != "" {
			data["kms_key_self_link"] = disk.DiskEncryptionKey.KmsKeyName
		}
	}

	return data
}

func convertAttachedDisk(disk *compute.AttachedDisk) map[string]interface{} {
	data := map[string]interface{}{
		"source":      disk.Source,
		"mode":        disk.Mode,
		"device_name": disk.DeviceName,
	}
	if disk.DiskEncryptionKey != nil {
		if disk.DiskEncryptionKey.RawKey != "" {
			data["disk_encryption_key_raw"] = disk.DiskEncryptionKey.RawKey
		}
		if disk.DiskEncryptionKey.Sha256 != "" {
			data["disk_encryption_key_sha256"] = disk.DiskEncryptionKey.Sha256
		}
		if disk.DiskEncryptionKey.KmsKeyName != "" {
			data["kms_key_self_link"] = disk.DiskEncryptionKey.KmsKeyName
		}
	}
	return data
}

func convertScheduling(sched *compute.Scheduling) []map[string]interface{} {
	data := map[string]interface{}{
		"automatic_restart":   sched.AutomaticRestart,
		"preemptible":         sched.Preemptible,
		"on_host_maintenance": sched.OnHostMaintenance,
		// node_affinities are not converted into cai
		"node_affinities": convertSchedulingNodeAffinity(sched.NodeAffinities),
	}
	if sched.MinNodeCpus > 0 {
		data["min_node_cpus"] = sched.MinNodeCpus
	}
	if len(sched.ProvisioningModel) > 0 {
		data["provisioning_model"] = sched.ProvisioningModel
	}
	return []map[string]interface{}{data}
}

func convertSchedulingNodeAffinity(items []*compute.SchedulingNodeAffinity) []map[string]interface{} {
	arr := make([]map[string]interface{}, len(items))
	for ix, item := range items {
		arr[ix] = make(map[string]interface{})
		arr[ix]["key"] = item.Key
		arr[ix]["operator"] = item.Operator
		arr[ix]["values"] = item.Values
	}
	return arr
}

func convertMetadata(metadata *compute.Metadata) map[string]interface{} {
	md := flattenMetadata(metadata)

	// If the existing state contains "metadata_startup_script" instead of "metadata.startup-script",
	// we should move the remote metadata.startup-script to metadata_startup_script to avoid
	// specifying it in two places.
	if _, ok := md["metadata_startup_script"]; ok {
		md["metadata_startup_script"] = md["startup-script"]
		delete(md, "startup-script")
	}

	return md
}

func flattenMetadata(metadata *compute.Metadata) map[string]interface{} {
	metadataMap := make(map[string]interface{})
	for _, item := range metadata.Items {
		metadataMap[item.Key] = *item.Value
	}
	return metadataMap
}

func flattenGuestAccelerators(accelerators []*compute.AcceleratorConfig) []map[string]interface{} {
	acceleratorsSchema := make([]map[string]interface{}, len(accelerators))
	for i, accelerator := range accelerators {
		acceleratorsSchema[i] = map[string]interface{}{
			"count": accelerator.AcceleratorCount,
			"type":  accelerator.AcceleratorType,
		}
	}
	return acceleratorsSchema
}

func flattenShieldedVmConfig(shieldedVmConfig *compute.ShieldedInstanceConfig) []map[string]bool {
	if shieldedVmConfig == nil {
		return nil
	}

	return []map[string]bool{{
		"enable_secure_boot":          shieldedVmConfig.EnableSecureBoot,
		"enable_vtpm":                 shieldedVmConfig.EnableVtpm,
		"enable_integrity_monitoring": shieldedVmConfig.EnableIntegrityMonitoring,
	}}
}

func flattenEnableDisplay(displayDevice *compute.DisplayDevice) interface{} {
	if displayDevice == nil {
		return nil
	}

	return displayDevice.EnableDisplay
}

func flattenServiceAccounts(serviceAccounts []*compute.ServiceAccount) []map[string]interface{} {
	result := make([]map[string]interface{}, len(serviceAccounts))
	for i, serviceAccount := range serviceAccounts {
		result[i] = map[string]interface{}{
			"email":  serviceAccount.Email,
			"scopes": serviceAccount.Scopes,
		}
	}
	return result
}

func flattenNetworkInterfaces(networkInterfaces []*compute.NetworkInterface) []map[string]interface{} {
	flattened := make([]map[string]interface{}, len(networkInterfaces))
	for i, iface := range networkInterfaces {
		var ac []map[string]interface{}
		ac, _ = flattenAccessConfigs(iface.AccessConfigs)

		flattened[i] = map[string]interface{}{
			"network_ip":         iface.NetworkIP,
			"network":            iface.Network,
			"subnetwork":         iface.Subnetwork,
			"access_config":      ac,
			"alias_ip_range":     flattenAliasIpRange(iface.AliasIpRanges),
			"nic_type":           iface.NicType,
			"stack_type":         iface.StackType,
			"ipv6_access_config": flattenIpv6AccessConfigs(iface.Ipv6AccessConfigs),
			"queue_count":        iface.QueueCount,
		}
		// Instance template interfaces never have names, so they're absent
		// in the instance template network_interface schema. We want to use the
		// same flattening code for both resource types, so we avoid trying to
		// set the name field when it's not set at the GCE end.
		if iface.Name != "" {
			flattened[i]["name"] = iface.Name
		}
	}
	return flattened
}

func flattenIpv6AccessConfigs(ipv6AccessConfigs []*compute.AccessConfig) []map[string]interface{} {
	flattened := make([]map[string]interface{}, len(ipv6AccessConfigs))
	for i, ac := range ipv6AccessConfigs {
		flattened[i] = map[string]interface{}{
			"network_tier": ac.NetworkTier,
		}
		flattened[i]["public_ptr_domain_name"] = ac.PublicPtrDomainName
		flattened[i]["external_ipv6"] = ac.ExternalIpv6
		flattened[i]["external_ipv6_prefix_length"] = ac.ExternalIpv6PrefixLength
	}
	return flattened
}

func flattenAccessConfigs(accessConfigs []*compute.AccessConfig) ([]map[string]interface{}, string) {
	flattened := make([]map[string]interface{}, len(accessConfigs))
	natIP := ""
	for i, ac := range accessConfigs {
		flattened[i] = map[string]interface{}{
			"nat_ip":       ac.NatIP,
			"network_tier": ac.NetworkTier,
		}
		if ac.SetPublicPtr {
			flattened[i]["public_ptr_domain_name"] = ac.PublicPtrDomainName
		}
		if natIP == "" {
			natIP = ac.NatIP
		}
	}
	return flattened, natIP
}

func flattenAliasIpRange(ranges []*compute.AliasIpRange) []map[string]interface{} {
	rangesSchema := make([]map[string]interface{}, 0, len(ranges))
	for _, ipRange := range ranges {
		rangesSchema = append(rangesSchema, map[string]interface{}{
			"ip_cidr_range":         ipRange.IpCidrRange,
			"subnetwork_range_name": ipRange.SubnetworkRangeName,
		})
	}
	return rangesSchema
}

func flattenScratchDisk(disk *compute.AttachedDisk) map[string]interface{} {
	result := map[string]interface{}{
		"interface": disk.Interface,
	}
	return result
}
