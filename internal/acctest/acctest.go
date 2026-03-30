package acctest

import (
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"

	"terraform-provider-eci/internal/api"
	"terraform-provider-eci/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

const (
	ResourcePrefix = "tf-test"
	TestPassword   = "T3st!xKq8@mW"
)

var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"eci": providerserver.NewProtocol6WithError(provider.New("test", false)()),
}

var (
	sharedClient     *api.APIClient
	sharedClientOnce sync.Once
	sharedClientErr  error
)

func PreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("ECI_API_ACCESS_TOKEN") == "" {
		t.Fatal("ECI_API_ACCESS_TOKEN must be set")
	}
	if os.Getenv("ECI_API_ENDPOINT") == "" {
		t.Fatal("ECI_API_ENDPOINT must be set")
	}
	if os.Getenv("ECI_ZONE_ID") == "" {
		t.Fatal("ECI_ZONE_ID must be set")
	}
}

// SharedClient returns a shared API client for use in CheckDestroy and sweepers.
// The client is initialized once and reused across all tests.
func SharedClient() (*api.APIClient, error) {
	sharedClientOnce.Do(func() {
		endpoint := os.Getenv("ECI_API_ENDPOINT")
		token := os.Getenv("ECI_API_ACCESS_TOKEN")
		zoneId := os.Getenv("ECI_ZONE_ID")

		if endpoint == "" || token == "" || zoneId == "" {
			sharedClientErr = fmt.Errorf(
				"ECI_API_ACCESS_TOKEN, ECI_API_ENDPOINT, and ECI_ZONE_ID must be set",
			)
			return
		}

		parsedURL, err := url.Parse(endpoint)
		if err != nil {
			sharedClientErr = fmt.Errorf("failed to parse ECI_API_ENDPOINT: %w", err)
			return
		}

		pathPrefix := parsedURL.Path
		parsedURL.Path = ""
		parsedURL.RawPath = ""

		sharedClient, sharedClientErr = api.NewAPIClient(
			token, parsedURL.String(), pathPrefix, zoneId, false,
		)
	})
	return sharedClient, sharedClientErr
}

func RandomName(prefix string) string {
	return fmt.Sprintf("%s-%s-%s", ResourcePrefix, prefix, sdkacctest.RandString(8))
}

func ProviderConfig() string {
	return fmt.Sprintf(`
provider "eci" {
  api_access_token = %[1]q
  api_endpoint     = %[2]q
  zone_id          = %[3]q
}
`, os.Getenv("ECI_API_ACCESS_TOKEN"), os.Getenv("ECI_API_ENDPOINT"), os.Getenv("ECI_ZONE_ID"))
}
