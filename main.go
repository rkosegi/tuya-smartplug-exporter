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

package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/rkosegi/tuya-smartplug-exporter/pkg/exporter"
	"github.com/rkosegi/tuya-smartplug-exporter/pkg/types"
	"gopkg.in/yaml.v3"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	pv "github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

const (
	progName = "tuya_smartplug_exporter"
)

var (
	webConfig  = webflag.AddFlags(kingpin.CommandLine, ":9999")
	metricPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	configFile = kingpin.Flag("config.file", "Path to YAML file with configuration").Default("config.yaml").String()
)

func init() {
	prometheus.MustRegister(version.NewCollector(progName))
}

func newHandler(devices *[]types.Device, logger *slog.Logger, m types.Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewRegistry()
		registry.MustRegister(exporter.NewExporter(devices, logger, m))

		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func main() {
	promlogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(pv.Print(progName))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promslog.New(promlogConfig)
	logger.Info(fmt.Sprintf("Starting %s", progName), "version", pv.Info(), "config", *configFile)

	devs, err := loadConfig(*configFile)

	if err != nil {
		logger.Error("Error reading configuration", "err", err)
		os.Exit(1)
	}

	if len(*devs) == 0 {
		logger.Error("no devices configured")
		os.Exit(1)
	} else {
		logger.Info(fmt.Sprintf("Configured %d devices", len(*devs)))
	}

	var landingPage = []byte(`<html>
<head><title>tuya smartplug exporter</title></head>
<body>
<h1>tuya smartplug exporter</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
`)

	handlerFunc := newHandler(devs, logger, exporter.NewMetrics())
	http.Handle(*metricPath, promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handlerFunc))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err = w.Write(landingPage); err != nil {
			logger.Error("Unable to write page content", "err", err)
		}
	})
	srv := &http.Server{}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		logger.Error("Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}

func loadConfig(path string) (*[]types.Device, error) {
	if bytes, err := os.ReadFile(path); err != nil {
		return nil, err
	} else {
		devs := make([]types.Device, 0)
		if err = yaml.Unmarshal(bytes, &devs); err != nil {
			return nil, err
		} else {
			return &devs, nil
		}
	}
}
