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
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/rkosegi/tuya-smartplug-exporter/pkg/exporter"
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
	progName = "tuya smartplug exporter"
)

var (
	webConfig             = webflag.AddFlags(kingpin.CommandLine, ":9999")
	telemetryPath         = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	configFile            = kingpin.Flag("config.file", "Path to YAML file with configuration").Default("config.yaml").String()
	disableDefaultMetrics = kingpin.Flag("disable-default-metrics", "Exclude default metrics about the exporter itself (promhttp_*, process_*, go_*).").Bool()
	errNoDevs             = errors.New("no devices configured")
)

func main() {
	promlogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)

	kingpin.Version(pv.Print(progName))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promslog.New(promlogConfig)

	logger.Info("Exporter starting", "name", progName, "version", pv.Info(), "config.file", *configFile)
	logger.Info("Build context", "build_context", pv.BuildContext())

	devs, err := loadConfig(*configFile)

	if err != nil {
		logger.Error("Error reading configuration", "err", err, "config.file", *configFile)
		os.Exit(1)
	}

	r := prometheus.NewRegistry()
	r.MustRegister(version.NewCollector(strings.ReplaceAll(progName, " ", "_")))
	r.MustRegister(exporter.New(devs, logger))

	logger.Info("Device list loaded", "count", len(*devs))
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{r},
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		},
	)

	if !*disableDefaultMetrics {
		r.MustRegister(collectors.NewGoCollector())
		r.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		handler = promhttp.InstrumentMetricHandler(
			r, handler,
		)
	}

	landingPage, err := web.NewLandingPage(web.LandingConfig{
		Name:        progName,
		Description: "Prometheus Exporter for smart plug devices",
		Version:     pv.Info(),
		Links: []web.LandingLinks{
			{
				Address: *telemetryPath,
				Text:    "Metrics",
			},
			{
				Address: "/health",
				Text:    "Health",
			},
		},
	})
	if err != nil {
		logger.Error("Couldn't create landing page", "err", err)
		os.Exit(1)
	}
	http.Handle("/", landingPage)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	http.Handle(*telemetryPath, handler)

	srv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err = web.ListenAndServe(srv, webConfig, logger); err != nil {
		logger.Error("Error starting server", "err", err)
		os.Exit(1)
	}
}

func loadConfig(path string) (*[]exporter.Device, error) {
	var (
		err   error
		bytes []byte
		out   []exporter.Device
	)
	bytes, err = os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(bytes, &out)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, errNoDevs
	}
	return &out, nil
}
