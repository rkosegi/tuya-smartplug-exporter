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
	"github.com/rkosegi/tuya-proto/proto"
	"github.com/rkosegi/tuya-smartplug-exporter/pkg/internal"
	"github.com/samber/lo"
)

type exporter struct {
	m       GlobalMetrics
	cfg     *internal.ConfigSpec
	l       *slog.Logger
	clients map[string]internal.Client
}

func (e *exporter) Describe(ch chan<- *prometheus.Desc) {
	e.m.Error.Describe(ch)
	e.m.TotalScrapes.Describe(ch)
}

func (e *exporter) clientForDevice(dc internal.DeviceConnectionSpec) internal.Client {
	ver := proto.Version31
	if dc.Protocol == "tuya3.4" {
		ver = proto.Version34
	}
	if dc.ConnectTimeout == 0 {
		dc.ConnectTimeout = time.Second * 10
	}
	if dc.ReadTimeout == 0 {
		dc.ReadTimeout = time.Second * 10
	}
	if dc.WriteTimeout == 0 {
		dc.WriteTimeout = time.Second * 10
	}
	return internal.NewClient(ver, dc.Address, []byte(dc.Key),
		internal.WithTimeout(dc.ConnectTimeout),
		internal.WithReadTimeout(dc.ReadTimeout),
		internal.WithWriteTimeout(dc.WriteTimeout),
		internal.WithLogger(e.l.With("address", dc.Address, "protocol", dc.Protocol)),
	)
}

func (e *exporter) statusForDevice(name string) (*internal.DpQueryResponse, *internal.ProtoStats, error) {
	cl := e.clients[name]
	dc := e.cfg.Devices[name]
	var err error
	if !cl.IsConnected() {
		if err = cl.Connect(); err != nil {
			return nil, new(cl.Stats()), err
		}
	}
	if dc.Protocol == "tuya3.4" {
		if err = cl.Send(proto.CmdIdTypeDpQueryNew, make(map[string]interface{})); err != nil {
			return nil, new(cl.Stats()), err
		}
	} else {
		if err = cl.Send(proto.CmdIdTypeDpQuery, internal.DpQueryRequest{
			GwId:  dc.Id,
			DevId: dc.Id,
		}); err != nil {
			return nil, new(cl.Stats()), err
		}
	}
	defer func() {
		_ = cl.Close()
	}()
	var out internal.DpQueryResponse
	if err = cl.Read(&out); err != nil {
		return nil, new(cl.Stats()), err
	}
	return &out, new(cl.Stats()), nil
}

func (e *exporter) Collect(ch chan<- prometheus.Metric) {
	e.m.Error.Set(0)
	startAny := time.Now()
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		e.m.TotalScrapes.Observe(time.Since(startAny).Seconds())
		e.m.Error.Collect(ch)
		e.m.TotalScrapes.Collect(ch)
	}()
	for dname := range e.cfg.Devices {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := newDeviceMetrics()
			start := time.Now()
			labels := prometheus.Labels{"device": dname}
			status, stats, err := e.statusForDevice(dname)
			m.ReadPackets.With(labels).Add(float64(stats.ReadPkts))
			m.SentPackets.With(labels).Add(float64(stats.SentPkts))
			if stats.ReadErrs > 0 {
				m.ReadErrors.With(labels).Add(float64(stats.ReadErrs))
			}
			if stats.SentErrs > 0 {
				m.SentErrors.With(labels).Add(float64(stats.SentErrs))
			}
			if err != nil {
				e.l.Warn("error during scrape", "device", dname, "error", err)
				m.ScrapeErrors.With(labels).Inc()
				e.m.Error.Set(1)
			} else {
				e.l.Debug("Status of device", "device", dname, "status", status.Dps)
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
			m.ReadPackets.Collect(ch)
			m.SentPackets.Collect(ch)
			m.ReadErrors.Collect(ch)
			m.SentErrors.Collect(ch)
			m.ScrapeDuration.Collect(ch)
		}()
	}
}

func New(cfg *internal.ConfigSpec, logger *slog.Logger) prometheus.Collector {
	e := &exporter{
		m:   newCommonMetrics(),
		cfg: cfg,
		l:   logger,
	}
	// create mapping dev-name to client
	e.clients = lo.MapEntries(cfg.Devices, func(name string, dc internal.DeviceConnectionSpec) (string, internal.Client) {
		return name, e.clientForDevice(dc)
	})
	return e
}
