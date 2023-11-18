# Prometheus Query Language (PromQL)
PromQL is a query language and expression language used by Prometheus for querying and analyzing metric data. It allows users to retrieve and analyze metric values collected by Prometheus. Here is an overview of the basic usage and some common functions in PromQL:

Basic Usage
PromQL queries are used to track and analyze the values of a specific metric over time. A basic PromQL query looks like this:
```
my_metric_name
```
This query represents a metric named my_metric_name and fetches its values over time.

## Basic Operators
PromQL supports mathematical operators:

+ (addition)
- (subtraction)
* (multiplication)
  / (division)
  For example:

```
my_metric_a + my_metric_b
```
This query fetches the sum of my_metric_a and my_metric_b metrics.

## Vector Selection
PromQL uses vector selection to select metric data over a specific time range. Vector selection returns a continuous time series data.
```
my_metric_name{label_name="value"}
```
This query fetches the my_metric_name metric with the label_name label set to "value."

## Basic Functions
rate()
The rate() function calculates the rate of change of a specific metric over a unit time.

```
rate(my_metric_name[5m])
```
This query fetches the rate of change of the my_metric_name metric over the last 5 minutes.
```
sum()
```
The sum() function calculates the total sum of values in a specific vector.

```
sum(my_metric_name)
```
This query fetches the total sum of values in the my_metric_name vector.
```
avg()
```
The avg() function calculates the average of values in a specific vector.

```
avg(my_metric_name)
```
This query fetches the average of values in the my_metric_name vector.

[Documentation](https://prometheus.io/docs/prometheus/latest/querying/basics/)