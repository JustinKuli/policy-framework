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

package controllers

import (
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/JustinKuli/policy-framework/api/v1alpha1"
	policyv1alpha1 "github.com/JustinKuli/policy-framework/test/mockpolicy/api/v1alpha1"
)

// MockPolicyReconciler reconciles a MockPolicy object
type MockPolicyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=policy.open-cluster-management.io,resources=mockpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=policy.open-cluster-management.io,resources=mockpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=policy.open-cluster-management.io,resources=mockpolicies/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *MockPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	policy := &policyv1alpha1.MockPolicy{}
	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, probably deleted
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get MockPolicy")
		return ctrl.Result{}, err
	}

	switch policy.Spec.Foo {
	case "nstest":
		selectedNamespaces, err := policy.Spec.NamespaceSelector.GetNamespaces(ctx, r.Client)
		if err != nil {
			log.Error(err, "Failed to GetNamespaces using NamespaceSelector",
				"selector", policy.Spec.NamespaceSelector)
			return ctrl.Result{}, err
		}
		policy.Status.Debug = strings.Join(selectedNamespaces, ",")
	case "compliant":
		policy.Status.ComplianceState = v1alpha1.Compliant
	case "noncompliant":
		policy.Status.ComplianceState = v1alpha1.NonCompliant
	}

	err := r.Status().Update(ctx, policy)
	if err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	v1alpha1.RecordComplianceEvent(r.Recorder, policy, "because test")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MockPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&policyv1alpha1.MockPolicy{}).
		Complete(r)
}
