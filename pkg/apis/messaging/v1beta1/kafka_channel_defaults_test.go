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

package v1beta1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	duckv1 "knative.dev/eventing/pkg/apis/duck/v1"

	"knative.dev/eventing-kafka/pkg/common/constants"
)

const (
	testNumPartitions     = 10
	testReplicationFactor = 5
)

func TestKafkaChannelDefaults(t *testing.T) {
	testCases := map[string]struct {
		initial  KafkaChannel
		expected KafkaChannel
	}{
		"nil spec": {
			initial: KafkaChannel{},
			expected: KafkaChannel{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"messaging.knative.dev/subscribable": "v1"},
				},
				Spec: KafkaChannelSpec{
					NumPartitions:     constants.DefaultNumPartitions,
					ReplicationFactor: constants.DefaultReplicationFactor,
					ChannelableSpec: duckv1.ChannelableSpec{
						Delivery: &duckv1.DeliverySpec{
							Retry:         pointer.Int32Ptr(10),
							BackoffPolicy: (*duckv1.BackoffPolicyType)(pointer.StringPtr(string(duckv1.BackoffPolicyExponential))),
							BackoffDelay:  pointer.StringPtr("PT0.1S"),
						},
					},
				},
			},
		},
		"numPartitions not set": {
			initial: KafkaChannel{
				Spec: KafkaChannelSpec{
					ReplicationFactor: testReplicationFactor,
				},
			},
			expected: KafkaChannel{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"messaging.knative.dev/subscribable": "v1"},
				},
				Spec: KafkaChannelSpec{
					NumPartitions:     constants.DefaultNumPartitions,
					ReplicationFactor: testReplicationFactor,
					ChannelableSpec: duckv1.ChannelableSpec{
						Delivery: &duckv1.DeliverySpec{
							Retry:         pointer.Int32Ptr(10),
							BackoffPolicy: (*duckv1.BackoffPolicyType)(pointer.StringPtr(string(duckv1.BackoffPolicyExponential))),
							BackoffDelay:  pointer.StringPtr("PT0.1S"),
						},
					},
				},
			},
		},
		"replicationFactor not set": {
			initial: KafkaChannel{
				Spec: KafkaChannelSpec{
					NumPartitions: testNumPartitions,
				},
			},
			expected: KafkaChannel{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"messaging.knative.dev/subscribable": "v1"},
				},
				Spec: KafkaChannelSpec{
					NumPartitions:     testNumPartitions,
					ReplicationFactor: constants.DefaultReplicationFactor,
					ChannelableSpec: duckv1.ChannelableSpec{
						Delivery: &duckv1.DeliverySpec{
							Retry:         pointer.Int32Ptr(10),
							BackoffPolicy: (*duckv1.BackoffPolicyType)(pointer.StringPtr(string(duckv1.BackoffPolicyExponential))),
							BackoffDelay:  pointer.StringPtr("PT0.1S"),
						},
					},
				},
			},
		},
		"retry set": {
			initial: KafkaChannel{
				Spec: KafkaChannelSpec{
					NumPartitions: testNumPartitions,
					ChannelableSpec: duckv1.ChannelableSpec{
						Delivery: &duckv1.DeliverySpec{
							Retry: pointer.Int32Ptr(10),
						},
					},
				},
			},
			expected: KafkaChannel{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"messaging.knative.dev/subscribable": "v1"},
				},
				Spec: KafkaChannelSpec{
					NumPartitions:     testNumPartitions,
					ReplicationFactor: constants.DefaultReplicationFactor,
					ChannelableSpec: duckv1.ChannelableSpec{
						Delivery: &duckv1.DeliverySpec{
							Retry: pointer.Int32Ptr(10),
						},
					},
				},
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {
			tc.initial.SetDefaults(context.TODO())
			if diff := cmp.Diff(tc.expected, tc.initial); diff != "" {
				t.Fatalf("Unexpected defaults (-want, +got): %s", diff)
			}
		})
	}
}
