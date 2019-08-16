// copied from genesys/pkg/cluster/internal/services/scalesetvms
// so this is actually VMSS VMs, not a scale set itself

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
)

type VMSSSpecOption func(*VMSSSpec) *VMSSSpec

type VMSSSpec struct {
	internal *compute.VirtualMachineScaleSet
}

type VMSSService interface {
	Get(context.Context, string, string) (compute.VirtualMachineScaleSet, error)
	Delete(context.Context, string, string) error
}

type VMSSClient struct {
	group    string
	internal VMSSService
}

type vmssClient struct {
	compute.VirtualMachineScaleSetsClient
}

func newVMSSClient(subID string) (*vmssClient, error) {
	c, err := NewScaleSetClient(subID)
	if err != nil {
		return nil, err
	}
	return &vmssClient{c}, nil
}

func defaultVMSSSpec() *VMSSSpec {
	return &VMSSSpec{internal: &compute.VirtualMachineScaleSet{
		Sku: &compute.Sku{},
		Identity: &compute.VirtualMachineScaleSetIdentity{
			UserAssignedIdentities: map[string]*compute.VirtualMachineScaleSetIdentityUserAssignedIdentitiesValue{},
		},
		VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
			Overprovision: to.BoolPtr(false),
			UpgradePolicy: &compute.UpgradePolicy{
				Mode: compute.Manual,
			},
			VirtualMachineProfile: &compute.VirtualMachineScaleSetVMProfile{
				Priority: compute.Regular,
				OsProfile: &compute.VirtualMachineScaleSetOSProfile{
					AdminUsername: to.StringPtr("azureuser"),
					LinuxConfiguration: &compute.LinuxConfiguration{
						DisablePasswordAuthentication: to.BoolPtr(true),
						SSH: &compute.SSHConfiguration{
							PublicKeys: &[]compute.SSHPublicKey{},
						},
					},
				},
				DiagnosticsProfile: &compute.DiagnosticsProfile{
					BootDiagnostics: &compute.BootDiagnostics{},
				},
				StorageProfile: &compute.VirtualMachineScaleSetStorageProfile{
					ImageReference: &compute.ImageReference{
						Offer:     to.StringPtr("UbuntuServer"),
						Publisher: to.StringPtr("Canonical"),
						Sku:       to.StringPtr("18.04-LTS"),
						Version:   to.StringPtr("latest"),
					},
					OsDisk: &compute.VirtualMachineScaleSetOSDisk{},
				},
				NetworkProfile: &compute.VirtualMachineScaleSetNetworkProfile{
					NetworkInterfaceConfigurations: &[]compute.VirtualMachineScaleSetNetworkConfiguration{},
				},
			},
		},
	}}
}

func NewVMSSClient(subID, group string) (*VMSSClient, error) {
	c, err := newVMSSClient(subID)
	if err != nil {
		return nil, err
	}

	return &VMSSClient{group: group, internal: c}, nil
}

func (c *vmssClient) Get(ctx context.Context, group, name string) (compute.VirtualMachineScaleSet, error) {
	return c.VirtualMachineScaleSetsClient.Get(ctx, group, name)
}

func (c *VMSSClient) Get(ctx context.Context, name string) (*VMSSSpec, error) {
	id, err := c.internal.Get(ctx, c.group, name)
	if err != nil && IsNotFound(err) {
		return defaultVMSSSpec(), nil
	} else if err != nil {
		return nil, err
	}

	return &VMSSSpec{&id}, nil
}

func (c *VMSSClient) Delete(ctx context.Context, name string) error {
	err := c.internal.Delete(ctx, c.group, name)
	if err != nil && IsNotFound(err) {
		return nil
	}
	return err
}
