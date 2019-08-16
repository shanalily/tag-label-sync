// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT license.

package scalesetvms

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
)

type SpecOption func(*Spec) *Spec

type Spec struct {
	internal []compute.VirtualMachineScaleSetVM
}

func Internal(spec *Spec) []compute.VirtualMachineScaleSetVM {
	return spec.internal
}

func defaultSpec() *Spec {
	return &Spec{internal: nil}
}

// Instances returns a map of computername -> instance list
func (spec *Spec) Instances() map[string]string {
	instances := make(map[string]string)
	for _, vm := range spec.internal {
		instances[strings.ToLower(*vm.VirtualMachineScaleSetVMProperties.OsProfile.ComputerName)] = *vm.InstanceID
	}
	return instances
}
