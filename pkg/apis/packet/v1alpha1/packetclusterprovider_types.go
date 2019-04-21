/*
Copyright 2019 The Kubernetes Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PacketClusterProviderSpec defines the desired state of PacketClusterProvider
type PacketClusterProviderSpec struct {
	Project      string `json:"project"`
	Facility     string `json:"facility"`
	Plan         string `json:"plan"`
	BillingCycle string `json:"billingCycle,omitempty"`
	SecretRef    string `json:"secretRef"`
}

// PacketClusterProviderStatus defines the observed state of PacketClusterProvider
type PacketClusterProviderStatus struct {
	ProjectID string `json:"projectID,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PacketClusterProvider is the Schema for the packetclusterproviders API
// +k8s:openapi-gen=true
type PacketClusterProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PacketClusterProviderSpec   `json:"spec,omitempty"`
	Status PacketClusterProviderStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PacketClusterProviderList contains a list of PacketClusterProvider
type PacketClusterProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PacketClusterProvider `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PacketClusterProvider{}, &PacketClusterProviderList{})
}
