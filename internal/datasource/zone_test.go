package datasource_test

import (
	"fmt"
	"os"
	"testing"

	"terraform-provider-eci/internal/acctest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccZoneDataSource_basic(t *testing.T) {
	regionName := os.Getenv("ECI_TEST_REGION_NAME")
	zoneName := os.Getenv("ECI_TEST_ZONE_NAME")
	if regionName == "" || zoneName == "" {
		t.Skip("ECI_TEST_REGION_NAME and ECI_TEST_ZONE_NAME must be set")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDataSourceConfig_basic(regionName, zoneName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.eci_zone.test", "id"),
					resource.TestCheckResourceAttr("data.eci_zone.test", "name", zoneName),
				),
			},
		},
	})
}

func testAccZoneDataSourceConfig_basic(regionName, zoneName string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
data "eci_region" "test" {
  name = %[1]q
}

data "eci_zone" "test" {
  name      = %[2]q
  region_id = data.eci_region.test.id
}
`, regionName, zoneName)
}
