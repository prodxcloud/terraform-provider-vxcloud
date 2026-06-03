terraform {
  required_providers {
    vxcloud = {
      source = "prodxcloud/vxcloud"
    }
  }
}

provider "vxcloud" {
  # Developer API key (xc_dev_… / xc_live_…), sent as X-API-Key.
  # Prefer the VXCLOUD_API_TOKEN (or VXCLOUD_API_KEY) env var over hard-coding.
  api_token = var.vxcloud_api_token

  # Tenant node where deploys + agentcontrol run. Defaults to node1.vxcloud.io.
  endpoint = "https://node1.vxcloud.io"

  # Required for vxcloud_agent (agentcontrol) resources.
  tenant_id = var.vxcloud_tenant_id
  username  = "your-username"
}

variable "vxcloud_api_token" {
  type      = string
  sensitive = true
}

variable "vxcloud_tenant_id" {
  type    = string
  default = ""
}
