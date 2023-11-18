# Prometheus Kurulumu

Bu senaryomuzda Prometheus kurulumu ile başlayacağız. Ardında Grafana kurulumumuzu da yaparak bağlamalarımızı yaparak,gerçek bir örnek göreceğiz.

## Docker
Bu çalışmalarımızı Docker üzerinde yapacağımız için makinemizde docker olmalıdır. Aşağıdaki yönergeyi takip ederek eğer ki makinemizde docker yoksa kurulumumuzu yapalım.

```bash
docker --version
```
Bu çıktı sonucunda docker versiyonu göremiyorsak bu aşamaları takip edelim.
Sistemimizi güncelleyelim ve dockeri ekleyelim.
```bash
apk update && apk add docker
```
Servis üzerinde dockeri başlatalım ve restart ettiğimizdede yeniden başlaması için gerekli fonksiyonu ekleyelim.
```bash
service docker start
rc-update add docker boot
```
Bazı durumlarda dockeri user olarak eklemek gerekli olabilir,bir sonraki aşamada çıktı göremezsek bu komutu çalıştıralım.
```bash
addgroup $USER docker
```
Her komutumuz başarılı şekildeyse Hello,World! çıktısını göreceğiz.
```bash
docker run hello-world
```

Prometheus imagemizi çekelim
```bash
docker pull prom/prometheus
```

Prometheus containerimizi başlatalım
```bash
docker run -d -p 9090:9090 --name prometheus prom/prometheus
```

Veeee sonunda Prometheus'umuz ayakta !