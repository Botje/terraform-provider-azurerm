// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package containers

import (
	"context"
	"time"

	"github.com/hashicorp/go-azure-sdk/resource-manager/containerregistry/2023-07-01/credentialsets"
	"github.com/hashicorp/go-azure-sdk/resource-manager/containerregistry/2023-07-01/registries"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/keyvault/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

var _ sdk.Resource = ContainerRegistryCredentialSet{}
var _ sdk.ResourceWithUpdate = ContainerRegistryCredentialSet{}

type ContainerRegistryCredentialSet struct{}

type credential struct {
	UserName string `tfschema:"user_key_vault_id"`
	Password string `tfschema:"password_key_vault_id"`
}

type ContainerRegistryCredentialSetModel struct {
	Name                string       `tfschema:"name"`
	ContainerRegistryId string       `tfschema:"container_registry_id"`
	LoginServer         string       `tfschema:"login_server"`
	Credentials         []credential `tfschema:"credential"`
	// Identity    string `tfschema:"identity"`
}

func (c ContainerRegistryCredentialSet) Arguments() map[string]*schema.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "The name of the credential set.",
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"container_registry_id": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "Resource ID of the parent container registry.",
			ValidateFunc: registries.ValidateRegistryID,
		},

		"login_server": {
			Type:         pluginsdk.TypeString,
			Required:     true,
			Description:  "Login server to be used with these credentials.",
			ValidateFunc: validation.StringIsNotEmpty,
		},

		"credential": {
			Type:        pluginsdk.TypeList,
			Required:    true,
			MinItems:    1,
			MaxItems:    2,
			Description: "Credentials. The first credential block is the primary credential, the second block is the secondary credential.",
			Elem: &pluginsdk.Resource{
				Schema: map[string]*schema.Schema{
					"user_key_vault_id": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						Description:  "Key Vault URI holding the username",
						ValidateFunc: validate.NestedItemIdWithOptionalVersion,
					},

					"password_key_vault_id": {
						Type:         pluginsdk.TypeString,
						Required:     true,
						Description:  "Key Vault URI holding the password",
						ValidateFunc: validate.NestedItemIdWithOptionalVersion,
					},
				},
			},
		},
	}
}

func (c ContainerRegistryCredentialSet) ModelObject() interface{} {
	return &ContainerRegistryCredentialSetModel{}
}

func (c ContainerRegistryCredentialSet) ResourceType() string {
	return "azurerm_container_registry_credential_set"
}

func (c ContainerRegistryCredentialSet) Attributes() map[string]*schema.Schema {
	return map[string]*schema.Schema{}
}

func (c ContainerRegistryCredentialSet) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			panic("unimplemented")
		},
	}
}

func (c ContainerRegistryCredentialSet) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			panic("unimplemented")
		},
	}}

func (c ContainerRegistryCredentialSet) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			panic("unimplemented")
		},
	}}

func (c ContainerRegistryCredentialSet) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			panic("unimplemented")
		},
	}}

func (c ContainerRegistryCredentialSet) IDValidationFunc() func(interface{}, string) ([]string, []error) {
	return credentialsets.ValidateCredentialSetID
}
