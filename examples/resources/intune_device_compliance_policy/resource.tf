# Windows 10 compliance policy with immediate block on non-compliance
resource "intune_device_compliance_policy" "windows_basic" {
  display_name = "Windows 10 Corporate Compliance"
  description  = "Requires BitLocker, up-to-date OS, and a PIN"
  odata_type   = "#microsoft.graph.windows10CompliancePolicy"

  settings = {
    bitLockerEnabled              = "true"
    osMinimumVersion              = "10.0.19041"
    passwordRequired              = "true"
    passwordMinimumLength         = "8"
    passwordRequiredType          = "alphanumeric"
    passwordBlockSimple           = "true"
    storageRequireEncryption      = "true"
    secureBootEnabled             = "true"
    antivirusRequired             = "true"
    antispywareRequired           = "true"
    defenderEnabled               = "true"
  }

  # Block access immediately if non-compliant
  scheduled_action_configs = [
    {
      action_type        = "block"
      grace_period_hours = 0
    }
  ]

  assignments = [
    "00000000-0000-0000-0000-000000000001", # All Devices group
  ]
}

# iOS compliance policy
resource "intune_device_compliance_policy" "ios_basic" {
  display_name = "iOS Corporate Compliance"
  description  = "Basic iOS compliance: passcode and minimum OS version"
  odata_type   = "#microsoft.graph.iosCompliancePolicy"

  settings = {
    passcodeRequired            = "true"
    passcodeMinimumLength       = "6"
    osMinimumVersion            = "15.0"
    securityBlockJailbrokenDevices = "true"
  }

  scheduled_action_configs = [
    {
      action_type        = "block"
      grace_period_hours = 12
    }
  ]

  assignments = [
    "00000000-0000-0000-0000-000000000002", # iOS Devices group
  ]
}
