
locals {
  sa_roles = [
    "roles/datastore.user"
  ]
}

resource "google_project_iam_member" "service_iam_roles" {
  for_each = toset(local.sa_roles)
  project = local.gcloud_project_id
  role    = each.value
  member  = "serviceAccount:${google_service_account.service.email}"
}


resource "google_service_account" "service" {
  account_id   = "${local.service_name}-${var.env}"
  display_name = "${local.service_name}-${var.env}"
}

# resource "google_service_account_key" "service" {
#   service_account_id = google_service_account.service.id
# }

# Allow github action runner to impersonate the service account (to run the tests)
resource "google_service_account_iam_member" "pacioli_wif_repo" {
  service_account_id = google_service_account.service.name
  role               = "roles/iam.workloadIdentityUser"
  member = "principalSet://iam.googleapis.com/projects/${data.google_project.current.number}/locations/global/workloadIdentityPools/cicd-pipeline/attribute.repository/Screevosrl/talk-to-me"
}
