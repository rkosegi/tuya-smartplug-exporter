package types

import "github.com/prometheus/client_golang/prometheus"

type PlugInfo struct {
	Voltage float64
	Power   float64
	Current float64
	On      float64
}

type Device struct {
	Name string
	Id   string
	Key  string
	Ip   string
}

type Metrics struct {
	TotalScrapes   prometheus.Counter
	ScrapeDuration *prometheus.SummaryVec
	ScrapeErrors   *prometheus.CounterVec
	Error          prometheus.Gauge
	Current        *prometheus.GaugeVec
	Voltage        *prometheus.GaugeVec
	Power          *prometheus.GaugeVec
	SwitchOn       *prometheus.GaugeVec
}
