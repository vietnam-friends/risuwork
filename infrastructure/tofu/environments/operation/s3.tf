resource "aws_s3_bucket" "static" {
  bucket = "risuwork-static"
}

resource "aws_s3_bucket_policy" "static" {
  bucket = aws_s3_bucket.static.id
  policy = templatefile("./files/policy/static_oac.json", {
    principal_identifier = aws_cloudfront_origin_access_identity.main.iam_arn,
    bucket_arn           = aws_s3_bucket.static.arn,
  })
}
