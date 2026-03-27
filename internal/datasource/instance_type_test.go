package datasource_test

import (
	"fmt"
	"os"
	"testing"

	"terraform-provider-eci/internal/acctest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInstanceTypeDataSource_basic(t *testing.T) {
	instanceTypeName := os.Getenv("ECI_TEST_INSTANCE_TYPE_NAME")
	if instanceTypeName == "" {
		t.Skip("ECI_TEST_INSTANCE_TYPE_NAME not set")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypeDataSourceConfig_basic(instanceTypeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.eci_instance_type.test", "id"),
					resource.TestCheckResourceAttr("data.eci_instance_type.test", "name", instanceTypeName),
					resource.TestCheckResourceAttrSet("data.eci_instance_type.test", "cpu_vcore"),
					resource.TestCheckResourceAttrSet("data.eci_instance_type.test", "memory_gib"),
				),
			},
		},
	})
}

func testAccInstanceTypeDataSourceConfig_basic(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
data "eci_instance_type" "test" {
  name = %[1]q
}
`, name)
}
