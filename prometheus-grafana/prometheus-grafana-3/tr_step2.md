# Grafana Kurulumu

## Grafana İmajını Çekelin
```bash
docker pull grafana/grafana
```

## Docker üzerinde koşturalım
```bash
docker run -d -p 3000:3000 --name grafana grafana/grafana
```

Şuanda 3000 portu üzerinde çalışan bir Grafanamız var !

Sol taraftan 3000 portuna gidelim ve UI kısmını inceleyelim
```text
Default user ve password kısmını internetten araştırın !!!
```