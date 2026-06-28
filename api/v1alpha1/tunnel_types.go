/*
Copyright 2026.

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
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TunnelSpec defines the desired state of Tunnel
type TunnelSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// Domain is the external domain to expose via Cloudflare Tunnel
	// +kubebuilder:validation:Required
	Domain string `json:"domain"`

	// ServiceName is the name of the Kubernetes Service to expose
	// +kubebuilder:validation:Required
	ServiceName string `json:"serviceName"`

	// ServicePort is the port on the Service to expose
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	ServicePort int32 `json:"servicePort"`

	// SecretRef is the name of the Kubernetes Secret containing the Cloudflare API Token.
	// The secret should have a key "api-token".
	// +kubebuilder:validation:Required
	SecretRef string `json:"secretRef"`

	// SecretNamespace is the namespace of the Kubernetes Secret.
	// If not specified, the namespace of the Tunnel object is used.
	// +optional
	SecretNamespace string `json:"secretNamespace,omitempty"`

	// SharedTunnelName is the name of the master Cloudflare Tunnel to attach this route to.
	// +kubebuilder:validation:Required
	SharedTunnelName string `json:"sharedTunnelName"`

	// AccountID is your Cloudflare Account ID.
	// +kubebuilder:validation:Required
	AccountID string `json:"accountID"`
}

// TunnelStatus defines the observed state of Tunnel.
type TunnelStatus struct {
	// Conditions represent the current state of the Tunnel resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedState indicates if the tunnel route is currently active in Cloudflare.
	// +optional
	State string `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=".status.state"

// Tunnel is the Schema for the tunnels API
type Tunnel struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Tunnel
	// +required
	Spec TunnelSpec `json:"spec"`

	// status defines the observed state of Tunnel
	// +optional
	Status TunnelStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// TunnelList contains a list of Tunnel
type TunnelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Tunnel `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &Tunnel{}, &TunnelList{})
		return nil
	})
}
