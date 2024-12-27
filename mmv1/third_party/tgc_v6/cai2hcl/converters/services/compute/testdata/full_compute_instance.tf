resource "google_compute_instance" "test1" {
  advanced_machine_features {
    enable_nested_virtualization = true
    enable_uefi_networking       = false
    threads_per_core             = 0
    visible_core_count           = 0
  }

  attached_disk {
    device_name = "test-device_name-1"
    mode        = "READ_ONLY"
    source      = "https://www.googleapis.com/compute/v1/projects/terraform-dev-zhenhuali/zones/us-central1-a/disks/tf-test-disk-1"
  }

  attached_disk {
    device_name = "test-device_name-2"
    mode        = "READ_WRITE"
    source      = "https://www.googleapis.com/compute/v1/projects/terraform-dev-zhenhuali/zones/us-central1-a/disks/tf-test-disk-2"
  }

  boot_disk {
    auto_delete = true
    device_name = "test-device_name"
    interface   = "SCSI"
    mode        = "READ_WRITE"
    source      = "https://www.googleapis.com/compute/v1/projects/terraform-dev-zhenhuali/zones/us-central1-a/disks/tf-test-disk-3"
  }

  can_ip_forward             = true
  deletion_protection        = true
  description                = "test-description"
  desired_status             = "RUNNING"
  enable_display             = true
  hostname                   = "test1.test"
  key_revocation_action_type = "STOP_ON_KEY_REVOCATION"

  labels = {
    label_foo1 = "label-bar1"
  }

  machine_type = "n1-standard-1"
  name         = "test1"

  network_interface {
    network     = "https://www.googleapis.com/compute/v1/projects/terraform-dev-zhenhuali/global/networks/default"
    network_ip  = "10.128.0.9"
    queue_count = 0
    stack_type  = "IPV4_ONLY"
    subnetwork  = "https://www.googleapis.com/compute/v1/projects/terraform-dev-zhenhuali/regions/us-central1/subnetworks/default"
  }

  network_performance_config {
    total_egress_bandwidth_tier = "DEFAULT"
  }

  scheduling {
    automatic_restart   = true
    min_node_cpus       = 0
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    provisioning_model  = "STANDARD"
  }

  scratch_disk {
    device_name = "test-device_name-3"
    interface   = "SCSI"
    size        = 375
  }

  scratch_disk {
    device_name = "local-ssd-1"
    interface   = "SCSI"
    size        = 375
  }

  service_account {
    email  = "316788426028-compute@developer.gserviceaccount.com"
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  tags = ["bar", "foo"]
  zone = "us-central1-a"
}
