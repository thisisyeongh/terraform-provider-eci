package datasource_test

import (
	"fmt"
	"os"
	"testing"

	"terraform-provider-eci/internal/acctest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRegionDataSource_basic(t *testing.T) {
	regionName := os.Getenv("ECI_TEST_REGION_NAME")
	if regionName == "" {
		t.Skip("ECI_TEST_REGION_NAME not set")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionDataSourceConfig_basic(regionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.eci_region.test", "id"),
					resource.TestCheckResourceAttr("data.eci_region.test", "name", regionName),
				),
			},
		},
	})
}

func testAccRegionDataSourceConfig_basic(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
data "eci_region" "test" {
  name = %[1]q
}
`, name)
}
