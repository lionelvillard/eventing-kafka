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

package mtsource

import (
	"context"
	"k8s.io/client-go/tools/cache"

	"knative.dev/eventing/pkg/apis/sources/v1alpha1"
	"knative.dev/eventing/pkg/reconciler/source"

	kubeclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"
	"knative.dev/pkg/system"

	sourcesv1beta1 "knative.dev/eventing-kafka/pkg/apis/sources/v1beta1"
	kafkaclient "knative.dev/eventing-kafka/pkg/client/injection/client"
	kafkainformer "knative.dev/eventing-kafka/pkg/client/injection/informers/sources/v1beta1/kafkasource"
	"knative.dev/eventing-kafka/pkg/client/injection/reconciler/sources/v1beta1/kafkasource"
	scheduler "knative.dev/eventing-kafka/pkg/common/scheduler/statefulset"
)

func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	kafkaInformer := kafkainformer.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)

	c := &Reconciler{
		KubeClientSet:    kubeclient.Get(ctx),
		kafkaClientSet:   kafkaclient.Get(ctx),
		kafkaLister:      kafkaInformer.Lister(),
		deploymentLister: deploymentInformer.Lister(),
		loggingContext:   ctx,
		configs:          source.WatchConfigurations(ctx, component, cmw),
	}

	impl := kafkasource.NewImpl(ctx, c)

	c.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	// Use a different set of conditions
	sourcesv1beta1.RegisterAlternateKafkaConditionSet(sourcesv1beta1.KafkaMTSourceCondSet)

	c.scheduler = scheduler.NewStatefulSetScheduler(ctx, system.Namespace(), mtadapterName, c.vpodLister, impl.EnqueueKey)

	logging.FromContext(ctx).Info("Setting up kafka event handlers")

	kafkaInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("KafkaSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
