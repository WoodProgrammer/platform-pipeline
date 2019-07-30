provider "aws" {
  region = "eu-central-1"
}

resource "aws_s3_bucket" "my_bucket" {
  bucket = "devops-turkey-demo-bucket"
  acl    = "private"
  policy = "${file("policy.json")}"
  tags = {
    Name        = "DevOpsTurkey"
    Environment = "Dev"
  }
}

output "bucket_id" {
  value = "${aws_s3_bucket.my_bucket.id}"
}
