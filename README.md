# Terraform Provider for Bitwarden Secrets

This project builds a Terraform Proivder on top of Bitwardens [Secrets Manager CLI](https://bitwarden.com/help/secrets-manager-cli/). It allows for reading secrets into Data Sources, or managing secrets or projects through Resources.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Secrets Manager CLI](https://bitwarden.com/help/secrets-manager-cli/) >= v0.5.0
- [Go](https://golang.org/doc/install) >= 1.21 (development)

_The CLI binary should be added to the path such that it is accessible by the Terraform provider!_

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

```tf
terraform {
  required_providers {
    bitwarden-secrets = {
      source  = "bitwarden-secrets"
      version = ">= 0.1.0"
    }
  }
}

# Configure the Bitwarden Secrets Provider
provider "bitwarden-secrets" {
  access_token = "Token acquired from Bitwarden Secrets Web UI"
}

# Create a Terraform managed project
resource "bitwarden-secrets_project" "example_project" {
    name = "Terraform-Secrets"
}

# Create a Terraform managed secret
resource "bitwarden-secrets_secret" "example" {
    key = "test-terraform"
    value = "hello world!"
    project_id = bitwarden-secrets_project.example_project.id
}

# Or get a secret directly by using its id
data "bitwarden-secrets_secret" "vpn" {
    id = "Id of the secret"
}
```

When reading secrets make sure the current provided access token has permissions to read from the associated project. Furthermore, when making use of a secret resource on a project managed outside of Terraform Read & Write permissions should be enabled for the access token.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
