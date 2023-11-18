# Alert Manager Nedir?

**Alert Manager**, açık kaynaklı bir uyarı yönetim aracıdır ve genellikle Prometheus ile birlikte kullanılır. Prometheus tarafından üretilen uyarıları işler ve bu uyarıları belirli kurallara göre yönetir. Alertmanager, sistem operatörlerine ve geliştiricilere uyarıları düzenleme, filtreleme ve daha fazla kontrole sahip olma imkanı tanır.

### Nasıl Çalışır?

1. **Uyarı Üretimi:** Prometheus, belirli bir metrik veya durumun belirli bir eşiği aştığında uyarılar üretir.

2. **Uyarı İşleme:** Üretilen uyarılar, Alertmanager tarafından alınır ve işlenir.

3. **Filtreleme ve Yönlendirme:** Alert Manager, tanımlanan kurallar ve önceden belirlenmiş konfigürasyonlara göre uyarıları filtreler ve belirli hedeflere yönlendirir.

4. **İnsanlar veya Sistemlere Bildirim:** Alert Manager, filtrelenen uyarıları belirli bildirim kanallarına (E-posta, Slack, Webhook vb.) ileterek ilgili kişilere veya sistemlere bildirimde bulunur.

## Örnek Yaml
```yaml
route:
  group_wait: 10s
  group_interval: 5m
  repeat_interval: 3h
  receiver: 'slack-notifications'

receivers:
- name: 'slack-notifications'
  slack_configs:
  - send_resolved: true
    username: 'alertmanager'
    channel: '#alerts'
```