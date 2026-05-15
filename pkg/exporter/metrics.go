/*
Copyright 2022 Richard Kosegi

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

package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "tuya"
	subsystem = "smartplug"
)

func (e *exporter) newDeviceMetrics() DeviceMetrics {
	devLabels := []string{"device"}
	if e.cfg.ExtraDeviceLabels != nil {
		for _, ln := range *e.cfg.ExtraDeviceLabels {
			devLabels = append(devLabels, ln)
		}
	}
	return DeviceMetrics{
		ScrapeDuration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrape_duration",
			Help:      "Summary of scrape operation",
		}, devLabels),
		ScrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred while scraping",
		}, devLabels),
		Current: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "current",
			Help:      "Electrical current drawn, in Amperes",
		}, devLabels),
		Voltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "voltage",
			Help:      "Electrical voltage, in Volts",
		}, devLabels),
		Power: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "power",
			Help:      "Total power used, in Watts",
		}, devLabels),
		SwitchOn: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "switch_on",
			Help:      "Whether the plug is switched on (1 for on, 0 for off).",
		}, devLabels),
		SentPackets: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "sent_packets_total",
			Help:      "Total number of sent packets",
		}, devLabels),
		ReadPackets: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "read_packets_total",
			Help:      "Total number of read packets",
		}, devLabels),
		SentErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "sent_errors_total",
			Help:      "Total number of sent errors",
		}, devLabels),
		ReadErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "read_errors_total",
			Help:      "Total number of read errors",
		}, devLabels),
	}
}

func newCommonMetrics() GlobalMetrics {
	return GlobalMetrics{
		// since devices are scraped in parallel, this metric captures overall duration
		TotalScrapes: prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrapes_total",
			Help:      "Summary of scrape operation over all devices",
		}),
		Error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics resulted in an error (1 for error, 0 for success).",
		}),
	}
}
