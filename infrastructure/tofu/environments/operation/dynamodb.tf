resource "aws_dynamodb_table" "main" {
  name         = "benchmark_results"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
}
