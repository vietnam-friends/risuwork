// ベンチ起動のリクエストを受け付けるLambda
resource "aws_iam_role" "trigger_bench" {
  name = "trigger-bench-lambda-role"
  assume_role_policy = templatefile("../../files/policy/assume_role.json", {
    service_name = "lambda"
  })
  managed_policy_arns = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
    "arn:aws:iam::aws:policy/AmazonSQSFullAccess",
    "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess",
  ]
}

data "archive_file" "trigger_bench" {
  type        = "zip"
  source_dir  = "files/trigger_bench_function"
  output_path = ".archive/trigger_bench_function.zip"
}

resource "aws_lambda_function" "trigger_bench" {
  function_name    = "trigger-bench"
  runtime          = "nodejs20.x"
  filename         = data.archive_file.trigger_bench.output_path
  source_code_hash = data.archive_file.trigger_bench.output_base64sha256
  handler          = "index.handler"
  role             = aws_iam_role.trigger_bench.arn
}

// TODO: 認証考慮
resource "aws_lambda_function_url" "trigger_bench" {
  function_name      = aws_lambda_function.trigger_bench.function_name
  authorization_type = "NONE"
}

// ベンチ実行完了後のSNSトピックをサブスクライブするLambda
resource "aws_iam_role" "sub_bench_score" {
  name = "sub-bench-score-lambda-role"
  assume_role_policy = templatefile("../../files/policy/assume_role.json", {
    service_name = "lambda"
  })
  managed_policy_arns = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
    "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess",
  ]
}

data "archive_file" "sub_bench_score" {
  type        = "zip"
  source_dir  = "files/sub_bench_score_function"
  output_path = ".archive/sub_bench_score_function.zip"
}

resource "aws_lambda_function" "sub_bench_score" {
  function_name    = "sub-bench-score"
  runtime          = "nodejs20.x"
  filename         = data.archive_file.sub_bench_score.output_path
  source_code_hash = data.archive_file.sub_bench_score.output_base64sha256
  handler          = "index.handler"
  role             = aws_iam_role.sub_bench_score.arn
  timeout          = 10

  environment {
    variables = {
      SLACK_TOKEN = local.secrets.slack_token
    }
  }
}

resource "aws_lambda_permission" "sub_bench_score" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.sub_bench_score.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.main.arn
}

resource "aws_sns_topic_subscription" "sub_bench_score" {
  topic_arn = aws_sns_topic.main.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.sub_bench_score.arn
}
