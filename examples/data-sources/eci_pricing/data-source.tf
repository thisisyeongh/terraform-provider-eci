data "eci_pricing" "vm_pricing" {
  name         = "M-8"
  pricing_type = "ondemand"
}

data "eci_pricing" "storage_pricing" {
  name         = "Block Storage"
  pricing_type = "ondemand"
}

data "eci_pricing" "ip_pricing" {
  name         = "Public IP"
  pricing_type = "ondemand"
}
