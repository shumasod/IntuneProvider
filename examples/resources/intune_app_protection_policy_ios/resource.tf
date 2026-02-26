# iOS MAM policy targeting Microsoft 365 apps
resource "intune_app_protection_policy_ios" "m365_apps" {
  display_name = "iOS Microsoft 365 App Protection"
  description  = "MAM policy for M365 apps on iOS BYOD devices"

  # PIN settings
  pin_required       = true
  minimum_pin_length = 6
  fingerprint_blocked = false
  disable_app_pin_if_device_pin_is_set = true

  # Data protection
  data_backup_blocked  = true
  save_as_blocked      = true
  print_blocked        = true
  organizer_sync_blocked = false

  # Access requirements
  device_compliance_required             = false
  managed_browser_to_open_links_required = true

  # Offline grace period (30 days)
  period_offline_before_wipe_hours = 720

  # OS version requirements
  minimum_required_os_version = "15.0"
  minimum_warning_os_version  = "14.0"

  # Target specific iOS apps by bundle ID
  apps = [
    "com.microsoft.Outlook",
    "com.microsoft.teams",
    "com.microsoft.Office",
    "com.microsoft.sharepoint",
    "com.microsoft.OneDrive",
  ]

  # Assign to group
  assignments = [
    "00000000-0000-0000-0000-000000000003", # All Users group
  ]
}
