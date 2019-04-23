/*
Copyright 2018 The Kubernetes Authors.

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

package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kubernetes/pkg/apis/core"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/node"

	"github.com/stretchr/testify/assert"
)

func TestValidateRuntimeClass(t *testing.T) {
	tests := []struct {
		name        string
		rc          node.RuntimeClass
		expectError bool
	}{{
		name:        "invalid name",
		expectError: true,
		rc: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "&!@#"},
			Handler:    "foo",
		},
	}, {
		name:        "invalid Handler name",
		expectError: true,
		rc: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Handler:    "&@#$",
		},
	}, {
		name:        "invalid empty RuntimeClass",
		expectError: true,
		rc: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "empty"},
		},
	}, {
		name:        "valid Handler",
		expectError: false,
		rc: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Handler:    "bar-baz",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := ValidateRuntimeClass(&test.rc)
			if test.expectError {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidateRuntimeUpdate(t *testing.T) {
	old := node.RuntimeClass{
		ObjectMeta: metav1.ObjectMeta{Name: "foo"},
		Handler:    "bar",
	}
	tests := []struct {
		name        string
		expectError bool
		old, new    node.RuntimeClass
	}{{
		name: "valid metadata update",
		old:  old,
		new: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "foo",
				Labels: map[string]string{"foo": "bar"},
			},
			Handler: "bar",
		},
	}, {
		name:        "invalid metadata update",
		expectError: true,
		old:         old,
		new: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "empty",
				ClusterName: "somethingelse", // immutable
			},
			Handler: "bar",
		},
	}, {
		name:        "invalid Handler update",
		expectError: true,
		old:         old,
		new: node.RuntimeClass{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Handler:    "somethingelse",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// So we don't need to write it in every test case...
			test.old.ObjectMeta.ResourceVersion = "1"
			test.new.ObjectMeta.ResourceVersion = "1"

			errs := ValidateRuntimeClassUpdate(&test.new, &test.old)
			if test.expectError {
				assert.NotEmpty(t, errs)
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidateResourceRequirements(t *testing.T) {
	successCase := []struct {
		Name     string
		overhead *node.Overhead
	}{
		{
			Name: "Overhead with Requests equal to Limits",
			overhead: &node.Overhead{
				PodFixed: &core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceName(core.ResourceCPU):    resource.MustParse("10"),
						core.ResourceName(core.ResourceMemory): resource.MustParse("10G"),
					},
					Limits: core.ResourceList{
						core.ResourceName(core.ResourceCPU):    resource.MustParse("10"),
						core.ResourceName(core.ResourceMemory): resource.MustParse("10G"),
					},
				},
			},
		},
		{
			Name: "Overhead with only Limits",
			overhead: &node.Overhead{
				PodFixed: &core.ResourceRequirements{
					Limits: core.ResourceList{
						core.ResourceName(core.ResourceCPU):    resource.MustParse("10"),
						core.ResourceName(core.ResourceMemory): resource.MustParse("10G"),
					},
				},
			},
		},
		{
			Name: "Overhead with only Requests",
			overhead: &node.Overhead{
				PodFixed: &core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceName(core.ResourceCPU):    resource.MustParse("10"),
						core.ResourceName(core.ResourceMemory): resource.MustParse("10G"),
					},
				},
			},
		},
		{
			Name: "Overhead with Requests Less Than Limits",
			overhead: &node.Overhead{
				PodFixed: &core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceName(core.ResourceCPU):    resource.MustParse("9"),
						core.ResourceName(core.ResourceMemory): resource.MustParse("9G"),
					},
					Limits: core.ResourceList{
						core.ResourceName(core.ResourceCPU):    resource.MustParse("10"),
						core.ResourceName(core.ResourceMemory): resource.MustParse("10G"),
					},
				},
			},
		},
	}
	for _, tc := range successCase {
		if errs := ValidateOverhead(tc.overhead, field.NewPath("overheads")); len(errs) != 0 {
			t.Errorf("%q unexpected error: %v", tc.Name, errs)
		}
	}

	errorCase := []struct {
		Name     string
		overhead *node.Overhead
	}{
		{
			Name: "Overhead with Requests Larger Than Limits",
			overhead: &node.Overhead{
				PodFixed: &core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceName(core.ResourceCPU):    resource.MustParse("10"),
						core.ResourceName(core.ResourceMemory): resource.MustParse("10G"),
					},
					Limits: core.ResourceList{
						core.ResourceName(core.ResourceCPU):    resource.MustParse("9"),
						core.ResourceName(core.ResourceMemory): resource.MustParse("9G"),
					},
				},
			},
		},
		{
			Name: "Invalid Resources with Requests",
			overhead: &node.Overhead{
				PodFixed: &core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceName("my.org"): resource.MustParse("10m"),
					},
				},
			},
		},
		{
			Name: "Invalid Resources with Limits",
			overhead: &node.Overhead{
				PodFixed: &core.ResourceRequirements{
					Limits: core.ResourceList{
						core.ResourceName("my.org"): resource.MustParse("9m"),
					},
				},
			},
		},
	}
	for _, tc := range errorCase {
		if errs := ValidateOverhead(tc.overhead, field.NewPath("resources")); len(errs) == 0 {
			t.Errorf("%q expected error", tc.Name)
		}
	}
}
