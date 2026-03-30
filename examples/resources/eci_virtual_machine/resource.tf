data "eci_pricing" "vm_pricing" {
  name         = "M-8"
  pricing_type = "ondemand"
}

resource "eci_virtual_machine" "my_virtual_machine" {
  name="my-vm-1"
  instance_type_id="d0ba1aed-1414-4388-9c2a-9083ae3154d2"
  pricing_id="${data.eci_pricing.vm_pricing.id}"
  always_on=false
  username="elice"
  password="secretpassword1!"
  on_init_script="#!/bin/bash\necho 'Hello, Elice!' > /home/elice/hello.txt\nchmod 644 /home/elice/hello.txt"
  dr=false
  tags = {
    "created-by": "terraform"
  }
}
