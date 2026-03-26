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

func TestAccVirtualMachine_basic(t *testing.T) {
	instanceTypeName := os.Getenv("ECI_TEST_INSTANCE_TYPE_NAME")
	pricingName := os.Getenv("ECI_TEST_VM_PRICING_NAME")
	pricingType := os.Getenv("ECI_TEST_VM_PRICING_TYPE")
	if instanceTypeName == "" || pricingName == "" || pricingType == "" {
		t.Skip("ECI_TEST_INSTANCE_TYPE_NAME, ECI_TEST_VM_PRICING_NAME, and ECI_TEST_VM_PRICING_TYPE must be set")
	}

	rName := acctest.RandomName("vm")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineConfig_basic(rName, instanceTypeName, pricingName, pricingType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("eci_virtual_machine.test", "id"),
					resource.TestCheckResourceAttr("eci_virtual_machine.test", "name", rName),
					resource.TestCheckResourceAttrSet("eci_virtual_machine.test", "status"),
				),
			},
		},
	})
}

func TestAccVirtualMachine_update(t *testing.T) {
	instanceTypeName := os.Getenv("ECI_TEST_INSTANCE_TYPE_NAME")
	pricingName := os.Getenv("ECI_TEST_VM_PRICING_NAME")
	pricingType := os.Getenv("ECI_TEST_VM_PRICING_TYPE")
	if instanceTypeName == "" || pricingName == "" || pricingType == "" {
		t.Skip("ECI_TEST_INSTANCE_TYPE_NAME, ECI_TEST_VM_PRICING_NAME, and ECI_TEST_VM_PRICING_TYPE must be set")
	}

	rName := acctest.RandomName("vm")
	rNameUpdated := acctest.RandomName("vm-upd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVirtualMachineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineConfig_basic(rName, instanceTypeName, pricingName, pricingType),
				Check:  resource.TestCheckResourceAttr("eci_virtual_machine.test", "name", rName),
			},
			{
				Config: testAccVirtualMachineConfig_basic(rNameUpdated, instanceTypeName, pricingName, pricingType),
				Check:  resource.TestCheckResourceAttr("eci_virtual_machine.test", "name", rNameUpdated),
			},
		},
	})
}

func testAccCheckVirtualMachineDestroy(s *terraform.State) error {
	client, err := acctest.SharedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "eci_virtual_machine" {
			continue
		}

		resp, err := client.GetVirtualMachine(rs.Primary.ID)
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

		return fmt.Errorf("virtual machine %s still exists (status: %s)", rs.Primary.ID, resp.Status)
	}
	return nil
}

func testAccVirtualMachineConfig_basic(rName, instanceTypeName, pricingName, pricingType string) string {
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
  name             = %[1]q
  instance_type_id = data.eci_instance_type.test.id
  pricing_id       = data.eci_pricing.test.id
  always_on        = false
  dr               = false
  username         = "testuser"
  password         = %[5]q
  on_init_script   = ""
  tags             = {}
}
`, rName, instanceTypeName, pricingName, pricingType, acctest.TestPassword)
}
