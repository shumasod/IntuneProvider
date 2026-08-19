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

# iOS App Protection Policy
resource "intune_app_protection_policy_ios" "corp_ios" {
  display_name = "Corporate iOS MAM Policy"
  description  = "Protects corporate data on iOS devices"

  period_offline_before_access_check = "PT12H"
  period_online_before_access_check  = "PT5M"
  allowed_inbound_data_transfer_sources = "managedApps"
  allowed_outbound_data_transfer_destinations = "managedApps"
  organizational_credentials_required = false
  allowed_outbound_clipboard_sharing_level = "managedAppsWithPasteIn"
  data_backup_blocked = true
  device_compliance_required = true
  managed_browser_to_open_links_required = false
  save_as_blocked = true
  periodOfflineBeforeWipeIsEnforced = "P90D"
  pin_required = true
  maximum_pin_retries = 5
  simple_pin_blocked = true
  minimum_pin_length = 6
  pin_character_set = "numeric"
  period_before_pin_reset = "PT0S"
  allowed_data_storage_locations = ["oneDriveForBusiness", "sharePoint"]
  contact_sync_blocked = false
  print_blocked = false
  fingerprint_blocked = false
  disable_app_pin_if_device_pin_is_set = false
  managed_apps = []
}

# Android App Protection Policy
resource "intune_app_protection_policy_android" "corp_android" {
  display_name = "Corporate Android MAM Policy"
  description  = "Protects corporate data on Android devices"

  period_offline_before_access_check = "PT12H"
  period_online_before_access_check  = "PT5M"
  allowed_inbound_data_transfer_sources = "managedApps"
  allowed_outbound_data_transfer_destinations = "managedApps"
  organizational_credentials_required = false
  allowed_outbound_clipboard_sharing_level = "managedAppsWithPasteIn"
  data_backup_blocked = true
  device_compliance_required = true
  managed_browser_to_open_links_required = false
  save_as_blocked = true
  periodOfflineBeforeWipeIsEnforced = "P90D"
  pin_required = true
  maximum_pin_retries = 5
  simple_pin_blocked = true
  minimum_pin_length = 6
  pin_character_set = "numeric"
  period_before_pin_reset = "PT0S"
  screen_capture_blocked = true
  disable_app_encryption_if_device_encryption_enabled = false
  encrypt_app_data = true
  managed_apps = []
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
