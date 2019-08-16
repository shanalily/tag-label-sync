package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AzureMeta struct {
	SubscriptionID string `json:"subscriptionID,omitempty"`
	ResourceGroup  string `json:"resourceGroup,omitempty"`
	Location       string `json:"location,omitempty"`
}

// service principal info
type AzureManagedIdentity struct {
	PrincipalID string `json:"principalID,omitempty"`
	TenantID    string `json:"tenantID,omitempty"`
}

type AzureClusterSpec struct {
	metav1.TypeMeta `json:",inline"`
}

type AzureClusterStatus struct {
	metav1.TypeMeta `json:",inline"`

	ResourceGroup string               `json:"resourceGroupID,omitempty"`
	Identity      AzureManagedIdentity `json:"managedIdentity,omitempty"`
}
