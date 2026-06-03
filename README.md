# Terraform Provider for VxCloud

[![Registry](https://img.shields.io/badge/Terraform%20Registry-prodxcloud%2Fvxcloud-623CE4?style=flat-square&logo=terraform)](https://registry.terraform.io/providers/prodxcloud/vxcloud/latest)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MPL%202.0-blue?style=flat-square)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/prodxcloud/terraform-provider-vxcloud?style=flat-square)](https://goreportcard.com/report/github.com/prodxcloud/terraform-provider-vxcloud)

> **The official Terraform provider for [VxCloud](https://prodxcloud.com)** — provision and manage multi-cloud infrastructure across AWS, Azure, GCP, Alibaba Cloud, Linode, Vultr, and private VxAI environments, all from a single Terraform workflow.

---

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Quick Start](#quick-start)
- [Provider Configuration](#provider-configuration)
- [Resources](#resources)
  - [vxcloud_redis](#vxcloud_redis)
- [Data Sources](#data-sources)
- [Authentication](#authentication)
- [Environment Variables](#environment-variables)
- [Examples](#examples)
- [Server Types & Datacenters](#server-types--datacenters)
- [Local Development](#local-development)
- [Contributing](#contributing)
- [Releasing](#releasing)
- [License](#license)

---

## Overview

The **VxCloud Terraform Provider** enables infrastructure-as-code workflows for the VxCloud managed cloud platform. It uses the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) (protocol v6) and is published to the [Terraform Registry](https://registry.terraform.io/providers/prodxcloud/vxcloud/latest).

With this provider you can:

- Provision managed **Redis**, **databases**, **compute**, and **Kubernetes** resources on VxCloud
- Use **GitOps** and **CI/CD pipelines** to drive infrastructure changes
- Manage resources across **8+ cloud providers** through a unified VxCloud API
- Enforce **policy-as-code** and compliance guardrails at the infrastructure layer

---

## Requirements

| Tool | Version |
|---|---|
| [Terraform](https://developer.hashicorp.com/terraform/downloads) | >= 1.5 |
| [Go](https://golang.org/doc/install) | >= 1.22 (for local development) |
| VxCloud account | [Sign up free](https://prodxcloud.com/auth/register/) |

---

## Quick Start

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    vxcloud = {
      source  = "prodxcloud/vxcloud"
      version = "~> 0.1"
    }
  }
}

provider "vxcloud" {
  email     = var.vxcloud_email
  api_token = var.vxcloud_api_token
}
```

Then initialize:

```bash
terraform init
terraform plan
terraform apply
```

---

## Provider Configuration

```hcl
provider "vxcloud" {
  email     = "your@email.com"   # or set VXCLOUD_EMAIL env var
  api_token = "your-api-token"   # or set VXCLOUD_API_TOKEN env var
}
```

### Schema

| Argument | Type | Required | Description |
|---|---|---|---|
| `email` | string | Yes | Your VxCloud account email address |
| `api_token` | string | Yes | Your VxCloud API token (generate at prodxcloud.com) |

Credentials are read from the provider block first, then from environment variables `VXCLOUD_EMAIL` and `VXCLOUD_API_TOKEN`.

---

## Resources

### vxcloud_redis

Provisions a managed Redis instance on VxCloud.

#### Example Usage

```hcl
resource "vxcloud_redis" "cache" {
  project_id    = "1234"
  server_name   = "my-redis"
  server_type   = "SMALL-2C"
  datacenter    = "fsn"
  support_level = "level1"
}

output "redis_id" {
  value = vxcloud_redis.cache.id
}
```

#### Argument Reference

| Argument | Type | Required | Description |
|---|---|---|---|
| `project_id` | string | Yes | The VxCloud project ID to deploy into |
| `server_name` | string | Yes | Unique name for this Redis instance |
| `server_type` | string | Yes | Instance size (see [Server Types](#server-types--datacenters)) |
| `datacenter` | string | Yes | Datacenter/region code (see [Datacenters](#server-types--datacenters)) |
| `support_level` | string | Yes | Support tier: `level1`, `level2`, or `level3` |

#### Attribute Reference

| Attribute | Type | Description |
|---|---|---|
| `id` | string | The unique identifier of the Redis instance |

#### Import

Existing Redis instances can be imported using the resource ID:

```bash
terraform import vxcloud_redis.cache <resource-id>
```

---

## Data Sources

Data sources are coming in a future release. Track progress on the [issues page](https://github.com/prodxcloud/terraform-provider-vxcloud/issues).

---

## Authentication

The provider supports two authentication methods (evaluated in order):

**1. Static credentials (provider block)**
```hcl
provider "vxcloud" {
  email     = "user@example.com"
  api_token = "vxc_xxxxxxxxxxxx"
}
```

**2. Environment variables** *(recommended for CI/CD)*
```bash
export VXCLOUD_EMAIL="user@example.com"
export VXCLOUD_API_TOKEN="vxc_xxxxxxxxxxxx"
```

> **Security tip:** Never commit API tokens to source control. Use environment variables, Vault, or a secrets manager in production pipelines.

---

## Environment Variables

| Variable | Description |
|---|---|
| `VXCLOUD_EMAIL` | Account email (overrides provider block `email`) |
| `VXCLOUD_API_TOKEN` | API token (overrides provider block `api_token`) |

---

## Examples

Full examples are in the [`examples/`](./examples) directory.

### Minimal Redis

```hcl
terraform {
  required_providers {
    vxcloud = {
      source  = "prodxcloud/vxcloud"
      version = "~> 0.1"
    }
  }
}

provider "vxcloud" {}  # reads VXCLOUD_EMAIL and VXCLOUD_API_TOKEN

resource "vxcloud_redis" "app_cache" {
  project_id    = "1234"
  server_name   = "app-cache"
  server_type   = "SMALL-2C"
  datacenter    = "fsn"
  support_level = "level1"
}
```

### With Variables

```hcl
variable "vxcloud_email"     { sensitive = false }
variable "vxcloud_api_token" { sensitive = true  }
variable "project_id"        {}

provider "vxcloud" {
  email     = var.vxcloud_email
  api_token = var.vxcloud_api_token
}

resource "vxcloud_redis" "primary" {
  project_id    = var.project_id
  server_name   = "primary-cache"
  server_type   = "MEDIUM-4C"
  datacenter    = "fsn"
  support_level = "level2"
}
```

### CI/CD Pipeline (GitHub Actions)

```yaml
name: Terraform Apply

on:
  push:
    branches: [main]

jobs:
  apply:
    runs-on: ubuntu-latest
    env:
      VXCLOUD_EMAIL:     ${{ secrets.VXCLOUD_EMAIL }}
      VXCLOUD_API_TOKEN: ${{ secrets.VXCLOUD_API_TOKEN }}
    steps:
      - uses: actions/checkout@v4
      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "~> 1.9"
      - run: terraform init
      - run: terraform plan -out=tfplan
      - run: terraform apply tfplan
```

---

## Server Types & Datacenters

### Server Types

| Code | vCPU | RAM | Use Case |
|---|---|---|---|
| `SMALL-2C` | 2 | 2 GB | Dev / staging caches |
| `MEDIUM-4C` | 4 | 8 GB | Production apps |
| `LARGE-8C` | 8 | 32 GB | High-throughput workloads |
| `XL-16C` | 16 | 64 GB | Enterprise / data-intensive |

### Datacenters

| Code | Region | Location |
|---|---|---|
| `fsn` | EU | Falkenstein, Germany |
| `nbg` | EU | Nuremberg, Germany |
| `hel` | EU | Helsinki, Finland |
| `ash` | US | Ashburn, Virginia |
| `hil` | US | Hillsboro, Oregon |
| `sin` | APAC | Singapore |

---

## Local Development

### Prerequisites

- Go 1.22+
- Terraform CLI 1.5+
- A VxCloud account and API token

### Build & Install

```bash
# Clone the repository
git clone https://github.com/prodxcloud/terraform-provider-vxcloud.git
cd terraform-provider-vxcloud

# Download dependencies
go mod tidy

# Build
go build -o terraform-provider-vxcloud

# Install into local Terraform plugin cache
make install
```

The `make install` target copies the binary to:
```
~/.terraform.d/plugins/registry.terraform.io/prodxcloud/vxcloud/<version>/<os>_<arch>/
```

### Use the Local Build

Add an override to your `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "prodxcloud/vxcloud" = "<path-to-repo>"
  }
  direct {}
}
```

Then run `terraform plan` without `terraform init` (the override skips the registry).

### Run Tests

```bash
# Unit tests
go test ./...

# Acceptance tests (requires live VxCloud credentials)
VXCLOUD_EMAIL=your@email.com VXCLOUD_API_TOKEN=your-token go test ./internal/provider/... -v -run TestAcc
```

### Makefile Targets

```bash
make build    # compile the provider binary
make install  # build + copy to local plugin cache
make test     # run unit tests
make lint     # run golangci-lint
make docs     # generate provider documentation
```

---

## Contributing

Contributions are welcome! Please follow these steps:

1. **Fork** the repository and create a feature branch:
   ```bash
   git checkout -b feat/my-new-resource
   ```
2. **Write code** following the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) guidelines
3. **Add tests** — all new resources must include acceptance tests
4. **Update docs** — add schema documentation and examples in `examples/`
5. **Open a Pull Request** against `main` with a clear description

### Code Structure

```
terraform-provider-vxcloud/
├── main.go                        # Provider entry point
├── internal/
│   └── provider/
│       ├── provider.go            # Provider schema & configuration
│       ├── client.go              # VxCloud API client
│       └── redis_resource.go      # vxcloud_redis resource
├── examples/
│   ├── provider/                  # Provider usage examples
│   └── resources/vxcloud_redis/   # Redis resource examples
├── Makefile                       # Build, test, and release targets
├── go.mod                         # Go module definition
└── terraform-registry-manifest.json
```

### Adding a New Resource

1. Create `internal/provider/<resource>_resource.go`
2. Implement the `resource.Resource` interface (Metadata, Schema, Configure, Create, Read, Update, Delete)
3. Register it in `provider.go` under `Resources()`
4. Add an example in `examples/resources/<resource>/`
5. Run `make docs` to regenerate documentation

---

## Releasing

Releases are automated via GitHub Actions. To publish a new version:

1. Tag the release:
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```
2. The release workflow will:
   - Cross-compile binaries for Linux, macOS, and Windows (amd64 + arm64)
   - Sign artifacts with the GPG key
   - Publish to GitHub Releases
   - The Terraform Registry indexes the release automatically

### Required Repository Secrets

| Secret | Description |
|---|---|
| `GPG_PRIVATE_KEY` | ASCII-armored GPG private key used to sign release artifacts |
| `PASSPHRASE` | Passphrase for the GPG key |

> The matching **public key** must be registered on your [Terraform Registry account](https://registry.terraform.io/sign-in).

---

## Resources & Links

| Resource | URL |
|---|---|
| VxCloud Platform | [prodxcloud.com](https://prodxcloud.com) |
| Terraform Registry | [registry.terraform.io/providers/prodxcloud/vxcloud](https://registry.terraform.io/providers/prodxcloud/vxcloud/latest) |
| VxCloud API Docs | [prodxcloud.com/docs](https://prodxcloud.com/docs/) |
| Plugin Framework Docs | [developer.hashicorp.com/terraform/plugin/framework](https://developer.hashicorp.com/terraform/plugin/framework) |
| Issues & Feature Requests | [github.com/prodxcloud/terraform-provider-vxcloud/issues](https://github.com/prodxcloud/terraform-provider-vxcloud/issues) |
| VxCloud GitHub Org | [github.com/prodxcloud](https://github.com/prodxcloud) |

---

## License

This provider is licensed under the [Mozilla Public License 2.0](LICENSE).

Copyright (c) 2026 ProdXCloud. All rights reserved.
