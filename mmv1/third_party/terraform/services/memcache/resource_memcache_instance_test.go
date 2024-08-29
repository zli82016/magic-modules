package memcache_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-google/google/acctest"
)

func TestAccMemcacheInstance_update(t *testing.T) {
	t.Parallel()

	prefix := fmt.Sprintf("%d", acctest.RandInt(t))
	name := fmt.Sprintf("tf-test-%s", prefix)
	network := acctest.BootstrapSharedServiceNetworkingConnection(t, "memcache-instance-update-1")

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckMemcacheInstanceDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemcacheInstance_update(prefix, name, network),
			},
			{
				ResourceName:            "google_memcache_instance.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reserved_ip_range_id"},
			},
			{
				Config: testAccMemcacheInstance_update2(prefix, name, network),
			},
			{
				ResourceName:      "google_memcache_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMemcacheInstance_tags(t *testing.T) {
	t.Skip()

	t.Parallel()

	instance := fmt.Sprintf("tf-test%s.org1.com", acctest.RandString(t, 5))
	context := map[string]interface{}{
		"instance": instance,
		"resource_name": "instance",
	}

	resourceName := acctest.Nprintf("google_memcache_instance.%{resource_name}", context)
	org := envvar.GetTestOrgFromEnv(t)

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckMemcacheInstanceDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemcacheInstanceTags(context, map[string]string{org + "/env": "test"}),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"labels", "location", "name", "terraform_labels", "zone"},
			},
			{
				Config: testAccMemcacheInstanceTags_allowDestroy(context, map[string]string{org + "/env": "test"}),
			},
		},
	})
}

func testAccMemcacheInstance_update(prefix, name, network string) string {
	return fmt.Sprintf(`
resource "google_memcache_instance" "test" {
  name = "%s"
  region = "us-central1"
  authorized_network = data.google_compute_network.memcache_network.id

  node_config {
    cpu_count      = 1
    memory_size_mb = 1024
  }
  node_count = 1

  memcache_parameters {
    params = {
      "listen-backlog" = "2048"
      "max-item-size" = "8388608"
    }
  }
  reserved_ip_range_id = ["tf-bootstrap-addr-memcache-instance-update-1"]
}

data "google_compute_network" "memcache_network" {
  name = "%s"
}
`, name, network)
}

func testAccMemcacheInstance_update2(prefix, name, network string) string {
	return fmt.Sprintf(`
resource "google_memcache_instance" "test" {
  name = "%s"
  region = "us-central1"
  authorized_network = data.google_compute_network.memcache_network.id

  node_config {
    cpu_count      = 1
    memory_size_mb = 1024
  }
  node_count = 2

  memcache_parameters {
    params = {
      "listen-backlog" = "2048"
      "max-item-size" = "8388608"
    }
  }

  memcache_version = "MEMCACHE_1_6_15"
}

data "google_compute_network" "memcache_network" {
  name = "%s"
}
`, name, network)
}

func testAccMemcacheInstanceTags(context map[string]interface{}, tags map[string]string) string {

	r := acctest.Nprintf(`
	resource "google_memcache_instance" "%{resource_name}" {
	  name = "tf-instance-%s"
	  authorized_network = google_service_networking_connection.private_service_connection.network
          node_config {
            cpu_count      = 1
            memory_size_mb = 1024
          }
         node_count = 1
         memcache_version = "MEMCACHE_1_5"
         labels = {
           "key1" = "value1"
           "key2" = "value2"
         }
	 tags = {`, context)

	l := ""
	for key, value := range tags {
		l += fmt.Sprintf("%q = %q\n", key, value)
	}

	l += fmt.Sprintf("}\n}")
	return r + l
}

func testAccMemcacheInstanceTags_allowDestroy(context map[string]interface{}, tags map[string]string) string {

	r := acctest.Nprintf(`
	resource "google_memcache_instance" "%{resource_name}" {
	  name = "tf-instance-%s"
	  authorized_network = google_service_networking_connection.private_service_connection.network
          node_config {
            cpu_count      = 1
            memory_size_mb = 1024
          }
         node_count = 1
         memcache_version = "MEMCACHE_1_5"
          labels = {
            "key1" = "value1"
            "key2" = "value2"
          }
	  tags = {`, context)

	l := ""
	for key, value := range tags {
		l += fmt.Sprintf("%q = %q\n", key, value)
	}

	l += fmt.Sprintf("}\n}")
	return r + l
}

