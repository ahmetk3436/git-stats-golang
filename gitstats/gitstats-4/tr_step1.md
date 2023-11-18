# Git Stats Demo

## Amaç
Bu denemede, Git üzerindeki istatistikleri inceleyeceğiz. Lines Of Code (LOC) dediğimiz bir repo üzerindeki bütün kod satırlarını bulacağız. Ayrıca her geliştiricinin kaç adet koda ekleme ve çıkarma yaptığını göreceğiz.

## Kodu İndirme
```bash
git clone https://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@github.com/ahmetk3436/git-stats-golang
```

## Klasöre Gitme

```bash
cd git-stats-golang
```

## Docker Compose'u Çalıştırma
```bash
docker compose up -d
```

## Değişiklik Yapılırsa
Herhangi bir değişiklik yaparsanız, aşağıdaki komutu kullanarak tekrar build alabilirsiniz:

```bash
docker compose up --build -d
```
Bu Docker Compose dosyası, bir adet backend,Redis , Prometheus ve Grafana sunucusunu başlatacaktır. Bu sistemler otomatik olarak birbirine bağlıdır. Backend'in /metrics alanı altında Prometheus için metrik çıktıları bulunmaktadır. Ayrıca, yaml klasörü altında Prometheus, Grafana ve Golang metriklerini takip eden bir dashboard ayakta olacaktır.

## Verileri İnceleme

Bulunduğunuz ortamda 80 portuna giderek benim repolarımın verilerine göz atabilirsiniz. İyi eğlenceler :)