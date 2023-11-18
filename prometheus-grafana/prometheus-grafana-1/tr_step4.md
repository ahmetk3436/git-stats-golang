# Prometheus Metrik Türleri

## Gauge Metriği

Gauge, anlık olarak ölçülen ve belirli bir değeri temsil eden bir Prometheus metrik türüdür. Bu metrik, bir anlık ölçümü ifade eder ve değeri artabilir veya azalabilir.

**Kullanım Alanları:**
- Sistem kaynakları, CPU kullanımı gibi anlık değerlerin izlenmesi.

**PromQL Kullanımı:**
```promql
my_gauge_metric
```

## Counter Metriği
Counter, her artışında birim değer eklenen bir Prometheus metrik türüdür. Genellikle olayların sayısını izlemek için kullanılır.

Kullanım Alanları:

İstek sayıları, hata sayıları gibi olayların sayısını izleme.
PromQL Kullanımı:
``` 
my_counter_metric_total
```

## Histogram Metriği
Histogram, bir dizi ölçümü ve bu ölçümlerin dağılımını izlemek için kullanılan bir Prometheus metrik türüdür. Genellikle ölçümlerin çeşitli yüzdelik dilimlerini incelemek için kullanılır.

Kullanım Alanları:

Servis yanıt süreleri, işleme süreleri gibi ölçümlerin dağılımını izleme.
PromQL Kullanımı:
```
histogram_quantile(0.95, my_histogram_metric)
```
## Summary Metriği
Summary, bir dizi ölçümü ve bu ölçümlerin istatistiksel özetini takip etmek için kullanılan bir Prometheus metrik türüdür. Histograma benzer ancak daha az ayrıntılıdır.

Kullanım Alanları:

Servis yanıt süreleri, işleme süreleri gibi ölçümlerin genel istatistiksel özetini izleme.
PromQL Kullanımı:
```
my_summary_metric_sum
```