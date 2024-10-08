//go:build windows

package logon

import (
	"errors"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus-community/windows_exporter/pkg/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

const Name = "logon"

type Config struct{}

var ConfigDefaults = Config{}

// A Collector is a Prometheus Collector for WMI metrics.
type Collector struct {
	config Config
	logger log.Logger

	logonType *prometheus.Desc
}

func New(logger log.Logger, config *Config) *Collector {
	if config == nil {
		config = &ConfigDefaults
	}

	c := &Collector{
		config: *config,
	}

	c.SetLogger(logger)

	return c
}

func NewWithFlags(_ *kingpin.Application) *Collector {
	return &Collector{}
}

func (c *Collector) GetName() string {
	return Name
}

func (c *Collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *Collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *Collector) Close() error {
	return nil
}

func (c *Collector) Build() error {
	c.logonType = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "logon_type"),
		"Number of active logon sessions (LogonSession.LogonType)",
		[]string{"status"},
		nil,
	)
	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *Collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", "failed collecting user metrics", "err", err)
		return err
	}
	return nil
}

// Win32_LogonSession docs:
// - https://docs.microsoft.com/en-us/windows/win32/cimwin32prov/win32-logonsession
type Win32_LogonSession struct {
	LogonType uint32
}

func (c *Collector) collect(ch chan<- prometheus.Metric) error {
	var dst []Win32_LogonSession
	q := wmi.QueryAll(&dst, c.logger)
	if err := wmi.Query(q, &dst); err != nil {
		return err
	}
	if len(dst) == 0 {
		return errors.New("WMI query returned empty result set")
	}

	// Init counters
	system := 0
	interactive := 0
	network := 0
	batch := 0
	service := 0
	proxy := 0
	unlock := 0
	networkcleartext := 0
	newcredentials := 0
	remoteinteractive := 0
	cachedinteractive := 0
	cachedremoteinteractive := 0
	cachedunlock := 0

	for _, entry := range dst {
		switch entry.LogonType {
		case 0:
			system++
		case 2:
			interactive++
		case 3:
			network++
		case 4:
			batch++
		case 5:
			service++
		case 6:
			proxy++
		case 7:
			unlock++
		case 8:
			networkcleartext++
		case 9:
			newcredentials++
		case 10:
			remoteinteractive++
		case 11:
			cachedinteractive++
		case 12:
			cachedremoteinteractive++
		case 13:
			cachedunlock++
		}
	}

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(system),
		"system",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(interactive),
		"interactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(network),
		"network",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(batch),
		"batch",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(service),
		"service",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(proxy),
		"proxy",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(unlock),
		"unlock",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(networkcleartext),
		"network_clear_text",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(newcredentials),
		"new_credentials",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(remoteinteractive),
		"remote_interactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(cachedinteractive),
		"cached_interactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(remoteinteractive),
		"cached_remote_interactive",
	)

	ch <- prometheus.MustNewConstMetric(
		c.logonType,
		prometheus.GaugeValue,
		float64(cachedunlock),
		"cached_unlock",
	)
	return nil
}
