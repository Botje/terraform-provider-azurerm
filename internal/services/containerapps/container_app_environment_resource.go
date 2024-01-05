// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package containerapps

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/location"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/tags"
	"github.com/hashicorp/go-azure-sdk/resource-manager/containerapps/2023-05-01/managedenvironments"
	"github.com/hashicorp/go-azure-sdk/resource-manager/operationalinsights/2020-08-01/workspaces"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/containerapps/helpers"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/containerapps/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

const (
	LogDestinationMissing      = ""
	LogDestinationNone         = "none"
	LogDestinationLogAnalytics = "log-analytics"
	LogDestinationAzureMonitor = "azure-monitor"
)

type ContainerAppEnvironmentResource struct{}

type ContainerAppEnvironmentModel struct {
	Name                                    string                         `tfschema:"name"`
	ResourceGroup                           string                         `tfschema:"resource_group_name"`
	Location                                string                         `tfschema:"location"`
	DaprApplicationInsightsConnectionString string                         `tfschema:"dapr_application_insights_connection_string"`
	LogDestination                          string                         `tfschema:"log_destination"`
	LogAnalyticsWorkspaceId                 string                         `tfschema:"log_analytics_workspace_id"`
	InfrastructureSubnetId                  string                         `tfschema:"infrastructure_subnet_id"`
	InternalLoadBalancerEnabled             bool                           `tfschema:"internal_load_balancer_enabled"`
	ZoneRedundant                           bool                           `tfschema:"zone_redundancy_enabled"`
	Tags                                    map[string]interface{}         `tfschema:"tags"`
	WorkloadProfiles                        []helpers.WorkloadProfileModel `tfschema:"workload_profile"`

	DefaultDomain         string `tfschema:"default_domain"`
	DockerBridgeCidr      string `tfschema:"docker_bridge_cidr"`
	PlatformReservedCidr  string `tfschema:"platform_reserved_cidr"`
	PlatformReservedDnsIP string `tfschema:"platform_reserved_dns_ip_address"`
	StaticIP              string `tfschema:"static_ip_address"`
}

var _ sdk.ResourceWithUpdate = ContainerAppEnvironmentResource{}
var _ sdk.ResourceWithCustomizeDiff = ContainerAppEnvironmentResource{}

func (r ContainerAppEnvironmentResource) ModelObject() interface{} {
	return &ContainerAppEnvironmentModel{}
}

func (r ContainerAppEnvironmentResource) ResourceType() string {
	return "azurerm_container_app_environment"
}

func (r ContainerAppEnvironmentResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return managedenvironments.ValidateManagedEnvironmentID
}

func (r ContainerAppEnvironmentResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.ManagedEnvironmentName,
			Description:  "The name of the Container Apps Managed Environment.",
		},

		"location": commonschema.Location(),

		"resource_group_name": commonschema.ResourceGroupName(),

		"dapr_application_insights_connection_string": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ForceNew:     true,
			Sensitive:    true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  "Application Insights connection string used by Dapr to export Service to Service communication telemetry.",
		},

		"log_destination": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{LogDestinationNone, LogDestinationLogAnalytics, LogDestinationAzureMonitor}, false),
			Description:  "Destination for application logs. Can be set to `none` for no logging , set to `log-analytics` (see `log_analytics_workspace_id`) or set to `azure-monitor`.",
		},

		"log_analytics_workspace_id": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ValidateFunc: workspaces.ValidateWorkspaceID,
			Description:  "The ID for the Log Analytics Workspace to link this Container Apps Managed Environment to.",
		},

		"infrastructure_subnet_id": {
			Type:         pluginsdk.TypeString,
			Optional:     true,
			ForceNew:     true,
			ValidateFunc: commonids.ValidateSubnetID,
			Description:  "The existing Subnet to use for the Container Apps Control Plane. **NOTE:** The Subnet must have a `/21` or larger address space.",
		},

		"internal_load_balancer_enabled": {
			Type:         pluginsdk.TypeBool,
			Optional:     true,
			ForceNew:     true,
			Default:      false,
			RequiredWith: []string{"infrastructure_subnet_id"},
			Description:  "Should the Container Environment operate in Internal Load Balancing Mode? Defaults to `false`. **Note:** can only be set to `true` if `infrastructure_subnet_id` is specified.",
		},

		"workload_profile": helpers.WorkloadProfileSchema(),

		"zone_redundancy_enabled": {
			Type:         pluginsdk.TypeBool,
			Optional:     true,
			ForceNew:     true,
			Default:      false,
			RequiredWith: []string{"infrastructure_subnet_id"},
		},

		"tags": commonschema.Tags(),
	}
}

func (r ContainerAppEnvironmentResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"default_domain": {
			Type:        pluginsdk.TypeString,
			Computed:    true,
			Description: "The default publicly resolvable name of this Container App Environment",
		},

		"docker_bridge_cidr": {
			Type:        pluginsdk.TypeString,
			Computed:    true,
			Description: "The network addressing in which the Container Apps in this Container App Environment will reside in CIDR notation.",
		},

		"platform_reserved_cidr": {
			Type:        pluginsdk.TypeString,
			Computed:    true,
			Description: "The IP range, in CIDR notation, that is reserved for environment infrastructure IP addresses.",
		},

		"platform_reserved_dns_ip_address": {
			Type:        pluginsdk.TypeString,
			Computed:    true,
			Description: "The IP address from the IP range defined by `platform_reserved_cidr` that is reserved for the internal DNS server.",
		},

		"static_ip_address": {
			Type:        pluginsdk.TypeString,
			Computed:    true,
			Description: "The Static IP Address of the Environment.",
		},
	}
}

func (r ContainerAppEnvironmentResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.ContainerApps.ManagedEnvironmentClient
			logAnalyticsClient := metadata.Client.LogAnalytics.SharedKeyWorkspacesClient
			subscriptionId := metadata.Client.Account.SubscriptionId

			var containerAppEnvironment ContainerAppEnvironmentModel

			if err := metadata.Decode(&containerAppEnvironment); err != nil {
				return err
			}

			id := managedenvironments.NewManagedEnvironmentID(subscriptionId, containerAppEnvironment.ResourceGroup, containerAppEnvironment.Name)

			existing, err := client.Get(ctx, id)
			if err != nil {
				if !response.WasNotFound(existing.HttpResponse) {
					return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
				}
			}

			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			managedEnvironment := managedenvironments.ManagedEnvironment{
				Location: containerAppEnvironment.Location,
				Name:     pointer.To(containerAppEnvironment.Name),
				Properties: &managedenvironments.ManagedEnvironmentProperties{
					VnetConfiguration: &managedenvironments.VnetConfiguration{},
					ZoneRedundant:     pointer.To(containerAppEnvironment.ZoneRedundant),
				},
				Tags: tags.Expand(containerAppEnvironment.Tags),
			}

			if containerAppEnvironment.DaprApplicationInsightsConnectionString != "" {
				managedEnvironment.Properties.DaprAIConnectionString = pointer.To(containerAppEnvironment.DaprApplicationInsightsConnectionString)
			}

			appLogsConfiguration, err := constructAppLogsConfigurationFromModel(ctx, &id, containerAppEnvironment, logAnalyticsClient)
			if err != nil {
				return err
			}
			managedEnvironment.Properties.AppLogsConfiguration = appLogsConfiguration

			if containerAppEnvironment.InfrastructureSubnetId != "" {
				managedEnvironment.Properties.VnetConfiguration.InfrastructureSubnetId = pointer.To(containerAppEnvironment.InfrastructureSubnetId)
				managedEnvironment.Properties.VnetConfiguration.Internal = pointer.To(containerAppEnvironment.InternalLoadBalancerEnabled)
			}

			managedEnvironment.Properties.WorkloadProfiles = helpers.ExpandWorkloadProfiles(containerAppEnvironment.WorkloadProfiles)

			if err := client.CreateOrUpdateThenPoll(ctx, id, managedEnvironment); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
	}
}

func (r ContainerAppEnvironmentResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.ContainerApps.ManagedEnvironmentClient
			id, err := managedenvironments.ParseManagedEnvironmentID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			existing, err := client.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(existing.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("reading %s: %+v", *id, err)
			}

			var state ContainerAppEnvironmentModel

			if model := existing.Model; model != nil {
				state.Name = id.ManagedEnvironmentName
				state.ResourceGroup = id.ResourceGroupName
				state.Location = location.Normalize(model.Location)
				state.Tags = tags.Flatten(model.Tags)

				if props := model.Properties; props != nil {
					if vnet := props.VnetConfiguration; vnet != nil {
						state.InfrastructureSubnetId = pointer.From(vnet.InfrastructureSubnetId)
						state.InternalLoadBalancerEnabled = pointer.From(vnet.Internal)
						state.DockerBridgeCidr = pointer.From(vnet.DockerBridgeCidr)
						state.PlatformReservedCidr = pointer.From(vnet.PlatformReservedCidr)
						state.PlatformReservedDnsIP = pointer.From(vnet.PlatformReservedDnsIP)
					}

					state.ZoneRedundant = pointer.From(props.ZoneRedundant)
					state.StaticIP = pointer.From(props.StaticIP)
					state.DefaultDomain = pointer.From(props.DefaultDomain)
					state.WorkloadProfiles = helpers.FlattenWorkloadProfiles(props.WorkloadProfiles)
					if appsLogs := props.AppLogsConfiguration; appsLogs != nil && appsLogs.Destination != nil {
						state.LogDestination = pointer.From(appsLogs.Destination)

						if state.LogDestination == LogDestinationLogAnalytics {
							// Reading in log_analytics_workspace_id is not possible, so reading from config. Import will need to ignore_changes unfortunately
							if v := metadata.ResourceData.Get("log_analytics_workspace_id").(string); v != "" {
								state.LogAnalyticsWorkspaceId = v
							}
						}

					} else {
						state.LogDestination = LogDestinationNone
					}
				}
			}

			// `dapr_application_insights_connection_string` is sensitive and not returned by API
			if v := metadata.ResourceData.Get("dapr_application_insights_connection_string").(string); v != "" {
				state.DaprApplicationInsightsConnectionString = v
			}

			if err := metadata.Encode(&state); err != nil {
				return fmt.Errorf("encoding: %+v", err)
			}

			return nil
		},
	}
}

func (r ContainerAppEnvironmentResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.ContainerApps.ManagedEnvironmentClient
			id, err := managedenvironments.ParseManagedEnvironmentID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			if err = client.DeleteThenPoll(ctx, *id); err != nil {
				return fmt.Errorf("deleting %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (r ContainerAppEnvironmentResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.ContainerApps.ManagedEnvironmentClient
			logAnalyticsClient := metadata.Client.LogAnalytics.SharedKeyWorkspacesClient
			id, err := managedenvironments.ParseManagedEnvironmentID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var state ContainerAppEnvironmentModel
			if err := metadata.Decode(&state); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			existing, err := client.Get(ctx, *id)
			if err != nil {
				return fmt.Errorf("reading %s: %+v", *id, err)
			}

			if metadata.ResourceData.HasChange("tags") {
				existing.Model.Tags = tags.Expand(state.Tags)
			}

			if state.LogDestination == LogDestinationMissing {
				if state.LogAnalyticsWorkspaceId != "" {
					state.LogDestination = LogDestinationLogAnalytics
				} else {
					state.LogDestination = LogDestinationNone
				}
			}

			appLogDestination := LogDestinationNone
			if appLogs := existing.Model.Properties.AppLogsConfiguration; appLogs != nil && appLogs.Destination != nil {
				appLogDestination = *appLogs.Destination
			}

			if state.LogDestination == appLogDestination {
				if state.LogDestination == LogDestinationLogAnalytics {
					// (@jackofallops) This is not updatable and needs to be removed since the read does not return the sensitive Key field.
					// Whilst not ideal, this means we don't need to try and retrieve it again just to send a no-op.
					existing.Model.Properties.AppLogsConfiguration = nil
				} else {
					existing.Model.Properties.AppLogsConfiguration = nil
				}
			} else if state.LogDestination == LogDestinationNone {
				existing.Model.Properties.AppLogsConfiguration = pointer.To(managedenvironments.AppLogsConfiguration{Destination: nil})
			} else {
				appLogsFromState, err := constructAppLogsConfigurationFromModel(ctx, id, state, logAnalyticsClient)
				if err != nil {
					return err
				}
				existing.Model.Properties.AppLogsConfiguration = appLogsFromState
			}

			if err := client.CreateOrUpdateThenPoll(ctx, *id, *existing.Model); err != nil {
				return fmt.Errorf("updating %s: %+v", id, err)
			}

			return nil
		},
	}
}

