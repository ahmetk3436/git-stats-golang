# Git Stats Golang

[![Go Sürümü](https://img.shields.io/badge/Go-1.21.4-blue.svg)](https://golang.org/)
[![Lisans](https://img.shields.io/badge/Lisans-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Destekleniyor-blue.svg)](https://www.docker.com/)
[![Redis](https://img.shields.io/badge/Redis-Önbellek-red.svg)](https://redis.io/)
[![Prometheus](https://img.shields.io/badge/Prometheus-Metrikler-orange.svg)](https://prometheus.io/)
[![Grafana](https://img.shields.io/badge/Grafana-Dashboard-yellow.svg)](https://grafana.com/)

> Go ile geliştirilmiş kapsamlı bir Git istatistik uygulaması. GitHub ve GitLab depolarına yönelik detaylı analitik, gerçek zamanlı izleme ve önbellekleme yetenekleri sunar.

[🇺🇸 English README](README.md)

## 🚀 Özellikler

### Temel İşlevsellik
- **Çoklu Platform Desteği**: GitHub ve GitLab API'leri ile çalışır
- **Depo Analitiği**: Kapsamlı depo istatistikleri ve metrikleri
- **Commit Analizi**: Detaylı commit geçmişi ve katkıda bulunan analizleri
- **Kod Satırı Sayımı**: Depolar için doğru LOC hesaplaması
- **Katkıda Bulunan İstatistikleri**: Detaylı katkıda bulunan analizi ve sıralaması

### Teknik Özellikler
- **Yüksek Performanslı Önbellekleme**: Gelişmiş yanıt süreleri için Redis tabanlı önbellekleme
- **Gerçek Zamanlı İzleme**: Prometheus metrik entegrasyonu
- **Güzel Dashboard'lar**: Önceden yapılandırılmış Grafana dashboard'ları
- **RESTful API**: Temiz ve iyi dokümante edilmiş API uç noktaları
- **Web Arayüzü**: Modern, duyarlı web kullanıcı arayüzü
- **CLI Desteği**: Otomasyon için komut satırı arayüzü
- **Docker Desteği**: Tamamen konteynerleştirilmiş dağıtım
- **HTTPS Desteği**: Güvenli iletişim için SSL/TLS şifreleme

## 🏗️ Mimari

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Frontend  │    │   Go Backend    │    │   Git Sağlayıcı │
│   (Nginx)       │◄──►│   (API Server)  │◄──►│ GitHub/GitLab   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       ▼                       │
         │              ┌─────────────────┐              │
         │              │     Redis       │              │
         │              │   (Önbellek)    │              │
         │              └─────────────────┘              │
         │                       │                       │
         │                       ▼                       │
         │              ┌─────────────────┐              │
         └─────────────►│   Prometheus    │◄─────────────┘
                        │   (Metrikler)   │
                        └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │    Grafana      │
                        │  (Dashboard)    │
                        └─────────────────┘
```

## 📋 Gereksinimler

- **Go 1.21.4+**
- **Docker & Docker Compose**
- **Git**
- **GitHub Kişisel Erişim Jetonu** (GitHub entegrasyonu için)
- **GitLab Kişisel Erişim Jetonu** (GitLab entegrasyonu için)

## 🛠️ Kurulum

### Seçenek 1: Docker Compose (Önerilen)

1. **Depoyu klonlayın**:
   ```bash
   git clone https://github.com/ahmetk3436/git-stats-golang.git
   cd git-stats-golang
   ```

2. **Ortam değişkenlerini ayarlayın**:
   ```bash
   export GITHUB_TOKEN="github_jetonunuz_buraya"
   export GITLAB_TOKEN="gitlab_jetonunuz_buraya"
   export GITLAB_HOST="https://gitlab.com"  # veya GitLab örneğiniz
   export REDIS_PASSWORD="toor"
   ```

3. **Uygulamayı başlatın**:
   ```bash
   docker-compose up -d
   ```

4. **Servislere erişin**:
   - **Web Arayüzü**: http://localhost
   - **API**: http://localhost:1323
   - **Prometheus**: http://localhost:9090
   - **Grafana**: http://localhost:3000 (admin/admin)

### Seçenek 2: Yerel Geliştirme

1. **Bağımlılıkları yükleyin**:
   ```bash
   go mod download
   ```

2. **Redis'i başlatın** (önbellekleme için gerekli):
   ```bash
   docker run -d --name redis -p 6379:6379 redis:latest redis-server --requirepass toor
   ```

3. **Uygulamayı çalıştırın**:
   ```bash
   # API Modu
   go run cmd/main.go api
   
   # CLI Modu
   go run cmd/main.go cli --help
   ```

## 🔧 Yapılandırma

### Ortam Değişkenleri

| Değişken | Açıklama | Varsayılan | Gerekli |
|----------|----------|------------|----------|
| `GITHUB_TOKEN` | GitHub Kişisel Erişim Jetonu | - | GitHub özellikleri için |
| `GITLAB_TOKEN` | GitLab Kişisel Erişim Jetonu | - | GitLab özellikleri için |
| `GITLAB_HOST` | GitLab örnek URL'si | `https://gitlab.com` | Hayır |
| `REDIS_HOST` | Redis sunucu adresi | `redis:6379` | Hayır |
| `REDIS_PASSWORD` | Redis şifresi | `toor` | Hayır |
| `CORS_ALLOWED_ORIGIN` | CORS izin verilen kaynaklar | `*` | Hayır |

### Erişim Jetonları Oluşturma

#### GitHub Jetonu
1. GitHub Ayarlar → Geliştirici ayarları → Kişisel erişim jetonları'na gidin
2. Bu kapsamlarla yeni jeton oluşturun:
   - `repo` (özel depolar için)
   - `public_repo` (genel depolar için)
   - `read:user` (kullanıcı bilgileri için)

#### GitLab Jetonu
1. GitLab Ayarlar → Erişim Jetonları'na gidin
2. Bu kapsamlarla jeton oluşturun:
   - `read_api`
   - `read_repository`
   - `read_user`

## 📚 API Dokümantasyonu

### GitHub Uç Noktaları

| Metod | Uç Nokta | Açıklama | Parametreler |
|-------|----------|----------|-------------|
| GET | `/api/github/repos` | Tüm depoları getir | `owner` (isteğe bağlı) |
| GET | `/api/github/repo` | Belirli depoyu getir | `projectID` (gerekli) |
| GET | `/api/github/commits` | Depo commit'lerini getir | `projectOwner`, `repoName` |
| GET | `/api/github/contributors` | Depo katkıda bulunanlarını getir | `projectOwner`, `repoName` |
| GET | `/api/github/loc` | Kod satırlarını getir | `projectOwner`, `repoName` |

### GitLab Uç Noktaları

| Metod | Uç Nokta | Açıklama | Parametreler |
|-------|----------|----------|-------------|
| GET | `/api/gitlab/repos` | Tüm depoları getir | `owner` (isteğe bağlı) |
| GET | `/api/gitlab/repo` | Belirli depoyu getir | `projectID` (gerekli) |
| GET | `/api/gitlab/commits` | Depo commit'lerini getir | `projectOwner`, `repoName` |

### Örnek API Çağrıları

```bash
# Tüm GitHub depolarını getir
curl "http://localhost:1323/api/github/repos"

# Belirli depoyu getir
curl "http://localhost:1323/api/github/repo?projectID=sahip/depo-adi"

# Depo commit'lerini getir
curl "http://localhost:1323/api/github/commits?projectOwner=sahip&repoName=depo-adi"

# Depo katkıda bulunanlarını getir
curl "http://localhost:1323/api/github/contributors?projectOwner=sahip&repoName=depo-adi"
```

## 🖥️ CLI Kullanımı

```bash
# GitHub işlemleri
go run cmd/main.go cli --github-token="jetonunuz" --help

# GitLab işlemleri
go run cmd/main.go cli --gitlab-token="jetonunuz" --gitlab-host="https://gitlab.com" --help

# Depo bilgilerini getir
go run cmd/main.go cli --github-token="jetonunuz" repo --owner="kullaniciadi" --repo="depo"
```

## 📊 İzleme ve Metrikler

### Prometheus Metrikleri

Uygulama `/metrics` adresinde şu metrikleri sunar:

- `gits_api_calls_total`: Toplam API çağrı sayısı
- `gits_repository_fetches_total`: Toplam depo getirme denemeleri
- `gits_api_call_duration_seconds`: API çağrı süresi histogramı

### Grafana Dashboard

Önceden yapılandırılmış dashboard şunları içerir:
- API yanıt süreleri
- İstek oranları
- Hata oranları
- Önbellek isabet oranları
- Depo istatistikleri

## 🧪 Test Etme

```bash
# Tüm testleri çalıştır
go test ./...

# Kapsama ile testleri çalıştır
go test -cover ./...

# Belirli paket testlerini çalıştır
go test ./pkg/api/
go test ./pkg/repository/
```

## 🏗️ Proje Yapısı

```
.
├── cmd/                    # Uygulama giriş noktaları
│   ├── main.go            # Ana uygulama
│   ├── cert.pem           # SSL sertifikası (sadece geliştirme)
│   └── key.pem            # SSL özel anahtarı (sadece geliştirme)
├── pkg/                    # Genel paketler
│   ├── api/               # HTTP API işleyicileri
│   ├── cli/               # CLI komutları
│   ├── common_types/      # Paylaşılan veri yapıları
│   ├── interfaces/        # Arayüz tanımları
│   ├── prometheus/        # Metrik tanımları
│   └── repository/        # Git sağlayıcı uygulamaları
├── internal/              # Özel paketler
│   └── inmemory_db.go     # Redis istemcisi
├── web/                   # Frontend varlıkları
│   ├── index.html         # Web arayüzü
│   └── api.js             # Frontend JavaScript
├── yaml/                  # Yapılandırma dosyaları
│   ├── prometheus.yml     # Prometheus yapılandırması
│   ├── dashboard.yml      # Grafana dashboard yapılandırması
│   └── golang.json        # Grafana dashboard JSON
├── docker-compose.yaml    # Docker kompozisyonu
├── Dockerfile             # Uygulama konteyneri
└── Dockerfile-Redis       # Redis konteyneri
```

## 🤝 Katkıda Bulunma

1. **Depoyu fork edin**
2. **Özellik dalı oluşturun**: `git checkout -b feature/harika-ozellik`
3. **Değişikliklerinizi commit edin**: `git commit -m 'Harika özellik ekle'`
4. **Dalı push edin**: `git push origin feature/harika-ozellik`
5. **Pull Request açın**

### Geliştirme Yönergeleri

- Go en iyi uygulamalarını ve konvansiyonlarını takip edin
- Yeni özellikler için kapsamlı testler yazın
- API değişiklikleri için dokümantasyonu güncelleyin
- Logrus ile yapılandırılmış loglama kullanın
- Uygun hata işleme uygulayın
- Yeni uç noktalar için Prometheus metrikleri ekleyin

## 📄 Lisans

Bu proje MIT Lisansı altında lisanslanmıştır - detaylar için [LICENSE](LICENSE) dosyasına bakın.

## 🙏 Teşekkürler

- [Go](https://golang.org/) - Programlama dili
- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP yönlendiricisi
- [Redis](https://redis.io/) - Bellek içi veri yapısı deposu
- [Prometheus](https://prometheus.io/) - İzleme sistemi
- [Grafana](https://grafana.com/) - Analitik platformu
- [Docker](https://www.docker.com/) - Konteynerleştirme platformu

## 📞 Destek

Herhangi bir sorunuz varsa veya yardıma ihtiyacınız varsa, lütfen:

1. [Dokümantasyonu](#-api-dokümantasyonu) kontrol edin
2. Mevcut [sorunları](https://github.com/ahmetk3436/git-stats-golang/issues) arayın
3. Yeni bir [sorun](https://github.com/ahmetk3436/git-stats-golang/issues/new) oluşturun

---

**❤️ ile [Ahmet Coşkun Kızılkaya](https://github.com/ahmetk3436) tarafından yapılmıştır**