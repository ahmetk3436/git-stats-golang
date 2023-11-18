#!/bin/bash
sleep 10
while true; do
  # Ngrok'tan halka açık URL'yi alın
  NGROK_PUBLIC_URL=$(curl -s http://localhost:4040/api/tunnels | jq -r '.tunnels[0].public_url')

  # Halka açık URL'yi Prometheus konfigürasyon dosyasına yerine koyun
  sed -i "s|TARGET_PLACEHOLDER|$NGROK_PUBLIC_URL|g" prometheus.yml

  # Belirli bir süre bekleyin (örneğin, her dakika kontrol)
  sleep 60
done
