package resource_test

import (
	"errors"
	"fmt"
	"testing"

	"terraform-provider-eci/internal/acctest"
	"terraform-provider-eci/internal/api"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSubnet_basic(t *testing.T) {
	rName := acctest.RandomName("subnet")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("eci_subnet.test", "id"),
					resource.TestCheckResourceAttr("eci_subnet.test", "name", rName),
					resource.TestCheckResourceAttr("eci_subnet.test", "purpose", "virtual_machine"),
					resource.TestCheckResourceAttr("eci_subnet.test", "network_gw", "192.168.0.1/24"),
					resource.TestCheckResourceAttrSet("eci_subnet.test", "attached_network_id"),
				),
			},
		},
	})
}

func TestAccSubnet_update(t *testing.T) {
	rName := acctest.RandomName("subnet")
	rNameUpdated := acctest.RandomName("subnet-upd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig_basic(rName),
				Check:  resource.TestCheckResourceAttr("eci_subnet.test", "name", rName),
			},
			{
				Config: testAccSubnetConfig_basic(rNameUpdated),
				Check:  resource.TestCheckResourceAttr("eci_subnet.test", "name", rNameUpdated),
			},
		},
	})
}

func testAccCheckSubnetDestroy(s *terraform.State) error {
	client, err := acctest.SharedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "eci_subnet" {
			continue
		}

		resp, err := client.GetSubnet(rs.Primary.ID)
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

		return fmt.Errorf("subnet %s still exists (status: %s)", rs.Primary.ID, resp.Status)
	}
	return nil
}

func testAccSubnetConfig_basic(rName string) string {
	vnetName := rName + "-vnet"
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "eci_virtual_network" "test" {
  name           = %[1]q
  network_cidr   = "192.168.0.0/16"
  tags           = {}
  firewall_rules = []
}

resource "eci_subnet" "test" {
  name                = %[2]q
  attached_network_id = eci_virtual_network.test.id
  purpose             = "virtual_machine"
  network_gw          = "192.168.0.1/24"
  tags                = {}
}
`, vnetName, rName)
}
