// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package containers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/containers"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

type KubernetesAutomaticClusterResource struct{}

func TestAccKubernetesAutomaticCluster_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_kubernetes_automatic_cluster", "test")
	r := KubernetesAutomaticClusterResource{}

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

func TestAccKubernetesAutomaticCluster_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_kubernetes_automatic_cluster", "test")
	r := KubernetesAutomaticClusterResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImportConfig(data),
			ExpectError: acceptance.RequiresImportError("azurerm_kubernetes_automatic_cluster"),
		},
	})
}

func TestAccKubernetesAutomaticCluster_tags(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_kubernetes_automatic_cluster", "test")
	r := KubernetesAutomaticClusterResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.tagsConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.tagsUpdatedConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccKubernetesAutomaticCluster_monitor(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_kubernetes_automatic_cluster", "test")
	r := KubernetesAutomaticClusterResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.monitorDisabledConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("monitor.0.metrics_enabled").HasValue("false"),
				check.That(data.ResourceName).Key("monitor.0.container_insights_enabled").HasValue("false"),
			),
		},
		data.ImportStep(),
		{
			Config: r.monitorEnabledConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("monitor.0.metrics_enabled").HasValue("true"),
				check.That(data.ResourceName).Key("monitor.0.container_insights_enabled").HasValue("true"),
				check.That(data.ResourceName).Key("monitor.0.log_analytics_workspace_id").Exists(),
			),
		},
		data.ImportStep(),
		{
			Config: r.monitorDisabledConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("monitor.0.metrics_enabled").HasValue("false"),
				check.That(data.ResourceName).Key("monitor.0.container_insights_enabled").HasValue("false"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccKubernetesAutomaticCluster_microsoftDefender(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_kubernetes_automatic_cluster", "test")
	r := KubernetesAutomaticClusterResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.microsoftDefenderConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("microsoft_defender.#").HasValue("1"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccKubernetesAutomaticCluster_DefaultNginxController(t *testing.T) {
	webAppRoutingIngress := containers.KubernetesAutomaticClusterResource{}.Arguments()["web_app_routing_ingress"]
	nginxController := webAppRoutingIngress.Elem.(*pluginsdk.Resource).Schema["default_nginx_controller"]

	// `None` must be accepted, as it is for `azurerm_kubernetes_cluster`
	for _, value := range []string{"AnnotationControlled", "External", "Internal", "None"} {
		if _, errs := nginxController.ValidateFunc(value, "default_nginx_controller"); len(errs) > 0 {
			t.Fatalf("expected %q to be a valid `default_nginx_controller`, got %+v", value, errs)
		}
	}

	// an omitted `default_nginx_controller` is sent to Azure as `None` and read back as an empty string, so
	// configurations which omit it and configurations which specify `None` must both be diff free
	testData := []struct {
		old      string
		new      string
		suppress bool
	}{
		{old: "", new: "None", suppress: true},
		{old: "None", new: "", suppress: true},
		{old: "", new: "Internal", suppress: false},
		{old: "Internal", new: "None", suppress: false},
		{old: "None", new: "External", suppress: false},
	}

	for _, v := range testData {
		if actual := containers.SuppressDefaultNginxControllerNone("", v.old, v.new, nil); actual != v.suppress {
			t.Fatalf("expected the difference between %q and %q to be suppressed: %t, got %t", v.old, v.new, v.suppress, actual)
		}
	}
}

func TestAccKubernetesAutomaticCluster_webAppRoutingIngressNginxNone(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_kubernetes_automatic_cluster", "test")
	r := KubernetesAutomaticClusterResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.webAppRoutingIngressNginxNoneConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.webAppRoutingIngressNginxInternalConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("web_app_routing_ingress.0.default_nginx_controller").HasValue("Internal"),
			),
		},
		data.ImportStep(),
		{
			Config: r.webAppRoutingIngressNginxNoneConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func (t KubernetesAutomaticClusterResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := commonids.ParseKubernetesClusterIDInsensitively(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Containers.KubernetesClustersClient_v2026_04_01.Get(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("reading Kubernetes Cluster (%s): %+v", id.String(), err)
	}

	return pointer.To(resp.Model != nil && resp.Model.Id != nil), nil
}

func (KubernetesAutomaticClusterResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-aks-%d"
  location = "%s"
}

resource "azurerm_kubernetes_automatic_cluster" "test" {
  name                = "acctestaks%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  identity {
    type = "SystemAssigned"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r KubernetesAutomaticClusterResource) requiresImportConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_kubernetes_automatic_cluster" "import" {
  name                = azurerm_kubernetes_automatic_cluster.test.name
  location            = azurerm_kubernetes_automatic_cluster.test.location
  resource_group_name = azurerm_kubernetes_automatic_cluster.test.resource_group_name

  identity {
    type = "SystemAssigned"
  }
}
`, r.basic(data))
}

func (KubernetesAutomaticClusterResource) tagsConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-aks-%d"
  location = "%s"
}

resource "azurerm_kubernetes_automatic_cluster" "test" {
  name                = "acctestaks%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  identity {
    type = "SystemAssigned"
  }

  tags = {
    dimension = "C-137"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (KubernetesAutomaticClusterResource) tagsUpdatedConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-aks-%d"
  location = "%s"
}

resource "azurerm_kubernetes_automatic_cluster" "test" {
  name                = "acctestaks%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  identity {
    type = "SystemAssigned"
  }

  tags = {
    dimension = "D-99"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (KubernetesAutomaticClusterResource) monitorDisabledConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-aks-%d"
  location = "%s"
}

resource "azurerm_kubernetes_automatic_cluster" "test" {
  name                = "acctestaks%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  identity {
    type = "SystemAssigned"
  }

  monitor {
    metrics_enabled            = false
    container_insights_enabled = false
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (KubernetesAutomaticClusterResource) monitorEnabledConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-aks-%[1]d"
  location = "%[2]s"
}

resource "azurerm_log_analytics_workspace" "test" {
  name                = "acctestLAW-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_kubernetes_automatic_cluster" "test" {
  name                = "acctestaks%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  identity {
    type = "SystemAssigned"
  }

  monitor {
    metrics_enabled            = true
    container_insights_enabled = true
    log_analytics_workspace_id = azurerm_log_analytics_workspace.test.id
  }
}
`, data.RandomInteger, data.Locations.Primary)
}

func (KubernetesAutomaticClusterResource) webAppRoutingIngressNginxNoneConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-aks-%[1]d"
  location = "%[2]s"
}

resource "azurerm_kubernetes_automatic_cluster" "test" {
  name                = "acctestaks%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  identity {
    type = "SystemAssigned"
  }

  web_app_routing_ingress {
    default_nginx_controller = "None"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}

func (KubernetesAutomaticClusterResource) webAppRoutingIngressNginxInternalConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-aks-%[1]d"
  location = "%[2]s"
}

resource "azurerm_kubernetes_automatic_cluster" "test" {
  name                = "acctestaks%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  identity {
    type = "SystemAssigned"
  }

  web_app_routing_ingress {
    default_nginx_controller = "Internal"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}

func (KubernetesAutomaticClusterResource) microsoftDefenderConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-aks-%[1]d"
  location = "%[2]s"
}

resource "azurerm_log_analytics_workspace" "test" {
  name                = "acctestLAW-%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_kubernetes_automatic_cluster" "test" {
  name                = "acctestaks%[1]d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  identity {
    type = "SystemAssigned"
  }

  microsoft_defender {
    log_analytics_workspace_id = azurerm_log_analytics_workspace.test.id
  }

  private_cluster {
    private_dns_zone_id = "System"
  }
}
`, data.RandomInteger, data.Locations.Primary)
}
