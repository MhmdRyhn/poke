// Concept taken from the following link:
// https://gabrieltanner.org/blog/collecting-prometheus-metrics-in-golang/

package poke

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	counterVector   *prometheus.CounterVec
	histogramVector *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	metrics := &Metrics{}
	return metrics
}

func (metrics *Metrics) WithCounterVector(namespace string, subsystem string, metricsName string) *Metrics {
	opts := prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      metricsName,
	}
	labels := []string{"path", "method", "status_code"}
	metrics.counterVector = prometheus.NewCounterVec(opts, labels)
	prometheus.MustRegister(metrics.counterVector)
	return metrics
}

func (metrics *Metrics) WithHistogramVector(
	namespace string, subsystem string, metricsName string,
	buckets []float64,
) *Metrics {
	opts := prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      metricsName,
		Buckets:   buckets,
	}
	labels := []string{"path", "method", "status_code"}
	metrics.histogramVector = prometheus.NewHistogramVec(opts, labels)
	prometheus.MustRegister(metrics.histogramVector)
	return metrics
}

func (metrics *Metrics) areAllMetricsTypeNil() bool {
	return metrics.counterVector == nil && metrics.histogramVector == nil
}

//
// If none of the metric type is set by calling at least one of `WithCounterVector`,
// `WithHistogramVector` methods, this method will PANIC.
//
func (metrics *Metrics) Middleware(next http.Handler) http.Handler {
	if metrics.areAllMetricsTypeNil() {
		panic(
			"At least one of `WithCounterVector`, `WithHistogramVector` methods " +
				"must be called before calling `Middleware`.",
		)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := newResponseWriter(w)
		startTime := time.Now()
		next.ServeHTTP(wr, r)
		if metrics.counterVector != nil {
			labels := []string{r.URL.Path, r.Method, fmt.Sprintf("%d", wr.statusCode)}
			metrics.counterVector.WithLabelValues(labels...).Inc()
		}
		if metrics.histogramVector != nil {
			duration := time.Since(startTime) / time.Second
			labels := []string{r.URL.Path, r.Method, fmt.Sprintf("%d", wr.statusCode)}
			metrics.histogramVector.WithLabelValues(labels...).Observe(float64(duration))
		}
	})
}
