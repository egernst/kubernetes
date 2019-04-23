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

package v1beta1_test

import (
	"k8s.io/api/core/v1"
	"k8s.io/api/node/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	nodev1beta1 "k8s.io/kubernetes/pkg/apis/node/v1beta1"
	"testing"

	// enforce that all types are installed
	_ "k8s.io/kubernetes/pkg/api/testapi"
)

func TestSetDefaultsOvehead(t *testing.T) {
	tests := []struct {
		overhead        *v1beta1.Overhead
		expectedRequest string
		expectedLimit   string
		test            string
	}{
		{
			overhead:        &v1beta1.Overhead{},
			expectedRequest: "0",
			expectedLimit:   "0",
			test:            "verify we default zeros if PodFixed is nil",
		},
		{
			overhead: &v1beta1.Overhead{
				PodFixed: &v1.ResourceRequirements{},
			},
			expectedRequest: "0",
			expectedLimit:   "0",
			test:            "verify we default zeros if req and limits are both nil",
		},
		{
			overhead: &v1beta1.Overhead{
				PodFixed: &v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU: resource.MustParse("4"),
					},
				},
			},
			expectedRequest: "4",
			expectedLimit:   "4",
			test:            "verify limits default to requests if only request is specified",
		},
		{
			overhead: &v1beta1.Overhead{
				PodFixed: &v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU: resource.MustParse("5"),
					},
				},
			},
			expectedRequest: "5",
			expectedLimit:   "5",
			test:            "verify requests default to limits if only limit is specified",
		},
		{
			overhead: &v1beta1.Overhead{
				PodFixed: &v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU: resource.MustParse("6"),
					},
					Limits: v1.ResourceList{
						v1.ResourceCPU: resource.MustParse("7"),
					},
				},
			},
			expectedRequest: "6",
			expectedLimit:   "7",
			test:            "verify unique requests and limits are honored",
		},
	}

	for _, test := range tests {
		nodev1beta1.SetDefaults_Overhead(test.overhead)

		if resultingRequest := test.overhead.PodFixed.Requests[v1.ResourceCPU]; resultingRequest.String() != test.expectedRequest {
			t.Errorf("Failure: %s: expected: %s, got: %s", test.test, test.expectedRequest, resultingRequest.String())
		}
		if resultingLimit := test.overhead.PodFixed.Limits[v1.ResourceCPU]; resultingLimit.String() != test.expectedLimit {
			t.Errorf("Failure: %s: expected: %s, got: %s", test.test, test.expectedLimit, resultingLimit.String())
		}

	}
}
