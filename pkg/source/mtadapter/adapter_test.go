/*
Copyright 2019 The Knative Authors

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

package mtadapter

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pkgtesting "knative.dev/pkg/reconciler/testing"
	"knative.dev/pkg/source"

	"knative.dev/eventing/pkg/adapter/v2"
	adaptertest "knative.dev/eventing/pkg/adapter/v2/test"
	"knative.dev/eventing/pkg/kncloudevents"

	sourcesv1beta1 "knative.dev/eventing-kafka/pkg/apis/sources/v1beta1"
)

var runningAdapterChan = make(chan *dummyAdapter)
var stoppingAdapterChan = make(chan *dummyAdapter)

func TestUpdateRemoveSources(t *testing.T) {
	ctx, _ := pkgtesting.SetupFakeContext(t)
	ctx, cancelAdapter := context.WithCancel(ctx)

	env := NewEnvConfig()
	ceClient := adaptertest.NewTestClient()

	adapter := newAdapter(ctx, env, ceClient, newDummyAdapter).(*Adapter)

	adapterStopped := make(chan bool)
	go func() {
		err := adapter.Start(ctx)
		if err != nil {
			t.Errorf("Unexpected error %v", err)
		}
		adapterStopped <- true
	}()

	adapter.Update(ctx, &sourcesv1beta1.KafkaSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-ns",
		},
		Spec:   sourcesv1beta1.KafkaSourceSpec{},
		Status: sourcesv1beta1.KafkaSourceStatus{},
	})

	if _, ok := adapter.sources["test-ns/test-name"]; !ok {
		t.Error(`Expected adapter to contain "test-ns/test-name"`)
	}

	select {
	case a := <-runningAdapterChan:
		if !a.running {
			t.Error("Expected adapter to be running")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("sub-adapter failed to start after 2 seconds")
	}

	adapter.Remove(ctx, &sourcesv1beta1.KafkaSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test-ns",
		},
		Spec:   sourcesv1beta1.KafkaSourceSpec{},
		Status: sourcesv1beta1.KafkaSourceStatus{},
	})

	if _, ok := adapter.sources["test-ns/test-name"]; ok {
		t.Error(`Expected adapter to not contain "test-ns/test-name"`)
	}

	select {
	case a := <-stoppingAdapterChan:
		if a.running {
			t.Error("Expected adapter to not be running")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("sub-adapter failed to stop after 2 seconds")
	}

	// Make sure the adapter is still running
	select {
	case <-ctx.Done():
		t.Error("Expected adapter to be running")
	default:
	}

	cancelAdapter()

	select {
	case <-adapterStopped:
	case <-time.After(2 * time.Second):
		t.Error("adapter failed to stop after 2 seconds")
	}
}

type dummyAdapter struct {
	running bool
}

func newDummyAdapter(ctx context.Context, env adapter.EnvConfigAccessor, adapter *kncloudevents.HTTPMessageSender, reporter source.StatsReporter) adapter.MessageAdapter {
	return &dummyAdapter{}
}

func (d *dummyAdapter) Start(ctx context.Context) error {
	d.running = true
	runningAdapterChan <- d

	<-ctx.Done()

	d.running = false
	stoppingAdapterChan <- d
	return nil
}
