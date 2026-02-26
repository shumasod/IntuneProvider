# Terraform Provider for Microsoft Intune

A Terraform provider for managing [Microsoft Intune](https://learn.microsoft.com/en-us/mem/intune/) resources via the [Microsoft Graph API](https://learn.microsoft.com/en-us/graph/overview).

## Features

| Resource / Data Source | Description |
|---|---|
| `intune_device_configuration_policy` (resource) | Device configuration profiles (Windows, iOS, Android) |
| `intune_device_compliance_policy` (resource) | Device compliance policies with non-compliance actions |
| `intune_app_protection_policy_ios` (resource) | iOS MAM (App Protection) policies |
| `intune_app_protection_policy_android` (resource) | Android MAM (App Protection) policies |
| `intune_device_configuration_policy` (data source) | Look up an existing device configuration policy |
| `intune_device_compliance_policy` (data source) | Look up an existing device compliance policy |

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.5.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- An **Azure AD Service Principal** with the following Microsoft Graph **Application** permissions:
  - `DeviceManagementConfiguration.ReadWrite.All`
  - `DeviceManagementApps.ReadWrite.All`

## Authentication

The provider uses **client credentials** (service principal) authentication against Azure AD.

```hcl
provider "intune" {
  tenant_id     = "your-tenant-id"
  client_id     = "your-client-id"
  client_secret = "your-client-secret"
}
```

Credentials can also be supplied via environment variables (recommended for CI/CD):

```bash
export INTUNE_TENANT_ID="your-tenant-id"
export INTUNE_CLIENT_ID="your-client-id"
export INTUNE_CLIENT_SECRET="your-client-secret"
```

## Quick Start

### 1. Create an Azure AD Service Principal

```bash
az ad sp create-for-rbac \
  --name "terraform-intune" \
  --role "Global Administrator" \
  --scopes /subscriptions/<subscription-id>
```

Then assign the required Graph permissions in the [Azure Portal](https://portal.azure.com) under
**Azure Active Directory → App Registrations → API Permissions**.

### 2. Configure the provider

```hcl
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
```

### 3. Create resources

```hcl
# Windows device compliance policy
resource "intune_device_compliance_policy" "windows" {
  display_name = "Windows 10 Compliance"
  odata_type   = "#microsoft.graph.windows10CompliancePolicy"

  settings = {
    bitLockerEnabled  = "true"
    passwordRequired  = "true"
    osMinimumVersion  = "10.0.19041"
  }

  scheduled_action_configs = [{
    action_type        = "block"
    grace_period_hours = 0
  }]

  assignments = ["<azure-ad-group-id>"]
}

# iOS App Protection Policy
resource "intune_app_protection_policy_ios" "outlook" {
  display_name       = "iOS Outlook MAM"
  pin_required       = true
  data_backup_blocked = true
  save_as_blocked    = true

  apps        = ["com.microsoft.Outlook"]
  assignments = ["<azure-ad-group-id>"]
}
```

## Building from Source

```bash
git clone https://github.com/shumasod/IntuneProvider.git
cd IntuneProvider
make build
make install   # installs to ~/.terraform.d/plugins/
```

## Running Tests

```bash
# Unit tests
make test

# Acceptance tests (requires real Azure AD credentials)
export INTUNE_TENANT_ID="..."
export INTUNE_CLIENT_ID="..."
export INTUNE_CLIENT_SECRET="..."
make testacc
```

## Resource Reference

### `intune_device_configuration_policy`

Manages a device configuration profile in Intune.

**Key attributes:**
- `display_name` (required) – Human-readable name
- `odata_type` (required, forces new) – Policy type, e.g. `#microsoft.graph.windows10GeneralConfiguration`
- `description` (optional) – Description
- `settings` (optional) – `map(string)` of type-specific settings
- `assignments` (optional) – List of Azure AD group IDs

### `intune_device_compliance_policy`

Manages a device compliance policy with non-compliance actions.

**Key attributes:**
- `display_name`, `odata_type`, `description`, `settings`, `assignments` – same as above
- `scheduled_action_configs` – List of `{ action_type, grace_period_hours }` blocks

### `intune_app_protection_policy_ios`

Manages an iOS MAM (App Protection) policy.

**Key attributes:**
- `display_name`, `description` – Identity
- `pin_required`, `minimum_pin_length` – PIN enforcement
- `data_backup_blocked`, `save_as_blocked`, `print_blocked` – Data protection
- `period_offline_before_wipe_hours` – Offline wipe grace period
- `apps` – List of iOS bundle IDs to target
- `assignments` – List of Azure AD group IDs

### `intune_app_protection_policy_android`

Manages an Android MAM (App Protection) policy. Similar attributes to iOS policy, plus:
- `encrypt_app_data` – Enforce app-level encryption
- `screen_capture_blocked` – Block screenshots within managed apps

## License

MIT
