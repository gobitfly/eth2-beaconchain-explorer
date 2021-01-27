package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of requests by path, method and status_code.",
	}, []string{"path", "method", "status_code"})
	HttpRequestsInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current requests being served.",
	}, []string{"path", "method"})
	HttpRequestsDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_requests_duration",
		Help: "Duration of HTTP requests in seconds by path and method.",
		// Buckets: []float64{0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
	}, []string{"path", "method"})
	ExporterExportEpochDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "exporter_export_epoch_duration",
		Help: "Time it took to export an epoch.",
		// Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 50, 100},
		Buckets: prometheus.ExponentialBuckets(1, 4, 6),
	})
)

// HttpMiddleware implements mux.MiddlewareFunc.
// This middleware uses the path template, so the label value will be /obj/{id} rather than /obj/123 which would risk a cardinality explosion.
// See https://www.robustperception.io/prometheus-middleware-for-gorilla-mux
func HttpMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := mux.CurrentRoute(r)
		path, err := route.GetPathTemplate()
		if err != nil {
			path = "UNDEFINED"
		}
		method := strings.ToUpper(r.Method)
		HttpRequestsInFlight.WithLabelValues(path, method).Inc()
		defer HttpRequestsInFlight.WithLabelValues(path, method).Dec()
		d := &responseWriterDelegator{ResponseWriter: w}
		next.ServeHTTP(d, r)
		status := strconv.Itoa(d.status)
		HttpRequestsTotal.WithLabelValues(path, method, status).Inc()
		HttpRequestsDuration.WithLabelValues(path, method).Observe(time.Since(start).Seconds())
	})
}

type responseWriterDelegator struct {
	http.ResponseWriter
	status      int
	written     int64
	wroteHeader bool
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

// Serve serves prometheus metrics on the given address under /metrics
func Serve(addr string) error {
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>prometheus-metrics</title></head>
<body>
<h1>prometheus-metrics</h1>
<p><a href='/metrics'>metrics</a></p>
</body>
</html>`))
	}))
	srv := &http.Server{
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		Handler:      router,
		Addr:         addr,
	}

	return srv.ListenAndServe()
}
