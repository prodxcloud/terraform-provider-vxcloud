resource "vxcloud_redis" "redis_service" {
  project_id    = "1234"
  server_name   = "my-redis"
  server_type   = "SMALL-2C"
  datacenter    = "fsn"
  support_level = "level1"
}
