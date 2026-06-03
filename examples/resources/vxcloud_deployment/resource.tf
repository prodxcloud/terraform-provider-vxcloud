# Deploy a Docker container/app onto one of your VMs over SSH.
# This is the Terraform equivalent of `vxcli deploy container`.

# 1) A managed Redis cache on an Ubuntu VM.
resource "vxcloud_deployment" "redis" {
  name          = "redis"
  image         = "redis:7"
  host          = "203.0.113.10"   # your VM's public IP
  ssh_user      = "ubuntu"
  key_pair_name = "prod-ssh-key"   # SSH key stored in your vxcloud Vault

  ports = ["6379:6379"]
  env   = ["REDIS_ARGS=--save 60 1000"]

  restart_policy = "unless-stopped"
}

# 2) A public web app with automatic Let's Encrypt TLS.
resource "vxcloud_deployment" "api" {
  name          = "finance-api"
  image         = "ghcr.io/your-org/finance-api:latest"
  host          = "203.0.113.20"
  ssh_user      = "root"
  key_pair_name = "prod-ssh-key"

  ports = ["80:8000"]
  env = [
    "ENV=production",
    "LOG_LEVEL=info",
  ]

  enable_ssl = true
  domain     = "api.example.com"
  ssl_email  = "ops@example.com"
}

output "redis_session" {
  value = vxcloud_deployment.redis.session_id
}
