/*
Copyright 2025 Rafal Jan.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.-
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IngressTemplate defines the template for creating an Ingress resource
type IngressTemplate struct {
	// Standard object's metadata.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XPreserveUnknownFields
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the Ingress
	// +kubebuilder:validation:Required
	Spec networkingv1.IngressSpec `json:"spec"`
}

// AppIngressSpec defines the desired state of AppIngress.
type AppIngressSpec struct {
	// Template defines the Ingress to be created
	// +kubebuilder:validation:Required
	Template IngressTemplate `json:"template"`

	// TargetNamespace is the namespace where the Ingress will be created
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	TargetNamespace string `json:"targetNamespace"`
}

// AppIngressStatus defines the observed state of AppIngress.
type AppIngressStatus struct {
	// Conditions represent the latest available observations of the AppIngress's current state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Target Namespace",type="string",JSONPath=".spec.targetNamespace"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// AppIngress is the Schema for the appingresses API.
type AppIngress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppIngressSpec   `json:"spec,omitempty"`
	Status AppIngressStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AppIngressList contains a list of AppIngress.
type AppIngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppIngress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppIngress{}, &AppIngressList{})
}
