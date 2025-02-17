resource "aws_sqs_queue" "main" {
  name                        = "bench-trigger-queue.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
}
