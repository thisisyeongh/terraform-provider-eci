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

func TestAccNetworkInterface_basic(t *testing.T) {
	instanceTypeName := os.Getenv("ECI_TEST_INSTANCE_TYPE_NAME")
	pricingName := os.Getenv("ECI_TEST_VM_PRICING_NAME")
	pricingType := os.Getenv("ECI_TEST_VM_PRICING_TYPE")
	if instanceTypeName == "" || pricingName == "" || pricingType == "" {
		t.Skip("Required ECI_TEST_* env vars not set for network interface test")
	}

	rName := acctest.RandomName("nic")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNetworkInterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkInterfaceConfig_basic(rName, instanceTypeName, pricingName, pricingType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("eci_network_interface.test", "id"),
					resource.TestCheckResourceAttr("eci_network_interface.test", "name", rName),
					resource.TestCheckResourceAttrSet("eci_network_interface.test", "attached_subnet_id"),
					resource.TestCheckResourceAttrSet("eci_network_interface.test", "status"),
				),
			},
		},
	})
}

func testAccCheckNetworkInterfaceDestroy(s *terraform.State) error {
	client, err := acctest.SharedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "eci_network_interface" {
			continue
		}

		resp, err := client.GetNetworkInterface(rs.Primary.ID)
		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) && apiErr.HttpCode == 404 {
				continue
			}
			return err
		}

		if resp.Status == "deleted" {
			continue
		}

		return fmt.Errorf("network interface %s still exists (status: %s)", rs.Primary.ID, resp.Status)
	}
	return nil
}

func testAccNetworkInterfaceConfig_basic(rName, instanceTypeName, pricingName, pricingType string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
data "eci_instance_type" "test" {
  name = %[2]q
}

data "eci_pricing" "test" {
  name          = %[3]q
  pricing_type  = %[4]q
  resource_kind = "vm_allocation"
}

resource "eci_virtual_machine" "test" {
  name             = "%[1]s-vm"
  instance_type_id = data.eci_instance_type.test.id
  pricing_id       = data.eci_pricing.test.id
  always_on        = false
  dr               = false
  username         = "testuser"
  password         = %[5]q
  on_init_script   = ""
  tags             = {}
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
  name                = %[1]q
  attached_subnet_id  = eci_subnet.test.id
  attached_machine_id = eci_virtual_machine.test.id
  dr                  = false
  tags                = {}
}
`, rName, instanceTypeName, pricingName, pricingType, acctest.TestPassword)
}
