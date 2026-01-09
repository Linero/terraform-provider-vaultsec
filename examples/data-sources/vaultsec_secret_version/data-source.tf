data "vaultsec_secret_version" "example" {
  mount = "secret"
  name  = "my/secret"
}
