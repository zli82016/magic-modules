package iap_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-provider-google/google/acctest"
)

func TestAccIapSettings_update(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"random_suffix": acctest.RandString(t, 10),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		CheckDestroy:             testAccCheckIapSettingsDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccIapSettings_basic(context),
			},
			{
				ResourceName:      "google_iap_settings.iap_settings",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccIapSettings_basic(context map[string]interface{}) string {
	return acctest.Nprintf(`

	data "google_project" "project" {
	}
	
	resource "google_compute_region_backend_service" "default" {
  	  name                            = "tf-test-iap-settings-tf%{random_suffix}"
  	  region                          = "us-central1"
  	  health_checks                   = [google_compute_health_check.default.id]
  	  connection_draining_timeout_sec = 10
  	  session_affinity                = "CLIENT_IP"
	}

	resource "google_compute_health_check" "default" {
  	  name               = "tf-test-iap-bs-health-check%{random_suffix}"
  	  check_interval_sec = 1
  	  timeout_sec        = 1
	  tcp_health_check {
  	    port = "80"
	  }
	}

	resource "google_iap_settings" "iap_settings" {
	  name = "projects/${data.google_project.project.number}/iap_web/compute-us-central1/services/${google_compute_region_backend_service.default.name}"
	  access_settings {
	    identity_sources = ["IDENTITY_SOURCE_UNSPECIFIED"]
	    cors_settings {
	      allow_http_options = true
	    }
	    reauth_settings {
	      method = "SECURE_KEY"
	      max_age = "305s"
	      policy_type = "MINIMUM"
	    }
	    oauth_settings {
	      login_hint = "test"
	    }
	    workforce_identity_settings {
	      workforce_pools = ["wif-pool"]
	      oauth2 {
	        client_id = "test-client-id"
	        client_secret = "test-client-secret"
	      }
	    }    
	  }
	  application_settings {
	    csm_settings {
	      rctoken_aud = "test-aud-set"
	    }
	    access_denied_page_settings {
	      access_denied_page_uri = "test-uri"
	      generate_troubleshooting_uri = true
	      remediation_token_generation_enabled = false
	    }
	    attribute_propagation_settings {
	      output_credentials = ["HEADER"]
	      enable = false
	    }
	  }
	}
`, context)
}
