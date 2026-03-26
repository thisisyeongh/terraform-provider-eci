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

func TestAccPublicIp_basic(t *testing.T) {
	instanceTypeName := os.Getenv("ECI_TEST_INSTANCE_TYPE_NAME")
	vmPricingName := os.Getenv("ECI_TEST_VM_PRICING_NAME")
	vmPricingType := os.Getenv("ECI_TEST_VM_PRICING_TYPE")
	pipPricingName := os.Getenv("ECI_TEST_PIP_PRICING_NAME")
	pipPricingType := os.Getenv("ECI_TEST_PIP_PRICING_TYPE")
	if instanceTypeName == "" || vmPricingName == "" || vmPricingType == "" ||
		pipPricingName == "" || pipPricingType == "" {
		t.Skip("Required ECI_TEST_* env vars not set for public IP test")
	}

	rName := acctest.RandomName("pip")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPublicIpDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPublicIpConfig_basic(
					rName, instanceTypeName, vmPricingName, vmPricingType,
					pipPricingName, pipPricingType,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("eci_public_ip.test", "id"),
					resource.TestCheckResourceAttrSet("eci_public_ip.test", "ip"),
					resource.TestCheckResourceAttrSet("eci_public_ip.test", "status"),
				),
			},
		},
	})
}

func testAccCheckPublicIpDestroy(s *terraform.State) error {
	client, err := acctest.SharedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "eci_public_ip" {
			continue
		}

		resp, err := client.GetPublicIp(rs.Primary.ID)
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

		return fmt.Errorf("public ip %s still exists (status: %s)", rs.Primary.ID, resp.Status)
	}
	return nil
}

func testAccPublicIpConfig_basic(
	rName, instanceTypeName, vmPricingName, vmPricingType,
	pipPricingName, pipPricingType string,
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

data "eci_pricing" "pip" {
  name          = %[5]q
  pricing_type  = %[6]q
  resource_kind = "public_ip"
}

resource "eci_virtual_machine" "test" {
  name             = "%[1]s-vm"
  instance_type_id = data.eci_instance_type.test.id
  pricing_id       = data.eci_pricing.vm.id
  always_on        = false
  dr               = false
  username         = "testuser"
  password         = %[7]q
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
  name                = "%[1]s-nic"
  attached_subnet_id  = eci_subnet.test.id
  attached_machine_id = eci_virtual_machine.test.id
  dr                  = false
  tags                = {}
}

resource "eci_public_ip" "test" {
  attached_network_interface_id = eci_network_interface.test.id
  dr                            = false
  pricing_id                    = data.eci_pricing.pip.id
  tags                          = {}
}
`, rName, instanceTypeName, vmPricingName, vmPricingType,
		pipPricingName, pipPricingType, acctest.TestPassword)
}
