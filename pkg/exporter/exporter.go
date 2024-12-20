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
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rkosegi/tuya-smartplug-exporter/pkg/proto"
)

type exporter struct {
	m    Metrics
	devs *[]Device
	l    *slog.Logger
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	e.m.Error.Describe(ch)
	e.m.TotalScrapes.Describe(ch)
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.m.Error.Set(0)
	e.m.TotalScrapes.Inc()

	var wg sync.WaitGroup
	defer wg.Wait()
	for _, dev := range *e.devs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := newDeviceMetrics()
			start := time.Now()
			labels := prometheus.Labels{"device": dev.Name}
			e.l.Debug("Connecting to device", "device", dev.Name, "address", dev.Ip)
			client := proto.NewClient(dev.Ip, dev.Id, []byte(dev.Key))
			status, err := client.Status()
			e.l.Debug("Status of device", "device", dev.Name, "status", status)
			if err != nil {
				e.l.Warn("error during scrape", "device", dev.Name, "error", err)
				m.ScrapeErrors.With(labels).Inc()
				e.m.Error.Set(1)
			} else {
				ison := 0
				m.Current.With(labels).Set(float64(status.Dps.Current) / 1000)
				m.Voltage.With(labels).Set(float64(status.Dps.Voltage) / 10)
				m.Power.With(labels).Set(float64(status.Dps.Power) / 10)
				if status.Dps.SwitchOn {
					ison = 1
				}
				m.SwitchOn.With(labels).Set(float64(ison))
			}
			m.ScrapeDuration.With(labels).Observe(time.Since(start).Seconds())

			m.SwitchOn.Collect(ch)
			m.Current.Collect(ch)
			m.Voltage.Collect(ch)
			m.Power.Collect(ch)
			m.ScrapeDuration.Collect(ch)
		}()
	}
	e.m.Error.Collect(ch)
	e.m.TotalScrapes.Collect(ch)
}

func New(devices *[]Device, logger *slog.Logger) prometheus.Collector {
	return &exporter{
		m:    newCommonMetrics(),
		devs: devices,
		l:    logger,
	}
}
