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
	"embed"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	policytypev1alpha1 "github.com/JustinKuli/policy-framework/api/v1alpha1"
	policyv1alpha1 "github.com/JustinKuli/policy-framework/test/mockpolicy/api/v1alpha1"
)

//go:embed testdata/*
var testfiles embed.FS

var _ = Describe("MockPolicy controller", func() {
	Context("Creating from YAML files", Ordered, func() {
		BeforeAll(deleteDefaultPolicies)
		AfterAll(deleteDefaultPolicies)

		tests := map[string]string{
			"Shouldn't create an empty spec":                                 "empty-spec.yaml",
			"Shouldn't create when namespaceSelector is empty":               "empty-ns-select.yaml",
			"Shouldn't create when include is empty":                         "empty-include.yaml",
			"Shouldn't create when include has 1 empty item":                 "empty-item-include.yaml",
			"Shouldn't create when include has 1 populated and 1 empty item": "mix-item-include.yaml",
			"Should create when include is valid":                            "valid-single-include.yaml",
			"Should create when include is valid with multiple items":        "valid-multi-include.yaml",
			"Should create when include is valid and exclude is empty":       "empty-exclude.yaml",
			"Should create when include and exclude are valid":               "valid-ns-select.yaml",
			"Shouldn't create when severity is a bad value":                  "invalid-severity.yaml",
			"Shouldn't create when severity is empty":                        "invalid-empty-severity.yaml",
			"Should create when severity is 'High'":                          "valid-severity.yaml",
			"Shouldn't create when remediation is empty":                     "invalid-empty-remediation.yaml",
			"Shouldn't create when remediation is a bad value":               "invalid-remediation.yaml",
			"Should create when remediation is 'inform'":                     "valid-remediation.yaml",
			"Should create when given the sample yaml":                       "valid-sample.yaml",
		}

		for testname, testfile := range tests {
			testfile := testfile
			matcher := Succeed()
			if strings.HasPrefix(testname, "Shouldn't") {
				matcher = HaveOccurred()
			}
			It(testname, func() {
				Expect(createFromFile(testfile)).Should(matcher)
			})
		}
	})

	Context("NamespaceSelector testing", Ordered, func() {
		BeforeAll(deleteDefaultPolicies)
		BeforeAll(func() {
			ns := &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: "foo"}}
			Expect(k8sClient.Create(context.TODO(), ns)).Should(Succeed())
			ns = &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: "bar"}}
			Expect(k8sClient.Create(context.TODO(), ns)).Should(Succeed())
			ns = &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: "kube-test"}}
			Expect(k8sClient.Create(context.TODO(), ns)).Should(Succeed())
			ns = &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: "openshift"}}
			Expect(k8sClient.Create(context.TODO(), ns)).Should(Succeed())
		})

		AfterAll(deleteDefaultPolicies)
		AfterAll(func() {
			ns := &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: "foo"}}
			Expect(k8sClient.Delete(context.TODO(), ns)).Should(Succeed())
			ns = &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: "bar"}}
			Expect(k8sClient.Delete(context.TODO(), ns)).Should(Succeed())
			ns = &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: "kube-test"}}
			Expect(k8sClient.Delete(context.TODO(), ns)).Should(Succeed())
			ns = &corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: "openshift"}}
			Expect(k8sClient.Delete(context.TODO(), ns)).Should(Succeed())
		})

		It("Verifies the namespaceSelector works", func() {
			By("Creating the nstest policy")
			Expect(createFromFile("nstest-example.yaml")).Should(Succeed())

			By("Verifying the known namespaces are matched")
			policy := &policyv1alpha1.MockPolicy{}
			Eventually(func() string {
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: "default",
					Name:      "nstest-example",
				}, policy)
				if err != nil {
					return err.Error()
				}
				return policy.Status.Debug
			}, time.Second*10, 1).Should(MatchRegexp(".*foo.*"))
			nstest := policy.Status.Debug
			Expect(nstest).Should(MatchRegexp(".*bar.*"))

			By("Verifying the known namespaces are not matched")
			Expect(nstest).ShouldNot(MatchRegexp(".*kube-test.*"))
			Expect(nstest).ShouldNot(MatchRegexp(".*openshift.*"))
		})
	})

	Context("Compliance Events", Ordered, func() {
		var ownerRefs []v1.OwnerReference
		BeforeAll(func() {
			deleteDefaultPolicies()

			owningpolicy := defaultMockPolicy("owning-policy")
			Expect(k8sClient.Create(context.TODO(), &owningpolicy)).Should(Succeed())

			owningpolicy = policyv1alpha1.MockPolicy{}
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
				Namespace: "default",
				Name:      "owning-policy",
			}, &owningpolicy)).Should(Succeed())

			ownerRefs = []v1.OwnerReference{{
				APIVersion: "policy.open-cluster-management.io/v1alpha1",
				Kind:       "MockPolicy",
				Name:       owningpolicy.Name,
				UID:        owningpolicy.UID,
			}}
		})

		AfterAll(deleteDefaultPolicies)

		It("Verifies that compliance events are emitted", func() {
			By("Creating the owned policy")
			owned := defaultMockPolicy("compliant-owned-policy")
			owned.OwnerReferences = ownerRefs
			owned.Spec.Foo = "compliant"
			Expect(k8sClient.Create(context.TODO(), &owned)).Should(Succeed())

			By("Verifying the policy becomes compliant")
			Eventually(func() string {
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: "default",
					Name:      "compliant-owned-policy",
				}, &owned)
				if err != nil {
					return err.Error()
				}
				return string(owned.Status.ComplianceState)
			}, time.Second*10, 1).Should(Equal("Compliant"))

			By("Verifying an event was emitted")
			Eventually(func() interface{} {
				eventList := &corev1.EventList{}
				err := k8sClient.List(context.TODO(), eventList, client.InNamespace("default"))
				if err != nil {
					return err
				}

				for _, event := range eventList.Items {
					if event.InvolvedObject.Kind == "MockPolicy" &&
						event.InvolvedObject.APIVersion == "policy.open-cluster-management.io/v1alpha1" &&
						event.InvolvedObject.Name == "owning-policy" &&
						event.Reason == "policy: default/compliant-owned-policy" &&
						event.Type == "Normal" {
						return true
					}
				}
				return false
			}, time.Second*10, 1).Should(BeTrue())
		})

		It("Verifies that non-compliance events are emitted", func() {
			By("Creating the owned policy")
			owned := defaultMockPolicy("noncompliant-owned-policy")
			owned.OwnerReferences = ownerRefs
			owned.Spec.Foo = "noncompliant"
			Expect(k8sClient.Create(context.TODO(), &owned)).Should(Succeed())

			By("Verifying the policy becomes noncompliant")
			Eventually(func() string {
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: "default",
					Name:      "noncompliant-owned-policy",
				}, &owned)
				if err != nil {
					return err.Error()
				}
				return string(owned.Status.ComplianceState)
			}, time.Second*10, 1).Should(Equal("NonCompliant"))

			By("Verifying an event was emitted")
			Eventually(func() interface{} {
				eventList := &corev1.EventList{}
				err := k8sClient.List(context.TODO(), eventList, client.InNamespace("default"))
				if err != nil {
					return err
				}

				for _, event := range eventList.Items {
					if event.InvolvedObject.Kind == "MockPolicy" &&
						event.InvolvedObject.APIVersion == "policy.open-cluster-management.io/v1alpha1" &&
						event.InvolvedObject.Name == "owning-policy" &&
						event.Reason == "policy: default/noncompliant-owned-policy" &&
						event.Type == "Warning" {
						return true
					}
				}
				return false
			}, time.Second*10, 1).Should(BeTrue())
		})
	})
})

func createFromFile(name string) error {
	objYAML, err := testfiles.ReadFile("testdata/" + name)
	if err != nil {
		return err
	}

	m := make(map[string]interface{})
	if err := yaml.UnmarshalStrict(objYAML, &m); err != nil {
		return err
	}

	return k8sClient.Create(context.TODO(), &unstructured.Unstructured{Object: m})
}

func deleteDefaultPolicies() {
	policy := &policyv1alpha1.MockPolicy{}
	Expect(k8sClient.DeleteAllOf(context.TODO(), policy, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{Namespace: "default"},
	})).Should(Succeed())
}

func defaultMockPolicy(name string) policyv1alpha1.MockPolicy {
	return policyv1alpha1.MockPolicy{
		ObjectMeta: v1.ObjectMeta{Name: name, Namespace: "default"},
		Spec: policyv1alpha1.MockPolicySpec{PolicyTypeSpec: policytypev1alpha1.PolicyTypeSpec{
			NamespaceSelector: policytypev1alpha1.NamespaceSelector{
				Include: []policytypev1alpha1.NonEmptyString{"*"},
			},
		}},
	}
}
