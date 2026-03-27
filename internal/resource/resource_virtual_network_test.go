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

func TestAccVirtualNetwork_basic(t *testing.T) {
	rName := acctest.RandomName("vnet")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNetworkConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("eci_virtual_network.test", "id"),
					resource.TestCheckResourceAttr("eci_virtual_network.test", "name", rName),
					resource.TestCheckResourceAttr("eci_virtual_network.test", "network_cidr", "192.168.0.0/16"),
					resource.TestCheckResourceAttrSet("eci_virtual_network.test", "status"),
					resource.TestCheckResourceAttrSet("eci_virtual_network.test", "zone_id"),
					resource.TestCheckResourceAttrSet("eci_virtual_network.test", "organization_id"),
				),
			},
		},
	})
}

func TestAccVirtualNetwork_update(t *testing.T) {
	rName := acctest.RandomName("vnet")
	rNameUpdated := acctest.RandomName("vnet-upd")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNetworkConfig_basic(rName),
				Check:  resource.TestCheckResourceAttr("eci_virtual_network.test", "name", rName),
			},
			{
				Config: testAccVirtualNetworkConfig_basic(rNameUpdated),
				Check:  resource.TestCheckResourceAttr("eci_virtual_network.test", "name", rNameUpdated),
			},
		},
	})
}

func TestAccVirtualNetwork_firewallRules(t *testing.T) {
	rName := acctest.RandomName("vnet-fw")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckVirtualNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualNetworkConfig_firewallRules(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("eci_virtual_network.test", "id"),
					resource.TestCheckResourceAttr("eci_virtual_network.test", "name", rName),
					resource.TestCheckResourceAttr("eci_virtual_network.test", "firewall_rules.#", "1"),
				),
			},
		},
	})
}

func testAccCheckVirtualNetworkDestroy(s *terraform.State) error {
	client, err := acctest.SharedClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "eci_virtual_network" {
			continue
		}

		resp, err := client.GetVirtualNetwork(rs.Primary.ID)
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

		return fmt.Errorf("virtual network %s still exists (status: %s)", rs.Primary.ID, resp.Status)
	}
	return nil
}

func testAccVirtualNetworkConfig_basic(rName string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "eci_virtual_network" "test" {
  name           = %[1]q
  network_cidr   = "192.168.0.0/16"
  tags           = {}
  firewall_rules = []
}
`, rName)
}

func testAccVirtualNetworkConfig_firewallRules(rName string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
resource "eci_virtual_network" "test" {
  name         = %[1]q
  network_cidr = "192.168.0.0/16"
  tags         = {}

  firewall_rules = [
    {
      proto       = "TCP"
      source      = "0.0.0.0/0"
      destination = "0.0.0.0/0"
      port        = 22
      port_end    = 22
      action      = "ACCEPT"
      comment     = "Allow SSH"
    }
  ]
}
`, rName)
}
