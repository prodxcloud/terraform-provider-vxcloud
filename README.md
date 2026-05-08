# Terraform Provider for vxcloud
dfdfdf
Manage [vxcloud](https://vxcloud.com) resources from Terraform.

## Usage

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
  email     = "your-email"
  api_token = "your-api-token"
}

resource "vxcloud_redis" "redis_service" {
  project_id    = "1234"
  server_name   = "my-redis"
  server_type   = "SMALL-2C"
  datacenter    = "fsn"
  support_level = "level1"
}
```

Credentials may also be supplied via `VXCLOUD_EMAIL` and `VXCLOUD_API_TOKEN`.

## Local development

```bash
go mod tidy
go build -o terraform-provider-vxcloud
make install
```

## Releasing

Tag a version (`git tag v0.1.0 && git push --tags`). The release workflow signs and publishes artifacts that the Terraform Registry indexes automatically.

Required repo secrets:

- `GPG_PRIVATE_KEY` — ASCII-armored private key
- `PASSPHRASE` — passphrase for that key

The matching public key must be registered on your Terraform Registry account.

## License

MPL-2.0
