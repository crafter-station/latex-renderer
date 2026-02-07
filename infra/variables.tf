variable "region" {
  default = "us-east-1"
}

variable "api_key" {
  type      = string
  sensitive = true
}

variable "image_uri" {
  description = "Full ECR image URI (account.dkr.ecr.region.amazonaws.com/latex-renderer:tag)"
  type        = string
}
