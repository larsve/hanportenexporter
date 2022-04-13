package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type (
	semData struct {
		received time.Time
		semData  SmartEnergyMeterData
	}
	promData struct {
		mu              sync.Mutex
		data            map[string]semData
		obisDescription *prometheus.Desc
	}
)

var (
	mux    *http.ServeMux
	server *http.Server
)

func startMetricsServer() {
	log.Println("Starting Prometheus metrics server...")
	conn, err := net.Listen("tcp", ":9102")
	if err != nil {
		panic(err)
	}

	mux = http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server = &http.Server{Handler: mux}
	go runMetricsWebServer(conn)
	log.Printf("Prometheus metrics available on http://%s/metrics", conn.Addr().String())
}

func stopMetricsServer(ctx context.Context) {
	log.Println("Stopping Prometheus metrics server...")
	err := server.Shutdown(ctx)
	if err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("Prometheus metrics server stopped")
}

func runMetricsWebServer(conn net.Listener) {
	err := server.Serve(conn)
	if err != nil && err != http.ErrServerClosed {
		log.Printf("Server ended with an error: %v", err)
	}
}

func (pd *promData) Collect(ch chan<- prometheus.Metric) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	sendGaugeValue := func(desc *prometheus.Desc, value float64, labels ...string) {
		m, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, value, labels...)
		if err == nil {
			ch <- m
		}
	}

	for _, i := range pd.data {
		if time.Since(i.received) > 20*time.Second {
			// Skip data older than 20 seconds
			continue
		}

		ident := i.semData.ID

		for _, v := range i.semData.Values {
			sendGaugeValue(pd.obisDescription, v.Value, ident, v.OBIS, v.Unit)
		}
	}
}

func (pd *promData) Describe(ch chan<- *prometheus.Desc) {
	registerDesc := func(name, desc string, labels []string) *prometheus.Desc {
		pd := prometheus.NewDesc(
			prometheus.BuildFQName("hanporten", "", name),
			desc,
			labels,
			nil,
		)
		ch <- pd
		return pd
	}
	pd.obisDescription = registerDesc("OBIS", "", []string{"ident", "obis", "unit"})
}

func (pd *promData) Write(sem *SmartEnergyMeterData) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	if pd.data == nil {
		pd.data = make(map[string]semData)
	}

	pd.data[sem.ID] = semData{received: time.Now(), semData: *sem}
}
