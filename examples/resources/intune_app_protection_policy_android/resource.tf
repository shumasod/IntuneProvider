# Android MAM policy targeting Microsoft 365 apps
resource "intune_app_protection_policy_android" "m365_apps" {
  display_name = "Android Microsoft 365 App Protection"
  description  = "MAM policy for M365 apps on Android BYOD devices"

  # PIN settings
  pin_required       = true
  minimum_pin_length = 6
  fingerprint_blocked = false

  # Encryption
  encrypt_app_data = true
  disable_app_encryption_if_device_encryption_is_enabled = true

  # Data protection
  data_backup_blocked    = true
  save_as_blocked        = true
  print_blocked          = true
  screen_capture_blocked = true

  # Access requirements
  device_compliance_required             = false
  managed_browser_to_open_links_required = true

  # Offline grace period (30 days)
  period_offline_before_wipe_hours = 720

  # OS version requirements
  minimum_required_os_version = "10.0"
  minimum_warning_os_version  = "9.0"

  # Target specific Android apps by package ID
  apps = [
    "com.microsoft.outlook",
    "com.microsoft.teams",
    "com.microsoft.office",
    "com.microsoft.sharepoint",
    "com.microsoft.skydrive",
  ]

  # Assign to group
  assignments = [
    "00000000-0000-0000-0000-000000000003", # All Users group
  ]
}
