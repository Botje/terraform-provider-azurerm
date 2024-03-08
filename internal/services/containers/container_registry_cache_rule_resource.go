// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package containers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-sdk/resource-manager/containerregistry/2023-07-01/cacherules"
	"github.com/hashicorp/go-azure-sdk/resource-manager/containerregistry/2023-07-01/registries"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
)

var _ sdk.Resource = ContainerRegistryCacheRule{}
var _ sdk.ResourceWithUpdate = ContainerRegistryCacheRule{}

type ContainerRegistryCacheRule struct{}

func (ContainerRegistryCacheRule) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:        pluginsdk.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The name of the cache rule.",
			ValidateFunc: validation.StringLenBetween(5, 40),
		},

		"credential_set_id": {
			Type:        pluginsdk.TypeString,
			Optional:    true,
			Description: "Credential set ID",
		},

		"container_registry_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "Resource ID of the parent container registry.",
			ValidateFunc: registries.ValidateRegistryID,
		},

		"source_repo": {
			Type:        pluginsdk.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The full source repository path such as 'docker.io/library/ubuntu'.",
		},

		"target_repo": {
			Type:        pluginsdk.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The target repository namespace such as 'ubuntu'.",
		},
	}
}

func (ContainerRegistryCacheRule) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (ContainerRegistryCacheRule) ModelObject() interface{} {
	return nil
}

func (ContainerRegistryCacheRule) ResourceType() string {
	return "azurerm_container_registry_cache_rule"
}

func (r ContainerRegistryCacheRule) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			cacheRulesClient := metadata.Client.Containers.ContainerRegistryClient_v2023_07_01.CacheRules
			registryClient := metadata.Client.Containers.ContainerRegistryClient_v2023_07_01.Registries
			subscriptionId := metadata.Client.Account.SubscriptionId
			ctx, cancel := timeouts.ForCreate(metadata.Client.StopContext, metadata.ResourceData)

			defer cancel()
			log.Printf("[INFO] preparing arguments for Container Registry Cache Rule creation.")

			registryId, err := registries.ParseRegistryID(metadata.ResourceData.Get("container_registry_id").(string))
			if err != nil {
				return err
			}

			_, err = registryClient.Get(ctx, *registryId)
			if err != nil {
				return fmt.Errorf("getting registry %s: %+v", registryId, err)
			}

			id := cacherules.NewCacheRuleID(subscriptionId,
				registryId.ResourceGroupName,
				registryId.RegistryName,
				metadata.ResourceData.Get("name").(string),
			)

			existing, err := cacheRulesClient.Get(ctx, id)

			if err != nil {
				if !response.WasNotFound(existing.HttpResponse) {
					return fmt.Errorf("checking for presence of existing %s: %s", id, err)
				}
			}

			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			// TODO: make a check that the repo is available in the registry.
			targetRepo := metadata.ResourceData.Get("target_repo").(string)

			// TODO: validate the source repo.
			sourceRepo := metadata.ResourceData.Get("source_repo").(string)

			parameters := cacherules.CacheRule{
				Name: &id.CacheRuleName,
				Properties: &cacherules.CacheRuleProperties{
					SourceRepository: &sourceRepo,
					TargetRepository: &targetRepo,
				},
			}

			if err := cacheRulesClient.CreateThenPoll(ctx, id, parameters); err != nil {
				return fmt.Errorf("creating Container Registry Cache Rule %s: %+v", id, err)
			}

			metadata.SetID(id)

			return nil
		},
	}
}

func (ContainerRegistryCacheRule) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			cacheRulesClient := metadata.Client.Containers.ContainerRegistryClient_v2023_07_01.CacheRules
			ctx, cancel := timeouts.ForRead(metadata.Client.StopContext, metadata.ResourceData)
			defer cancel()

			id, err := cacherules.ParseCacheRuleID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			resp, err := cacheRulesClient.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(resp.HttpResponse) {
					log.Printf("[DEBUG] Container Registry Cache Rule %s was not found.", *id)
					return metadata.MarkAsGone(id)
				}

				return fmt.Errorf("retrieving Container Registry Cache Rule %s: %+v", *id, err)
			}

			metadata.ResourceData.Set("name", id.CacheRuleName)
			registryId := registries.NewRegistryID(id.SubscriptionId, id.SubscriptionId, id.RegistryName)
			metadata.ResourceData.Set("container_registry_id", registryId)

			if model := resp.Model; model != nil {
				if properties := model.Properties; properties != nil {
					metadata.ResourceData.Set("source_repo", properties.SourceRepository)
					metadata.ResourceData.Set("target_repo", properties.TargetRepository)
				}
			}

			return nil
		},
	}
}

func (r ContainerRegistryCacheRule) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			cacheRulesClient := metadata.Client.Containers.ContainerRegistryClient_v2023_07_01.CacheRules
			ctx, cancel := timeouts.ForRead(metadata.Client.StopContext, metadata.ResourceData)

			defer cancel()
			log.Printf("[INFO] preparing arguments for Container Registry Cache Rule update.")

			id, err := cacherules.ParseCacheRuleID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			// TODO: You can only update the credential set. To be implemented
			parameters := cacherules.CacheRuleUpdateParameters{
				Properties: &cacherules.CacheRuleUpdateProperties{},
			}

			if err := cacheRulesClient.UpdateThenPoll(ctx, *id, parameters); err != nil {
				return fmt.Errorf("updating %s: %+v", id, err)
			}

			metadata.SetID(id)

			return nil
		},
	}
}

func (ContainerRegistryCacheRule) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			cacheRulesClient := metadata.Client.Containers.ContainerRegistryClient_v2023_07_01.CacheRules
			ctx, cancel := timeouts.ForDelete(metadata.Client.StopContext, metadata.ResourceData)
			defer cancel()

			id, err := cacherules.ParseCacheRuleID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			if err := cacheRulesClient.DeleteThenPoll(ctx, *id); err != nil {
				return fmt.Errorf("deleting Container Registry Cache Rule %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (ContainerRegistryCacheRule) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return cacherules.ValidateCacheRuleID
}
