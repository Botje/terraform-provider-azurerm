package containers_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/services/containers/validate"
)

type ContainerRegistryCacheRule struct{}

func TestAccContainerRegistryCacheRuleName_validation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "",
			ErrCount: 2,
		},
		{
			Value:    "four",
			ErrCount: 1,
		},
		{
			Value:    "5five",
			ErrCount: 0,
		},
		{
			Value:    "hello-world",
			ErrCount: 0,
		},
		{
			Value:    "hello-world-foo-bar-12345",
			ErrCount: 0,
		},
		{
			Value:    "hello_world",
			ErrCount: 1,
		},
		{
			Value:    "helloWorld",
			ErrCount: 0,
		},
		{
			Value:    "helloworld12",
			ErrCount: 0,
		},
		{
			Value:    "hello@world",
			ErrCount: 1,
		},
		{
			Value:    "qfvbdsbvipqdbwsbddbdcwqffewsqwcdw21ddwqwd3324120",
			ErrCount: 0,
		},
		{
			Value:    "qfvbdsbvipqdbwsbddbdcwqffewsqwcdw21ddwqwd33241202",
			ErrCount: 0,
		},
		{
			Value:    "qfvbdsbvipqdbwsbddbdcwqfjjfewsqwcdw21ddwqwd3324120",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validate.ContainerRegistryCacheRuleName(tc.Value, "azurerm_container_registry_cache_rule")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the Azure RM Container Registry Cache Rule Name to trigger a validation error: %v", errors)
		}
	}
}
