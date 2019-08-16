// example: azure:///subscriptions/8643025a-c059-4a48-85d0-d76f51d63a74/resourceGroups/shoshanargwestus2/providers/Microsoft.Compute/virtualMachineScaleSets/default-9c6ddaaa-f207fc77/virtualMachines/3

package azure

import (
	"github.com/Azure/go-autorest/autorest/azure"
)

func ParseProviderID(providerID string) (azure.Resource, error) {
	return azure.ParseResourceID(providerID)
}
