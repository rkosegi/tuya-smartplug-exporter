# Tuya exporter for prometheus

Prometheus exporter for [Tuya](https://iot.tuya.com/)-based smart plug devices.
Tested with Immax Neo Lite smart plug.

![smartplug](docs/smartplug.jpg)

### Setup

- Obtain device ID and key, [here is excellent guide](https://github.com/codetheweb/tuyapi/blob/master/docs/SETUP.md) by @codetheweb
- Populate config file, [here is an example](config.yaml):

```yaml
- name: plug-kitchen-1
  id: 87e98a987b87b12354a54c
  key: 0987654321abcdef
  ip: 192.168.1.5
```

### Example output

```shell
curl --silent localhost:9999/metrics | grep ^tuya
tuya_smartplug_current{device="livingroom-1"} 0.093
tuya_smartplug_exporter_build_info{branch="",goversion="go1.18.6",revision="",version=""} 1
tuya_smartplug_last_scrape_error 0
tuya_smartplug_power{device="livingroom-1"} 8.6
tuya_smartplug_scrape_duration_sum{device="livingroom-1"} 73
tuya_smartplug_scrape_duration_count{device="livingroom-1"} 1
tuya_smartplug_scrapes_total 60
tuya_smartplug_voltage{device="livingroom-1"} 242.9
```

### Sample Grafana dashboard

![dashboard](docs/dashboard.jpg)


### Acknowledgment

Portion of client code is inspired by @jasonacox implementation in [powermonitor](https://github.com/jasonacox/powermonitor)