func (r ContainerAppEnvironmentResource) CustomizeDiff() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			var model ContainerAppEnvironmentModel
			if err := metadata.DecodeDiff(&model); err != nil {
				return fmt.Errorf("DecodeDiff: %+v", err)
			}

			log_destination, exists := metadata.ResourceDiff.GetOk("log_destination")

			if exists && model.LogAnalyticsWorkspaceId != "" && log_destination != LogDestinationLogAnalytics {
				return fmt.Errorf("log_analytics_workspace_id is set for %s but log_destination is set to '%s' instead of 'log-analytics'", model.Name, model.LogDestination)
			}

			return nil
		},
	}
}

func constructAppLogsConfigurationFromModel(ctx context.Context, id *managedenvironments.ManagedEnvironmentId, model ContainerAppEnvironmentModel, logAnalyticsClient *workspaces.WorkspacesClient) (*managedenvironments.AppLogsConfiguration, error) {
	switch model.LogDestination {
	case LogDestinationNone:
		return nil, nil
	case LogDestinationAzureMonitor:
		return pointer.To(managedenvironments.AppLogsConfiguration{
			Destination: pointer.To(LogDestinationAzureMonitor),
		}), nil
	case LogDestinationLogAnalytics:
		{
			logAnalyticsId, err := workspaces.ParseWorkspaceID(model.LogAnalyticsWorkspaceId)
			if err != nil {
				return nil, err
			}

			workspace, err := logAnalyticsClient.Get(ctx, *logAnalyticsId)
			if err != nil {
				return nil, fmt.Errorf("retrieving %s for %s: %+v", logAnalyticsId, id, err)
			}

			if workspace.Model == nil || workspace.Model.Properties == nil {
				return nil, fmt.Errorf("reading customer ID from %s", logAnalyticsId)
			}

			if workspace.Model.Properties.CustomerId == nil {
				return nil, fmt.Errorf("reading customer ID from %s, `customer_id` is nil", logAnalyticsId)
			}

			keys, err := logAnalyticsClient.SharedKeysGetSharedKeys(ctx, *logAnalyticsId)
			if err != nil {
				return nil, fmt.Errorf("retrieving access keys to %s for %s: %+v", logAnalyticsId, id, err)
			}
			if keys.Model.PrimarySharedKey == nil {
				return nil, fmt.Errorf("reading shared key for %s in %s", logAnalyticsId, id)
			}

			return pointer.To(managedenvironments.AppLogsConfiguration{
				Destination: pointer.To(LogDestinationLogAnalytics),
				LogAnalyticsConfiguration: &managedenvironments.LogAnalyticsConfiguration{
					CustomerId: workspace.Model.Properties.CustomerId,
					SharedKey:  keys.Model.PrimarySharedKey,
				},
			}), nil
		}
	default:
		return nil, fmt.Errorf("invalid log destination %s", model.LogDestination)
	}
}
