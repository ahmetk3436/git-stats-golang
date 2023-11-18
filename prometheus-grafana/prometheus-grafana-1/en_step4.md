# Prometheus Metric Types

## Gauge Metric

Gauge is a Prometheus metric type that represents an instantaneous measurement with a specific value. This metric indicates an instantaneous measurement, and its value can increase or decrease.

**Use Cases:**
- Monitoring instantaneous values such as system resources, CPU usage, etc.

**PromQL Usage:**
```promql
my_gauge_metric
```
## Counter Metric

Counter is a Prometheus metric type that increments by one each time. It is commonly used to track the number of events.

Use Cases:

Monitoring event counts such as request numbers, error counts, etc.
PromQL Usage:
```
my_counter_metric_total
```
## Histogram Metric
Histogram is a Prometheus metric type used to track a series of measurements and their distribution. It is commonly used to examine various percentiles of measurements.

Use Cases:

Monitoring the distribution of measurements such as service response times, processing times, etc.
PromQL Usage:
```
histogram_quantile(0.95, my_histogram_metric)
```
## Summary Metric
Summary is a Prometheus metric type used to track a series of measurements and their statistical summary. It is similar to a histogram but less detailed.

Use Cases:

Monitoring the overall statistical summary of measurements such as service response times, processing times, etc.
PromQL Usage:
```
my_summary_metric_sum
```