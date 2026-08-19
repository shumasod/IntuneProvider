terraform {
  required_providers {
    intune = {
      source  = "shumasod/intune"
      version = "~> 0.1"
    }
  }
}

provider "intune" {
  tenant_id     = var.tenant_id
  client_id     = var.client_id
  client_secret = var.client_secret
}

# Create a device group for Autopilot devices
resource "intune_device_group" "autopilot_devices" {
  display_name     = "Autopilot-Devices"
  description      = "Group for Windows Autopilot registered devices"
  security_enabled = true
  mail_enabled     = false
  mail_nickname    = "autopilot-devices"
}

# Create a Windows Autopilot deployment profile
resource "intune_windows_autopilot_profile" "corp_standard" {
  display_name        = "Corporate Standard Autopilot Profile"
  description         = "Standard OOBE experience for corporate Windows devices"
  language            = "os-default"
  device_name_template = "CORP-%SERIAL%"
  device_type         = "windowsPc"

  extract_hardware_hash   = false
  enable_white_glove      = false
  hide_privacy_settings   = true
  hide_eula               = true
  hide_change_account_options = false
}

variable "tenant_id" {
  description = "Azure AD Tenant ID"
  type        = string
  sensitive   = true
}

variable "client_id" {
  description = "Service Principal Client ID"
  type        = string
  sensitive   = true
}

variable "client_secret" {
  description = "Service Principal Client Secret"
  type        = string
  sensitive   = true
}

output "autopilot_profile_id" {
  value       = intune_windows_autopilot_profile.corp_standard.id
  description = "The ID of the created Autopilot profile"
}

output "device_group_id" {
  value       = intune_device_group.autopilot_devices.id
  description = "The ID of the Autopilot device group"
}
