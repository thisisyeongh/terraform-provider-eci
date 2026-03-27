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

func TestAccBlockStorage_basic(t *testing.T) {
	instanceTypeName := os.Getenv("ECI_TEST_INSTANCE_TYPE_NAME")
	vmPricingName := os.Getenv("ECI_TEST_VM_PRICING_NAME")
	vmPricingType := os.Getenv("ECI_TEST_VM_PRICING_TYPE")
	bsPricingName := os.Getenv("ECI_TEST_BS_PRICING_NAME")
	bsPricingType := os.Getenv("ECI_TEST_BS_PRICING_TYPE")
	imageName := os.Getenv("ECI_TEST_IMAGE_NAME")
	if instanceTypeName == "" || vmPricingName == "" || vmPricingType == "" ||
		bsPricingName == "" || bsPricingType == "" || imageName == "" {
		t.Skip("Required ECI_TEST_* env vars not set for block storage test")
	}

	rName := acctest.RandomName("bs")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckBlockStorageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageConfig_basic(
					rName, instanceTypeName, vmPricingName, vmPricingType,
					bsPricingName, bsPricingType, imageName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("eci_block_storage.test", "id"),
					resource.TestCheckResourceAttr("eci_block_storage.test", "name", rName),
					resource.TestCheckResourceAttrSet("eci_block_storage.test", "status"),
				),
			},
		},
	})
}

func testAccCheckBlockStorageDestroy(s *terraform.State) error {
	client, err := acctest.SharedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "eci_block_storage" {
			continue
		}

		resp, err := client.GetBlockStorage(rs.Primary.ID)
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

		return fmt.Errorf("block storage %s still exists (status: %s)", rs.Primary.ID, resp.Status)
	}
	return nil
}

func testAccBlockStorageConfig_basic(
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
  name                = %[1]q
  attached_machine_id = eci_virtual_machine.test.id
  image_id            = data.eci_block_storage_image.test.id
  size_gib            = data.eci_block_storage_image.test.size_gib
  dr                  = false
  pricing_id          = data.eci_pricing.bs.id
  tags                = {}
}
`, rName, instanceTypeName, vmPricingName, vmPricingType,
		bsPricingName, bsPricingType, imageName, acctest.TestPassword)
}
