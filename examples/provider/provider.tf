terraform {
  required_providers {
    vxcloud = {
      source = "prodxcloud/vxcloud"
    }
  }
}

provider "vxcloud" {
  email     = "your-email"
  api_token = "your-api-token"
}
