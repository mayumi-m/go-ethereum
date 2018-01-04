package metrics

import "github.com/prometheus/client_golang/prometheus"

var HttpRequestsHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_request_seconds",
	Help: "Time taken to handle HTTP requests.",
}, []string{})

var NumberOfKadPeersGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	//Namespace: "our_company",
	//Subsystem: "blob_storage",
	Name: "number_kad_peers",
	Help: "Number of kad peers.",
})

var NumberOfPeersGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	//Namespace: "our_company",
	//Subsystem: "blob_storage",
	Name: "number_peers",
	Help: "Number of peers.",
})
