package bridge

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	pollsTotal    *prometheus.CounterVec
	pollDuration  prometheus.Histogram
	devicesTotal  prometheus.Gauge
	commandsTotal *prometheus.CounterVec
}

func newMetrics(reg *prometheus.Registry) *metrics {
	m := &metrics{
		pollsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "hologram_mqtt_polls_total",
				Help: "Total number of Hologram API poll cycles.",
			},
			[]string{"status"},
		),

		pollDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "hologram_mqtt_poll_duration_seconds",
				Help:    "Duration of each poll cycle in seconds.",
				Buckets: prometheus.DefBuckets,
			},
		),

		devicesTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "hologram_mqtt_devices_total",
				Help: "Number of known devices after the last poll.",
			},
		),

		commandsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "hologram_mqtt_commands_total",
				Help: "Total number of commands received via MQTT.",
			},
			[]string{"action"},
		),
	}

	reg.MustRegister(
		m.pollsTotal,
		m.pollDuration,
		m.devicesTotal,
		m.commandsTotal,
	)

	return m
}
