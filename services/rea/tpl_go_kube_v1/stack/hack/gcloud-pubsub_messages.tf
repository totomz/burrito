# data "google_pubsub_topic" "kargo_conversation_messages" {
#   name = "messages"
# }
# 
# resource "google_pubsub_subscription" "messages-exportdb" {
#   name                         = "${local.service_name}-conversation_messages"
#   topic                        = data.google_pubsub_topic.kargo_conversation_messages.name
#   ack_deadline_seconds         = 300
#   message_retention_duration   = "604800s"
#   enable_message_ordering      = true
#   enable_exactly_once_delivery = true
# }
# 
# resource "google_pubsub_subscription_iam_member" "allow_subscribe_kargo_messages-db" {
#   project      = google_pubsub_subscription.messages-exportdb.project
#   subscription = google_pubsub_subscription.messages-exportdb.name
#   role         = "roles/pubsub.subscriber"
#   member       = "serviceAccount:${google_service_account.service.email}"
# }
