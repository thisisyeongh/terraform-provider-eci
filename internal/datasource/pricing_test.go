package datasource_test

import (
	"fmt"
	"os"
	"testing"

	"terraform-provider-eci/internal/acctest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPricingDataSource_basic(t *testing.T) {
	pricingName := os.Getenv("ECI_TEST_PRICING_NAME")
	pricingType := os.Getenv("ECI_TEST_PRICING_TYPE")
	if pricingName == "" || pricingType == "" {
		t.Skip("ECI_TEST_PRICING_NAME and ECI_TEST_PRICING_TYPE must be set")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPricingDataSourceConfig_basic(pricingName, pricingType),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.eci_pricing.test", "id"),
					resource.TestCheckResourceAttr("data.eci_pricing.test", "name", pricingName),
					resource.TestCheckResourceAttrSet("data.eci_pricing.test", "price_per_hour"),
				),
			},
		},
	})
}

func testAccPricingDataSourceConfig_basic(name, pricingType string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
data "eci_pricing" "test" {
  name         = %[1]q
  pricing_type = %[2]q
}
`, name, pricingType)
}
