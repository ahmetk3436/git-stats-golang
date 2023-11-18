# Prometheus Sorgu Dili (PromQL)
PromQL, Prometheus'un metrik verilerini sorgulamak ve analiz etmek için kullandığı bir sorgu ve ifade dilidir. Prometheus tarafından toplanan metrik değerlerini almak ve analiz etmek için kullanıcılara olanak tanır. İşte PromQL'nin temel kullanımı ve bazı yaygın fonksiyonlarının genel bir bakışı:

## Temel Kullanım
PromQL sorguları, belirli bir metriğin değerlerini zaman içinde izlemek ve analiz etmek için kullanılır. Temel bir PromQL sorgusu şu şekildedir:

```
my_metric_name
```
Bu sorgu, my_metric_name adlı bir metriği temsil eder ve zaman içindeki değerlerini getirir.

## Temel Operatörler
PromQL, matematiksel operatörleri destekler:

+ (toplama)
- (çıkarma)
* (çarpma)
  / (bölme)
  Örneğin:
```
my_metric_a + my_metric_b
 ```
  Bu sorgu, my_metric_a ve my_metric_b metriklerinin toplamını getirir.

## Vektör Seçimi
PromQL, belirli bir zaman aralığındaki metrik verilerini seçmek için vektör seçimini kullanır. Vektör seçimi, sürekli bir zaman serisi verisi döndürür.

```
my_metric_name{label_name="value"}
```
Bu sorgu, label_name etiketi value olan my_metric_name metriğini getirir.

## Temel Fonksiyonlar

rate() fonksiyonu, belirli bir metriğin birim zamandaki değişimini hesaplar.

```
rate(my_metric_name[5m])
```
Bu sorgu, son 5 dakikadaki my_metric_name metriğinin birim zamandaki değişimini getirir.


sum() fonksiyonu, belirli bir vektördeki değerlerin toplamını hesaplar.

```
sum(my_metric_name)
```
Bu sorgu, my_metric_name vektöründeki tüm değerlerin toplamını getirir.


avg() fonksiyonu, belirli bir vektördeki değerlerin ortalamasını hesaplar.

```
avg(my_metric_name)
```
Bu sorgu, my_metric_name vektöründeki değerlerin ortalamasını getirir.

[Dökümantasyon](https://prometheus.io/docs/prometheus/latest/querying/basics/)