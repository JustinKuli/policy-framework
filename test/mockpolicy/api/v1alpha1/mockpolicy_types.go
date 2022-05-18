/*
Copyright 2022.

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
	framework "github.com/JustinKuli/policy-framework/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MockPolicySpec defines the desired state of MockPolicy
type MockPolicySpec struct {
	framework.PolicyTypeSpec `json:",inline"`

	// Foo is an example field of MockPolicy.
	Foo string `json:"foo,omitempty"`
}

// MockPolicyStatus defines the observed state of MockPolicy
type MockPolicyStatus struct {
	framework.PolicyTypeStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MockPolicy is the Schema for the mockpolicies API
type MockPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MockPolicySpec   `json:"spec,omitempty"`
	Status MockPolicyStatus `json:"status,omitempty"`
}

func (p MockPolicy) PolicySpec() *framework.PolicyTypeSpec {
	return &p.Spec.PolicyTypeSpec
}

func (p MockPolicy) PolicyStatus() *framework.PolicyTypeStatus {
	return &p.Status.PolicyTypeStatus
}

// blank assignment to verify that PolicyType implements PolicyTyper
var _ framework.PolicyTyper = &MockPolicy{}

//+kubebuilder:object:root=true

// MockPolicyList contains a list of MockPolicy
type MockPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MockPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MockPolicy{}, &MockPolicyList{})
}
