terraform {
  required_providers {
    intune = {
      source  = "shumasod/intune"
      version = "~> 0.1"
    }
  }
  required_version = ">= 1.5.0"
}

# Authentication via provider block attributes.
# Alternatively, set INTUNE_TENANT_ID, INTUNE_CLIENT_ID, INTUNE_CLIENT_SECRET
# environment variables and omit this block entirely.
provider "intune" {
  tenant_id     = var.tenant_id
  client_id     = var.client_id
  client_secret = var.client_secret
}

variable "tenant_id" {
  description = "Azure AD Tenant ID"
  type        = string
  sensitive   = false
}

variable "client_id" {
  description = "Service principal (application) client ID"
  type        = string
  sensitive   = false
}

variable "client_secret" {
  description = "Service principal client secret"
  type        = string
  sensitive   = true
}
