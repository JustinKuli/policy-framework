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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComplianceConditionType is the condition type that indicates whether the
// policy is compliant. If policies use `status.conditions`, they should report
// their compliance status on this condition type.
const ComplianceConditionType string = "Compliant"

// ReasonViolationsFound should be used when the policy is not compliant due to
// objects found on the cluster that cause violations. If the policy is not
// compliant for a different reason (like an error), then the policy should use
// another reason.
const ReasonViolationsFound string = "ViolationsFound"

// ReasonNoCompliantObjects should be used when the policy requires certain
// objects to be present on the cluster, but they are not found. It should be
// used to contrast with ViolationsFound, where the objects were found on the
// cluster, but they do not match the desired spec/state.
const ReasonNoCompliantObjects string = "NoCompliantObjects"

// ReasonPolicyError should be used when the policy is not compliant due to an
// error that occurred while evaluating the policy. The status of the condition
// should be either False or Unknown.
const ReasonPolicyError string = "PolicyError"

// ReasonPolicyCompliant should be used when the policy was evaluated without
// error, and found to be compliant. The status of the condition should be True.
const ReasonPolicyCompliant string = "PolicyCompliant"

// UpdateCondition sets the Compliance condition in the given status to match
// the ComplianceState, and to have the given reason and message. It will update
// the LastTransitionTime if the status has changed, and it will initialize the
// condition if it did not already exist.
func UpdateCondition(status *PolicyTypeStatus, reason, msg string) {
	compCond := meta.FindStatusCondition(status.Conditions, ComplianceConditionType)
	if compCond == nil {
		compCond = &metav1.Condition{Type: ComplianceConditionType}
	}

	switch status.ComplianceState {
	case Compliant:
		compCond.Status = metav1.ConditionTrue
	case NonCompliant:
		compCond.Status = metav1.ConditionFalse
	default:
		compCond.Status = metav1.ConditionUnknown
	}

	compCond.Reason = reason
	compCond.Message = msg

	meta.SetStatusCondition(&status.Conditions, *compCond)
}
