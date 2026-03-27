package datasource_test

import (
	"fmt"
	"os"
	"testing"

	"terraform-provider-eci/internal/acctest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBlockStorageImageDataSource_basic(t *testing.T) {
	imageName := os.Getenv("ECI_TEST_IMAGE_NAME")
	if imageName == "" {
		t.Skip("ECI_TEST_IMAGE_NAME not set")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBlockStorageImageDataSourceConfig_basic(imageName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.eci_block_storage_image.test", "id"),
					resource.TestCheckResourceAttr("data.eci_block_storage_image.test", "name", imageName),
					resource.TestCheckResourceAttrSet("data.eci_block_storage_image.test", "size_gib"),
				),
			},
		},
	})
}

func testAccBlockStorageImageDataSourceConfig_basic(name string) string {
	return acctest.ProviderConfig() + fmt.Sprintf(`
data "eci_block_storage_image" "test" {
  name = %[1]q
}
`, name)
}
