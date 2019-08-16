// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package scalesetvms

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"

	"tag-label-sync.io/azure"
)

type client struct {
	compute.VirtualMachineScaleSetVMsClient
}

func newClient(subID string) (*client, error) {
	c, err := NewScaleSetVMsClient(subID)
	if err != nil {
		return nil, err
	}
	return &client{c}, nil
}

func (c *client) List(ctx context.Context, group, name string) ([]compute.VirtualMachineScaleSetVM, error) {
	result, err := c.VirtualMachineScaleSetVMsClient.List(ctx, group, name, "", "", string(compute.InstanceView))
	if err != nil {
		return nil, err
	}
	vms := make([]compute.VirtualMachineScaleSetVM, 0, len(result.Values()))
	vms = append(vms, result.Values()...)
	return vms, nil
}

func (c *client) Delete(ctx context.Context, group, name, instanceID string) error {
	future, err := c.VirtualMachineScaleSetVMsClient.Delete(ctx, group, name, instanceID)
	if err != nil {
		return err
	}

	err = future.WaitForCompletionRef(ctx, c.VirtualMachineScaleSetVMsClient.Client)
	if err != nil {
		return err
	}

	_, err = future.Result(c.VirtualMachineScaleSetVMsClient)
	return err
}
