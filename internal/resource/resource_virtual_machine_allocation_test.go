package resource_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"terraform-provider-eci/internal/acctest"
	"terraform-provider-eci/internal/api"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccVirtualMachineAllocation_basic(t *testing.T) {
	instanceTypeName := os.Getenv("ECI_TEST_INSTANCE_TYPE_NAME")
	vmPricingName := os.Getenv("ECI_TEST_VM_PRICING_NAME")
	vmPricingType := os.Getenv("ECI_TEST_VM_PRICING_TYPE")
	bsPricingName := os.Getenv("ECI_TEST_BS_PRICING_NAME")
	bsPricingType := os.Getenv("ECI_TEST_BS_PRICING_TYPE")
	imageName := os.Getenv("ECI_TEST_IMAGE_NAME")
	if instanceTypeName == "" || vmPricingName == "" || vmPricingType == "" ||
		bsPricingName == "" || bsPricingType == "" || imageName == "" {
		t.Skip("Required ECI_TEST_* env vars not set for VM allocation test")
	}

	rName := acctest.RandomName("alloc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVirtualMachineAllocationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineAllocationConfig_basic(
					rName, instanceTypeName, vmPricingName, vmPricingType,
					bsPricingName, bsPricingType, imageName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("eci_virtual_machine_allocation.test", "id"),
					resource.TestCheckResourceAttrSet("eci_virtual_machine_allocation.test", "machine_id"),
					resource.TestCheckResourceAttrSet("eci_virtual_machine_allocation.test", "status"),
				),
			},
		},
	})
}

func testAccCheckVirtualMachineAllocationDestroy(s *terraform.State) error {
	client, err := acctest.SharedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "eci_virtual_machine_allocation" {
			continue
		}

		resp, err := client.GetVirtualMachineAllocation(rs.Primary.ID)
		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) && apiErr.HttpCode == 404 {
				continue
			}
			return err
		}

		if resp.Status == "terminated" {
			continue
		}

		return fmt.Errorf("virtual machine allocation %s still exists (status: %s)", rs.Primary.ID, resp.Status)
	}
	return nil
}

func testAccVirtualMachineAllocationConfig_basic(
	rName, instanceTypeName, vmPricingName, vmPricingType,
	bsPricingName, bsPricingType, imageName string,
) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
data "eci_instance_type" "test" {
  name = %[2]q
}

data "eci_pricing" "vm" {
  name          = %[3]q
  pricing_type  = %[4]q
  resource_kind = "vm_allocation"
}

data "eci_pricing" "bs" {
  name          = %[5]q
  pricing_type  = %[6]q
  resource_kind = "block_storage"
}

data "eci_block_storage_image" "test" {
  name = %[7]q
}

resource "eci_virtual_machine" "test" {
  name             = "%[1]s-vm"
  instance_type_id = data.eci_instance_type.test.id
  pricing_id       = data.eci_pricing.vm.id
  always_on        = false
  dr               = false
  username         = "testuser"
  password         = %[8]q
  on_init_script   = ""
  tags             = {}
}

resource "eci_block_storage" "test" {
  name                = "%[1]s-bs"
  attached_machine_id = eci_virtual_machine.test.id
  image_id            = data.eci_block_storage_image.test.id
  size_gib            = data.eci_block_storage_image.test.size_gib
  dr                  = false
  pricing_id          = data.eci_pricing.bs.id
  tags                = {}
}

resource "eci_virtual_network" "test" {
  name           = "%[1]s-vnet"
  network_cidr   = "192.168.0.0/16"
  tags           = {}
  firewall_rules = []
}

resource "eci_subnet" "test" {
  name                = "%[1]s-subnet"
  attached_network_id = eci_virtual_network.test.id
  purpose             = "virtual_machine"
  network_gw          = "192.168.0.1/24"
  tags                = {}
}

resource "eci_network_interface" "test" {
  name                = "%[1]s-nic"
  attached_subnet_id  = eci_subnet.test.id
  attached_machine_id = eci_virtual_machine.test.id
  dr                  = false
  tags                = {}
}

resource "eci_virtual_machine_allocation" "test" {
  machine_id = eci_virtual_machine.test.id
  tags       = {}

  depends_on = [eci_block_storage.test, eci_network_interface.test]
}
`, rName, instanceTypeName, vmPricingName, vmPricingType,
		bsPricingName, bsPricingType, imageName, acctest.TestPassword)
}
