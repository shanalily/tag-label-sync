// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package scalesets

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"

	"tag-label-sync.io/azure"
)

type Service interface {
	Get(context.Context, string, string) (compute.VirtualMachineScaleSet, error)
	CreateOrUpdate(context.Context, string, string, compute.VirtualMachineScaleSet) (compute.VirtualMachineScaleSet, error)
	Delete(ctx context.Context, group, name string) error
}

type Client struct {
	group    string
	internal Service
}

func NewClient(subID, group string) (*Client, error) {
	c, err := newClient(subID)
	if err != nil {
		return nil, err
	}

	return &Client{group: group, internal: c}, nil
}

func (c *Client) Get(ctx context.Context, name string) (*Spec, error) {
	id, err := c.internal.Get(ctx, c.group, name)
	if err != nil && azure.IsNotFound(err) {
		return defaultSpec(), nil
	} else if err != nil {
		return nil, err
	}

	return &Spec{&id}, nil
}

func (c *Client) Ensure(ctx context.Context, name string, spec *Spec) error {
	result, err := c.internal.CreateOrUpdate(ctx, c.group, name, *spec.internal)
	if err != nil {
		return err
	}
	spec.internal = &result
	return nil
}

func (c *Client) Delete(ctx context.Context, name string) error {
	err := c.internal.Delete(ctx, c.group, name)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	return err
}
