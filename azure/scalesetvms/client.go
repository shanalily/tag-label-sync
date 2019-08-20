// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package scalesetvms

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"tag-label-sync.io/azure"
)

type Service interface {
	List(context.Context, string, string) ([]compute.VirtualMachineScaleSetVM, error)
}

type Client struct {
	group    string
	internal Service
}

func NewClientService(group string, internal Service) *Client {
	return &Client{group: group, internal: internal}
}

func NewClient(subID, group string) (*Client, error) {
	c, err := newClient(subID)
	if err != nil {
		return nil, err
	}

	return &Client{group: group, internal: c}, nil
}

func (c *Client) List(ctx context.Context, name string) (*Spec, error) {
	vmlist, err := c.internal.List(ctx, c.group, name)
	if err != nil && azure.IsNotFound(err) {
		return defaultSpec(), nil
	} else if err != nil {
		return nil, err
	}

	return &Spec{internal: vmlist}, nil
}
