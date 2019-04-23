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

package v1beta1

import (
	"k8s.io/api/core/v1"
	"k8s.io/api/node/v1beta1"
)

func SetDefaults_Overhead(obj *v1beta1.Overhead) {
	// If no requests or limits are specified, default to zro
	if obj.PodFixed == nil {
		obj.PodFixed = new(v1.ResourceRequirements)
	} else {
		// Unlike container resource requiremnts, it doesn't make sense to allow
		// for an unbound limit if an overhead is being defined. If requests are
		// provided but not limits, limits should default to value of requests
		if obj.PodFixed.Limits == nil {
			obj.PodFixed.Limits = make(v1.ResourceList)
		}

		if obj.PodFixed.Requests == nil {
			obj.PodFixed.Requests = make(v1.ResourceList)
		}

		for key, value := range obj.PodFixed.Limits {
			if _, exists := obj.PodFixed.Requests[key]; !exists {
				obj.PodFixed.Requests[key] = *(value.Copy())
			}
		}
		for key, value := range obj.PodFixed.Requests {
			if _, exists := obj.PodFixed.Limits[key]; !exists {
				obj.PodFixed.Limits[key] = *(value.Copy())
			}
		}
	}
}
