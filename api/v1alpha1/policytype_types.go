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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:validation:MinLength=1
type NonEmptyString string

// ComplianceState shows the state of enforcement
//+kubebuilder:validation:Enum=Compliant;NonCompliant;UnknownCompliancy
type ComplianceState string

const (
	Compliant         ComplianceState = "Compliant"
	NonCompliant      ComplianceState = "NonCompliant"
	UnknownCompliancy ComplianceState = "UnknownCompliancy"
)

// PolicyTypeSpec includes all fields that should be implemented in the spec of
// all policy types in the policy framework.
type PolicyTypeSpec struct {
	// Severity is how serious the situation is when the policy is not
	// compliant. Accepted values include: low, medium, high, and critical.
	//+kubebuilder:validation:Enum=low;Low;medium;Medium;high;High;critical;Critical
	Severity string `json:"severity,omitempty"`

	// RemediationAction indicates what the policy controller should do when the
	// policy is not compliant. Accepted values include inform, and enforce.
	// Note that not all policy controllers will attempt to automatically
	// remediate a policy, even when set to "enforce".
	//+kubebuilder:validation:Enum=Inform;inform;Enforce;enforce
	RemediationAction string `json:"remediationAction,omitempty"`

	// NamepaceSelector indicates which namespaces on the cluster this policy
	// should apply to, when the policy applies to namespaced objects.
	NamespaceSelector NamespaceSelector `json:"namespaceSelector,omitempty"`

	// LabelSelector is a map of labels and values for the resources that the
	// policy should apply to. Not all policy controllers use this field, but
	// if they do, the resources must match all labels specified here.
	LabelSelector map[string]NonEmptyString `json:"labelSelector,omitempty"`
}

type NamespaceSelector struct {
	// Include is a list of namespaces the policy should apply to. UNIX style
	// wildcards will be expanded, for example "kube-*" will include both
	// "kube-system" and "kube-public".
	//+kubebuilder:validation:Required
	Include []NonEmptyString `json:"include,omitempty"`

	// Exclude is a list of namespaces the policy should _not_ apply to. UNIX
	// style wildcards will be expanded, for example "kube-*" will exclude both
	// "kube-system" and "kube-public".
	Exclude []NonEmptyString `json:"exclude,omitempty"`
}

// PolicyTypeStatus includes fields that are useful for policy types in the
// policy framework to implement in order to report status.
type PolicyTypeStatus struct {
	// ComplianceState indicates whether the policy is compliant or not.
	// Accepted values include: Compliant, NonCompliant, and UnknownCompliancy
	ComplianceState ComplianceState `json:"compliant,omitempty"`

	// CompliancyDetails is implemented differently in each controller, in ways
	// that are not compatible: configuration-policy uses a list, but iam-policy
	// uses a map. So we can not include it in this definition.

	// RelatedObjects are objects on the cluster that were examined in order to
	// determine compliance. Often these are objects that cause a violation, but
	// not always.
	RelatedObjects []RelatedObject `json:"relatedObjects,omitempty"`

	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type RelatedObject struct {
	Object          ObjectRef       `json:"object,omitempty"`
	ComplianceState ComplianceState `json:"compliant,omitempty"`
	Reason          NonEmptyString  `json:"reason,omitempty"`
}

type ObjectRef struct {
	metav1.TypeMeta `json:",inline"`
	Metadata        ObjectMetadata `json:"metadata,omitempty"`
}

// ObjectMetadata contains the resource metadata for an object being processed by the policy
type ObjectMetadata struct {
	// Name of the referent. More info:
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name,omitempty"`

	// Namespace of the referent. More info:
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
	Namespace string `json:"namespace,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PolicyType is the Schema for the policytypes API
type PolicyType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicyTypeSpec   `json:"spec,omitempty"`
	Status PolicyTypeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:generate=false
type PolicyTyper interface {
	client.Object
	PolicySpec() *PolicyTypeSpec
	PolicyStatus() *PolicyTypeStatus
}

func (p PolicyType) PolicySpec() *PolicyTypeSpec {
	return &p.Spec
}

func (p PolicyType) PolicyStatus() *PolicyTypeStatus {
	return &p.Status
}

// blank assignment to verify that PolicyType implements PolicyTyper
var _ PolicyTyper = &PolicyType{}

//+kubebuilder:object:root=true

// PolicyTypeList contains a list of PolicyType
type PolicyTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyType `json:"items"`
}
