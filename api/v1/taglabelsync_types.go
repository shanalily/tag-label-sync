/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TagLabelSyncSpec defines the desired state of TagLabelSync
type TagLabelSyncSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// should I assume only one cluster in one resource group? do I specify the cluster?
	// isn't this applied to only one cluster?

	// identify the Azure VMs
	Identity        AzureManagedIdentity `json:"identity,omitempty"`        // should this be a struct?
	ResourceGroupID string               `json:"resourceGroupID,omitempty"` // only looking at one resource group

	// identify the cluster nodes
}

// TagLabelSyncStatus defines the observed state of TagLabelSync
type TagLabelSyncStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// store node labels here?
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TagLabelSync is the Schema for the taglabelsyncs API
type TagLabelSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TagLabelSyncSpec   `json:"spec,omitempty"`
	Status TagLabelSyncStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TagLabelSyncList contains a list of TagLabelSync
type TagLabelSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TagLabelSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TagLabelSync{}, &TagLabelSyncList{})
}
