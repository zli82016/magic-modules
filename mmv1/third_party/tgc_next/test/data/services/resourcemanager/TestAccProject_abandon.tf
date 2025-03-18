resource "google_project" "tf-test-5075051301211104916" {
  project_id = "tf-test-tgc"
  name       = "tf-test-tgc"
  org_id = "529579013760"
  deletion_policy = "ABANDON"
}
