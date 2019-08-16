// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package azure

import (
	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	kvmgmt "github.com/Azure/azure-sdk-for-go/services/keyvault/mgmt/2018-02-14/keyvault"
	"github.com/Azure/azure-sdk-for-go/services/msi/mgmt/2018-11-30/msi"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-02-01/network"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"
)

const userAgent string = "genesys"

func NewResourceGroupClient(subID string) (resources.GroupsClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return resources.GroupsClient{}, err
	}
	client := resources.NewGroupsClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return resources.GroupsClient{}, err
	}
	return client, nil
}

func NewVNETClient(subID string) (network.VirtualNetworksClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return network.VirtualNetworksClient{}, err
	}
	client := network.NewVirtualNetworksClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return network.VirtualNetworksClient{}, err
	}
	return client, nil
}

func NewNSGClient(subID string) (network.SecurityGroupsClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return network.SecurityGroupsClient{}, err
	}
	client := network.NewSecurityGroupsClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return network.SecurityGroupsClient{}, err
	}
	return client, nil
}

func NewRouteTableClient(subID string) (network.RouteTablesClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return network.RouteTablesClient{}, err
	}
	client := network.NewRouteTablesClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return network.RouteTablesClient{}, err
	}
	return client, nil
}

func NewAvailabilitySetClient(subID string) (compute.AvailabilitySetsClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return compute.AvailabilitySetsClient{}, err
	}
	client := compute.NewAvailabilitySetsClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return compute.AvailabilitySetsClient{}, err
	}
	return client, nil
}

func NewLoadBalancerClient(subID string) (network.LoadBalancersClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return network.LoadBalancersClient{}, err
	}
	client := network.NewLoadBalancersClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return network.LoadBalancersClient{}, err
	}
	return client, nil
}

func NewSubnetClient(subID string) (network.SubnetsClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return network.SubnetsClient{}, err
	}
	client := network.NewSubnetsClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return network.SubnetsClient{}, err
	}
	return client, nil
}

func NewPublicIPClient(subID string) (network.PublicIPAddressesClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return network.PublicIPAddressesClient{}, err
	}
	client := network.NewPublicIPAddressesClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return network.PublicIPAddressesClient{}, err
	}
	return client, nil
}

func NewNICClient(subID string) (network.InterfacesClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return network.InterfacesClient{}, err
	}
	client := network.NewInterfacesClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return network.InterfacesClient{}, err
	}
	return client, nil
}

func NewVMClient(subID string) (compute.VirtualMachinesClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return compute.VirtualMachinesClient{}, err
	}
	client := compute.NewVirtualMachinesClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return compute.VirtualMachinesClient{}, err
	}
	return client, nil
}

func NewIdentityClient(subID string) (msi.UserAssignedIdentitiesClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return msi.UserAssignedIdentitiesClient{}, err
	}
	client := msi.NewUserAssignedIdentitiesClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return msi.UserAssignedIdentitiesClient{}, err
	}
	return client, nil
}

func NewSignedInUserClient(tenantID string) (graphrbac.SignedInUserClient, error) {
	a, err := injectGraphAuthorizer()
	if err != nil {
		return graphrbac.SignedInUserClient{}, err
	}
	client := graphrbac.NewSignedInUserClient(tenantID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return graphrbac.SignedInUserClient{}, err
	}
	return client, nil
}

func NewServicePrincipalClient(tenantID string) (graphrbac.ServicePrincipalsClient, error) {
	a, err := injectGraphAuthorizer()
	if err != nil {
		return graphrbac.ServicePrincipalsClient{}, err
	}
	client := graphrbac.NewServicePrincipalsClient(tenantID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return graphrbac.ServicePrincipalsClient{}, err
	}
	return client, nil
}

func NewApplicationClient(tenantID string) (graphrbac.ApplicationsClient, error) {
	a, err := injectGraphAuthorizer()
	if err != nil {
		return graphrbac.ApplicationsClient{}, err
	}
	client := graphrbac.NewApplicationsClient(tenantID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return graphrbac.ApplicationsClient{}, err
	}
	return client, nil
}

func NewVaultClient(subID string) (kvmgmt.VaultsClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return kvmgmt.VaultsClient{}, err
	}
	client := kvmgmt.NewVaultsClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return kvmgmt.VaultsClient{}, err
	}
	return client, nil
}

func NewSecretClient() (keyvault.BaseClient, error) {
	a, err := injectKeyvaultAuthorizer()
	if err != nil {
		return keyvault.BaseClient{}, err
	}
	client := keyvault.New()
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return keyvault.BaseClient{}, err
	}
	return client, nil
}

func NewRoleAssignmentClient(subID string) (authorization.RoleAssignmentsClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return authorization.RoleAssignmentsClient{}, err
	}
	client := authorization.NewRoleAssignmentsClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return authorization.RoleAssignmentsClient{}, err
	}
	return client, nil
}

func NewRoleDefinitionClient(subID string) (authorization.RoleDefinitionsClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return authorization.RoleDefinitionsClient{}, err
	}
	client := authorization.NewRoleDefinitionsClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return authorization.RoleDefinitionsClient{}, err
	}
	return client, nil
}

func NewScaleSetClient(subID string) (compute.VirtualMachineScaleSetsClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return compute.VirtualMachineScaleSetsClient{}, err
	}
	client := compute.NewVirtualMachineScaleSetsClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return compute.VirtualMachineScaleSetsClient{}, err
	}
	return client, nil
}

// change to use service principal (tenantID) instead of subscription ID?
func NewScaleSetVMsClient(subID string) (compute.VirtualMachineScaleSetVMsClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return compute.VirtualMachineScaleSetVMsClient{}, err
	}
	client := compute.NewVirtualMachineScaleSetVMsClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return compute.VirtualMachineScaleSetVMsClient{}, err
	}
	return client, nil
}

func NewResourceSkusClient(subID string) (compute.ResourceSkusClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return compute.ResourceSkusClient{}, err
	}
	client := compute.NewResourceSkusClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return compute.ResourceSkusClient{}, err
	}
	return client, nil
}

func NewStorageAccountClient(subID string) (storage.AccountsClient, error) {
	a, err := injectAuthorizer()
	if err != nil {
		return storage.AccountsClient{}, err
	}
	client := storage.NewAccountsClient(subID)
	client.Authorizer = a
	if err := client.AddToUserAgent(userAgent); err != nil {
		return storage.AccountsClient{}, err
	}
	return client, nil
}
