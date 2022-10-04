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
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rkosegi/tuya-smartplug-exporter/pkg/proto"
	"github.com/rkosegi/tuya-smartplug-exporter/pkg/types"
	"time"
)

type exporter struct {
	m    types.Metrics
	devs *[]types.Device
	l    log.Logger
}

func (e *exporter) Describe(descs chan<- *prometheus.Desc) {
	e.m.Error.Describe(descs)
	e.m.ScrapeDuration.Describe(descs)
	e.m.TotalScrapes.Describe(descs)
	e.m.ScrapeErrors.Describe(descs)
	e.m.Current.Describe(descs)
	e.m.Voltage.Describe(descs)
	e.m.Power.Describe(descs)
	e.m.SwitchOn.Describe(descs)
}

func (e *exporter) Collect(c chan<- prometheus.Metric) {
	e.m.Error.Set(0)
	e.m.TotalScrapes.Inc()
	e.m.ScrapeDuration.Reset()
	for _, dev := range *e.devs {
		start := time.Now().UnixMilli()
		labels := prometheus.Labels{"device": dev.Name}
		client := proto.NewClient(dev.Ip, dev.Id, []byte(dev.Key))
		status, err := client.Status()
		if err != nil {
			level.Warn(e.l).Log("msg", "error during scrape", "device", dev.Name, "error", err)
			e.m.ScrapeErrors.With(labels).Inc()
			e.m.Error.Set(1)
		} else {
			ison := 0
			e.m.Current.With(labels).Set(float64(status.Dps.Current) / 1000)
			e.m.Voltage.With(labels).Set(float64(status.Dps.Voltage) / 10)
			e.m.Power.With(labels).Set(float64(status.Dps.Power) / 10)
			if status.Dps.SwitchOn {
				ison = 1
			}
			e.m.SwitchOn.With(labels).Set(float64(ison))
		}
		e.m.ScrapeDuration.With(labels).Observe(float64(time.Now().UnixMilli() - start))
	}
	e.m.Error.Collect(c)
	e.m.ScrapeDuration.Collect(c)
	e.m.TotalScrapes.Collect(c)
	e.m.ScrapeErrors.Collect(c)
	e.m.Current.Collect(c)
	e.m.Voltage.Collect(c)
	e.m.Power.Collect(c)
	e.m.SwitchOn.Collect(c)
}

func NewExporter(devices *[]types.Device, logger log.Logger, m types.Metrics) prometheus.Collector {
	return &exporter{
		m:    m,
		devs: devices,
		l:    logger,
	}
}
