/*
Copyright 2022 The Kubernetes Authors.

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

package slis

import (
	"context"
	"errors"

	k8smetrics "k8s.io/component-base/metrics"
)

type HealthcheckStatus string

const (
	Success HealthcheckStatus = "success"
	Error   HealthcheckStatus = "error"
	Pending HealthcheckStatus = "pending"
)

type HealthcheckType string

const (
	Livez   HealthcheckType = "livez"
	Readyz  HealthcheckType = "readyz"
	Healthz HealthcheckType = "healthz"
)

var (
	// healthcheck is a Prometheus Gauge metrics used for recording the results of a k8s healthcheck.
	healthcheck = k8smetrics.NewGaugeVec(
		&k8smetrics.GaugeOpts{
			Namespace:      "kubernetes",
			Name:           "healthcheck",
			Help:           "This metric records the result of a single healthcheck.",
			StabilityLevel: k8smetrics.ALPHA,
		},
		[]string{"name", "type", "status"},
	)

	// healthchecksTotal is a Prometheus Counter metrics used for counting the results of a k8s healthcheck.
	healthchecksTotal = k8smetrics.NewCounterVec(
		&k8smetrics.CounterOpts{
			Namespace:      "kubernetes",
			Name:           "healthchecks_total",
			Help:           "This metric records the results of all healthcheck.",
			StabilityLevel: k8smetrics.ALPHA,
		},
		[]string{"name", "type", "status"},
	)
	statuses  = []HealthcheckStatus{Success, Error, Pending}
	statusSet = map[HealthcheckStatus]struct{}{Success: {}, Error: {}, Pending: {}}
)

func Register(registry k8smetrics.KubeRegistry) {
	registry.Register(healthcheck)
	registry.Register(healthchecksTotal)
}

func ResetHealthMetrics() {
	healthcheck.Reset()
	healthchecksTotal.Reset()
}

func ObserveHealthcheck(ctx context.Context, name string, healthcheckType string, status HealthcheckStatus) error {
	if _, ok := statusSet[status]; !ok {
		return errors.New("not a valid healthcheck status")
	}
	for _, s := range statuses {
		if status != s {
			healthcheck.WithContext(ctx).WithLabelValues(name, healthcheckType, string(s)).Set(0)
		}
	}
	healthchecksTotal.WithContext(ctx).WithLabelValues(name, healthcheckType, string(status)).Inc()
	healthcheck.WithContext(ctx).WithLabelValues(name, healthcheckType, string(status)).Set(1)
	return nil
}
