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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

//go:embed testdata/*
var testfiles embed.FS

var _ = Describe("MockPolicy controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PolicyName      = "test-policy"
		PolicyNamespace = "default"

		// timeout  = time.Second * 10
		// duration = time.Second * 10
		// interval = time.Millisecond * 250
	)

	Context("Creating from YAML files", func() {
		tests := []struct {
			testname string
			filename string
			matcher  types.GomegaMatcher
		}{
			{
				"Shouldn't create an empty spec",
				"empty-spec.yaml",
				HaveOccurred(),
			},
			{
				"Shouldn't create when namespaceSelector is empty",
				"empty-ns-select.yaml",
				HaveOccurred(),
			},
			{
				"Shouldn't create when include is empty",
				"empty-include.yaml",
				HaveOccurred(),
			},
			{
				"Shouldn't create when include has 1 empty item",
				"empty-item-include.yaml",
				HaveOccurred(),
			},
			{
				"Shouldn't create when include has 1 populated and 1 empty item",
				"mix-item-include.yaml",
				HaveOccurred(),
			},
			{
				"Should create when include is valid",
				"valid-single-include.yaml",
				Succeed(),
			},
			{
				"Should create when include is valid with multiple items",
				"valid-multi-include.yaml",
				Succeed(),
			},
			{
				"Should create when include is valid and exclude is empty",
				"empty-exclude.yaml",
				Succeed(),
			},
			{
				"Should create when include and exclude are valid",
				"valid-ns-select.yaml",
				Succeed(),
			},
			{
				"Shouldn't create when severity is a bad value",
				"invalid-severity.yaml",
				HaveOccurred(),
			},
			{
				"Shouldn't create when severity is empty",
				"invalid-empty-severity.yaml",
				HaveOccurred(),
			},
			{
				"Should create when severity is 'High'",
				"valid-severity.yaml",
				Succeed(),
			},
			{
				"Shouldn't create when remediation is empty",
				"invalid-empty-remediation.yaml",
				HaveOccurred(),
			},
			{
				"Shouldn't create when remediation is a bad value",
				"invalid-remediation.yaml",
				HaveOccurred(),
			},
			{
				"Should create when remediation is 'inform'",
				"valid-remediation.yaml",
				Succeed(),
			},
			{
				"Should create when given the sample yaml",
				"valid-sample.yaml",
				Succeed(),
			},
		}

		for _, tc := range tests {
			filename := tc.filename
			matcher := tc.matcher
			It(tc.testname, func() {
				Expect(createFromFile(filename)).Should(matcher)
			})
		}
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
