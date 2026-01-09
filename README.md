# Terraform Provider VaultSec

This custom Terraform provider allows you to handle Vault secrets ephemerally. It is designed to retrieve an existing password from HashiCorp Vault or generate a new random one if the secret does not exist, without storing the sensitive value in the Terraform state file.

Additionally, it includes a data source to retrieve the current version of a secret, which requires access to Vault metadata (depending on your Vault policy), a feature that differentiates it from the official Vault provider in specific use cases.

## Provider Configuration

The provider requires the Vault address and a valid token.

```hcl
provider "vaultsec" {
  address = "https://vault.example.com"
  token   = var.vault_token
}
```

## Resources

### Ephemeral Resource: `vaultsec_secret`

This ephemeral resource attempts to read a secret from Vault. If the secret exists, it returns the value. If the secret does not exist (404), it generates a new random password.

**Note:** As an ephemeral resource, the values retrieved or generated are not persisted in the Terraform state.

#### Arguments

* `mount` (String, Required) The mount path of the KV-v2 secret engine.
* `name` (String, Required) The name (path) of the secret.
* `key` (String, Optional) The key within the secret to retrieve/generate. Defaults to "password".
* `password_len` (Number, Optional) The length of the generated password if a new one is created.
* `override_special` (String, Optional) A string of special characters to use when generating a new password. If not provided, defaults to `!@#$%^&*()`.

#### Attributes

* `password` (String, Sensitive) The retrieved or generated password.
* `version` (Number) The version of the secret found in Vault (or 1 if newly generated).

### Data Source: `vaultsec_secret_version`

This data source retrieves the current version of a secret from Vault metadata.

#### Arguments

* `mount` (String, Required) The mount path of the KV-v2 secret engine.
* `name` (String, Required) The name (path) of the secret.
* `key` (String, Optional) The key to look up (present for schema consistency, defaults to "password").

#### Attributes

* `version` (Number) The current version of the secret in Vault. Returns `0` if the secret does not exist.

## Installation

To build the provider from source and register it for local use with Terraform, follow these steps:

```bash
# Set version and platform variables
VERSION=0.2.0
PLATFORM=darwin_arm64 # Change to your platform (e.g., linux_amd64, darwin_amd64)

# Create the local plugin directory
mkdir -p $HOME/.terraform.d/plugins/registry.terraform.io/linero/vaultsec/$VERSION/$PLATFORM

# Build the provider and move it to the plugin directory
go build -o $HOME/.terraform.d/plugins/registry.terraform.io/linero/vaultsec/$VERSION/$PLATFORM/terraform-provider-vaultsec_v$VERSION
```

After building, you can reference the provider in your Terraform configuration:

```hcl
terraform {
  required_providers {
    vaultsec = {
      source  = "registry.terraform.io/linero/vaultsec"
      version = "0.2.0"
    }
  }
}
```

## Example Usage

The following example demonstrates how to use the provider to manage PostgreSQL users. It retrieves a password from Vault (or generates a new one), checks the current secret version, and then creates/updates the Vault secret and the PostgreSQL role accordingly.

```hcl
# Define the ephemeral resource to fetch/generate the password
ephemeral "vaultsec_secret" "secret" {
  mount        = vault_mount.kvv2.path
  name         = "/users/example"
  password_len = 32
}

# Define the data source to get the current secret version
data "vaultsec_secret_version" "secret_version" {
  mount = vault_mount.kvv2.path
  name  = "/users/example"
}

# Store the password in Vault
resource "vault_kv_secret_v2" "user" {
  mount               = vault_mount.kvv2.path
  name                = "/users/example"
  delete_all_versions = true

  data_json_wo = jsonencode({
    password = ephemeral.vaultsec_secret.secret.password
  })

  data_json_wo_version = 1
}

# Create the PostgreSQL role using the retrieved/generated password
resource "postgresql_role" "user" {
  name  = "example"
  login = true

  superuser       = true

  # Use the password from the ephemeral resource
  password_wo = ephemeral.vaultsec_secret.secret[each.key].password

  # Manage password rotation based on version changes
  # If version is 0 (new secret), use 1, otherwise use the found version
  password_wo_version = data.vaultsec_secret_version.secret_version.version == 0 ? 1 : data.vaultsec_secret_version.secret_version.version

  depends_on = [vault_kv_secret_v2.user]
}
```
