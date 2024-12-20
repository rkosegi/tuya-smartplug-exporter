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
	TotalScrapes prometheus.Counter
	Error        prometheus.Gauge
}
