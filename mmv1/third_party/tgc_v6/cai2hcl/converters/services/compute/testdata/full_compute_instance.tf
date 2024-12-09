data "google_project" "project" {
  provider = google-beta
}

data "google_compute_image" "my_image" {
  family  = "debian-11"
  project = "debian-cloud"
}

resource "google_compute_network" "test_network_1"{
	name = "tf-test-compute-network-1"
	auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "test_subnetwork_1" {
	name             = "tf-test-compute-subnet-1"
	ip_cidr_range    = "10.0.0.0/16"
	region           = "us-central1"
	network          = google_compute_network.test_network_1.id
  stack_type       = "IPV4_IPV6"
  ipv6_access_type = "EXTERNAL"

  secondary_ip_range {
    range_name    = "inst-test-tertiary"
    ip_cidr_range = "10.1.0.0/16"
  }
}

resource "google_compute_address" "ipv6_address_1" {
  region             = "us-central1"
  name               = "tf-test-addr-ipv62-1"
  address_type       = "EXTERNAL"
  ip_version         = "IPV6"
  network_tier       = "PREMIUM"
  ipv6_endpoint_type = "VM"
  subnetwork         = google_compute_subnetwork.test_subnetwork_1.name
}

resource "google_compute_address" "normal_address_1" {
  region   = "us-central1"
  name     = "tf-test-addr-normal-1"
}

resource "google_compute_network_attachment" "test_network_attachment_1" {
	name = "tf-test-compute-network-attachment-1"
  region = "us-central1"
  description = "network attachment description"
  connection_preference = "ACCEPT_AUTOMATIC"

  subnetworks = [
    google_compute_subnetwork.test_subnetwork_1.self_link
  ]
}

resource "google_compute_region_security_policy" "policyforinstance_1" {
  region      = "europe-west1"
  name        = "tf-test-compute-region-security-policy-1"
  description = "region security policy to set to instance"
  type        = "CLOUD_ARMOR_NETWORK"
}

resource "google_compute_disk" "foobar1" {
  name = "tf-test-disk-1"
  size = 10
  type = "pd-ssd"
  zone = "us-central1-a"
}

resource "google_compute_disk" "foobar2" {
  name = "tf-test-disk-1"
  size = 10
  type = "pd-ssd"
  zone = "us-central1-a"
}

resource "google_compute_disk" "foobar3" {
  name = "tf-test-disk-1"
  size = 10
  type = "pd-ssd"
  zone = "us-central1-a"
}

resource "google_compute_disk" "foobar4" {
  name = "tf-test-disk-1"
  size = 10
  type = "pd-ssd"
  zone = "us-central1-a"
}

resource "google_compute_resource_policy" "foo1" {
  name   = "tf-test-policy-1"
  region = "us-central1"
  group_placement_policy {
    vm_count = 2
    collocation = "COLLOCATED"
  }
}

resource "google_kms_key_ring" "key_ring_1" {
  project  = data.google_project.project.project_id
  name     = "tf-test-kma-key-1"
  location = "us-central1"
}

resource "google_kms_crypto_key" "crypto_key_1" {
  name            = "tf-test-kms-crypto-key-1"
  key_ring        = google_kms_key_ring.key_ring_1.id
  rotation_period = "100000s"
}

resource "google_kms_crypto_key" "crypto_key_2" {
  name            = "tf-test-kms-crypto-key-2"
  key_ring        = google_kms_key_ring.key_ring_1.id
  rotation_period = "100000s"
}

resource "google_compute_instance" "test1" {
  attached_disk {
    device_name       = "test-device_name-1"
    kms_key_self_link = google_kms_crypto_key.crypto_key_1.name
    mode              = "READ_ONLY"
    source            = google_compute_disk.foobar1.self_link
  }

  attached_disk {
    device_name = "test-device_name-2"
    mode        = "READ_WRITE"
    source      = google_compute_disk.foobar2.self_link
  }

  boot_disk {
    auto_delete       = true
    device_name       = "test-device_name"
    interface         = "SCSI"
    kms_key_self_link = google_kms_crypto_key.crypto_key_2.name
    mode              = "READ_WRITE"
    source            = google_compute_disk.foobar3.self_link
  }

  can_ip_forward      = true
  deletion_protection = true
  description         = "test-description"
  enable_display      = true

  guest_accelerator {
    count = 1
    type  = "nvidia-tesla-t4"
  }

  hostname = "test1.test"

  labels = {
    label_foo1 = "label-bar1"
  }

  machine_type     = "n1-standard-1"
  name             = "test1"

  key_revocation_action_type = "STOP"

  network_interface {
    access_config {
      network_tier = "STANDARD"
      nat_ip = google_compute_address.normal_address_1.address
    }

    alias_ip_range {
      ip_cidr_range         = "10.1.1.0/24"
      subnetwork_range_name = "inst-test-tertiary"
    }

    network            = google_compute_network.test_network_1.self_link
    network_ip         = "10.3.0.3"
    queue_count        = 0
    nic_type           = "VIRTIO_NET"
    security_policy    = google_compute_region_security_policy.policyforinstance_1.self_link
  }

  network_interface {
    queue_count                 = 0
    subnetwork                  = google_compute_subnetwork.test_subnetwork_1.self_link
    stack_type                  = "IPV4_IPV6"
    internal_ipv6_prefix_length = "96"
    ipv6_address                = google_compute_address.normal_address_1.address
  }

  network_interface {
    ipv6_access_config {
      external_ipv6               = google_compute_address.ipv6_address_1.address
      external_ipv6_prefix_length = "96"
      network_tier                = "PREMIUM"
      public_ptr_domain_name      = "tf-test-1.gcp.tfacc.hashicorptest.com"
      name                        = "External IPv6"
    }

    network_attachment = google_compute_network_attachment.test_network_attachment_1.self_link
    stack_type  = "IPV4_IPV6"
    queue_count = 0
  }

  network_performance_config {
    total_egress_bandwidth_tier = "DEFAULT"
  }

/*
  scheduling {
    automatic_restart   = false
    on_host_maintenance = "MIGRATE"
    preemptible         = false
    node_affinities {
      key      = "tfacc"
      operator = "IN"
      values   = ["test"]
    }
    node_affinities {
      key      = "tfacc"
      operator = "NOT_IN"
      values   = ["not_here"]
    }
    provisioning_model          = "STANDARD"
    instance_termination_action = "STOP"
    max_run_duration {
      nanos = 123
      seconds = 60
    }
	  on_instance_stop_action {
		  discard_local_ssd = true
	  }
    local_ssd_recovery_timeout {
      nanos = 0
      seconds = 3600
    }
  }
*/

  scratch_disk {
    interface   = "SCSI"
    device_name = "test-device_name-3"
    size        = 375
  }

  scratch_disk {
    interface = "SCSI"
  }

  service_account {
    email  = data.google_compute_default_service_account.default.email
    scopes = ["https://www.googleapis.com/auth/cloud-platform"]
  }

  advanced_machine_features {
    enable_nested_virtualization = true
    visible_core_count           = 1
  }

  tags              = ["bar", "foo"]
  zone              = "us-central1-a"
  desired_status    = "RUNNING"
  resource_policies = [google_compute_resource_policy.foo1.self_link]

  reservation_affinity {
    type = "SPECIFIC_RESERVATION"

	  specific_reservation {
		  key    = "compute.googleapis.com/reservation-name"
		  values = ["%[1]s"]
	  }
  }
}
