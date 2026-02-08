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

variable "alert_email" {
  description = "Email address for billing alerts"
  type        = string
  default     = "ignacioruedaboada@gmail.com"
}

variable "billing_threshold" {
  description = "USD amount that triggers the billing alarm"
  type        = number
  default     = 4
}
