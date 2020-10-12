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

package kafka

import (
	"context"
	"time"

	"github.com/Shopify/sarama"

	"knative.dev/eventing-kafka/pkg/common/consumer"
)

type rateLimiter struct {
	handler func(context.Context, *sarama.ConsumerMessage) (bool, error)
	guard   chan bool
}

func NewRateLimiter(rps int, handler func(context.Context, *sarama.ConsumerMessage) (bool, error)) consumer.KafkaConsumerHandler {
	rl := rateLimiter{
		handler: handler,
		guard:   make(chan bool),
	}

	interval := time.Duration(1.0 / float64(rps) * float64(time.Second))
	go func() {
		for {
			time.Sleep(interval)
			rl.guard <- true
		}
	}()

	return &rl
}

func (rl *rateLimiter) Handle(ctx context.Context, msg *sarama.ConsumerMessage) (bool, error) {
	<-rl.guard

	return rl.handler(ctx, msg)
}
