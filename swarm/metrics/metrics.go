package metrics

import "github.com/prometheus/client_golang/prometheus"

var HttpRequestsHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_request_seconds",
	Help: "Time taken to handle HTTP requests",
}, []string{})
