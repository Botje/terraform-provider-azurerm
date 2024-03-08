package containers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-sdk/resource-manager/containerregistry/2023-07-01/cacherules"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type ContainerRegistryCacheRuleResource struct{}

func TestAccContainerRegistryCacheRule_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_container_registry_cache_rule", "test")
	r := ContainerRegistryCacheRuleResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccContainerRegistryCacheRule_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_container_registry_cache_rule", "test")
	r := ContainerRegistryCacheRuleResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_container_registry_cache_rule"),
		},
	})
}

func (t ContainerRegistryCacheRuleResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := cacherules.ParseCacheRuleID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Containers.ContainerRegistryClient_v2023_07_01.CacheRules.Get(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("retrieving %s: %+v", id, err)
	}

	return utils.Bool(resp.Model != nil), nil
}

func (ContainerRegistryCacheRuleResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
	features {}
}

resource "azurerm_resource_group" "test" {
	name     = "accTestRG-acr-cache-rule-%[1]d"
	location = "%[2]s"
}

resource "azurerm_container_registry" "test" {
  name                = "testacccr%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "Basic"
}

resource "azurerm_container_registry_cache_rule" "test" {
  name                = "testacc-cr-cache-rule-%[1]d"
  container_registry_id = azurerm_container_registry.test.id
  target_repo         = "target"
  source_repo         = "docker.io/hello-world"
}
`, data.RandomInteger, data.Locations.Primary)
}

func (r ContainerRegistryCacheRuleResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_container_registry_cache_rule" "import" {
  name                = azurerm_container_registry_cache_rule.test.name
  resource_group_name = azurerm_container_registry_cache_rule.test.container_registry_id
  target_repo         = azurerm_container_registry_cache_rule.test.target_repo
  source_repo         = azurerm_container_registry_cache_rule.test.source_repo
}
`, r.basic(data))
}
