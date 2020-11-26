/*
Copyright 2020 The Knative Authors

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

package scheduler

import (
	duckv1alpha1 "knative.dev/eventing-kafka/pkg/apis/duck/v1alpha1"
)

func GetTotalReplicas(placements []duckv1alpha1.Placement) int32 {
	r := int32(0)
	for _, p := range placements {
		r += p.Replicas
	}
	return r
}

func GetPlacementForPod(placements []duckv1alpha1.Placement, podName string) *duckv1alpha1.Placement {
	for i := 0; i < len(placements); i++ {
		if placements[i].PodName == podName {
			return &placements[i]
		}
	}
	return nil
}

func CopyPlacements(placements []duckv1alpha1.Placement) []duckv1alpha1.Placement {
	result := make([]duckv1alpha1.Placement, len(placements))
	for i, p := range placements {
		result[i] = p
	}
	return result
}
