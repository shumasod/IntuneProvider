# Windows 10 Endpoint Protection configuration policy
resource "intune_device_configuration_policy" "windows_defender" {
  display_name = "Windows Defender Antivirus Configuration"
  description  = "Enables Windows Defender with real-time protection and cloud-delivered protection"
  odata_type   = "#microsoft.graph.windows10EndpointProtectionConfiguration"

  settings = {
    defenderEnabled                              = "true"
    defenderMonitorFileActivity                  = "monitorAllFiles"
    defenderScanNetworkFiles                     = "true"
    defenderScanArchiveFiles                     = "true"
    defenderCloudBlockLevel                      = "high"
    defenderCloudExtendedTimeoutInSeconds        = "50"
    defenderSubmitSamplesConsentType             = "sendSafeSamplesAutomatically"
  }

  # Assign to an Azure AD group
  assignments = [
    "00000000-0000-0000-0000-000000000001", # All Devices group
  ]
}

# iOS General Device Configuration
resource "intune_device_configuration_policy" "ios_general" {
  display_name = "iOS General Device Restrictions"
  description  = "Restricts iOS device features for corporate devices"
  odata_type   = "#microsoft.graph.iosGeneralDeviceConfiguration"

  settings = {
    screenCaptureBlocked         = "true"
    iCloudBlockBackup            = "false"
    passcodeMinimumLength        = "6"
    passcodeRequired             = "true"
    passcodeRequiredType         = "numeric"
  }

  assignments = [
    "00000000-0000-0000-0000-000000000002", # iOS Devices group
  ]
}
