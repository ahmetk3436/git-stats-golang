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
## CLI için Docker Build
```bash
docker build -t app .
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

# Özel API Noktaları

Uygulamada bulunan bazı özel API noktaları:

## GitHub Noktaları

1. **Tüm Depoları Al:**
    - **Nokta:** `/api/github/repos`
    - **Açıklama:** "ahmetk3436" kullanıcı adındaki hesaptaki tüm genel depoları alır.

2. **Proje Kimliğine Göre Bir Depo Al:**
    - **Nokta:** `/api/github/repo?projectID=<projectID>`
    - **Açıklama:** GitHub API'sinde belirtilen Proje Kimliğine göre bir depoyu getirir.

3. **Bir Depodaki Tüm Commitleri Al:**
    - **Nokta:** `/api/github/commits?projectOwner=<projectOwner>&repoOwner=<repoOwner>`
    - **Açıklama:** Proje sahibi ve depo sahibi tarafından belirtilen bir depodaki tüm commitleri alır.

4. **Bir Depodaki Satır Sayısını (LOC) Hesapla:**
    - **Nokta:** `/api/github/loc?repoUrl=<repoUrl>`
    - **Açıklama:** Depoları indirir ve Satır Sayısını (LOC) hesaplar.

## GitLab Noktaları

GitLab için benzer noktalar `/api/gitlab/` temel yolunu kullanır.

1. **Tüm Depoları Al:**
    - **Nokta:** `/api/gitlab/repos`
    - **Açıklama:** GitLab hesabındaki tüm genel depoları alır.

2. **Proje Kimliğine Göre Bir Depo Al:**
    - **Nokta:** `/api/gitlab/repo?projectID=<projectID>`
    - **Açıklama:** GitLab API'sinde belirtilen Proje Kimliğine göre bir depoyu getirir.

3. **Bir Depodaki Tüm Commitleri Al:**
    - **Nokta:** `/api/gitlab/commits?projectOwner=<projectOwner>&repoOwner=<repoOwner>`
    - **Açıklama:** Proje sahibi ve depo sahibi tarafından belirtilen bir depodaki tüm commitleri alır.

4. **Bir Depodaki Satır Sayısını (LOC) Hesapla:**
    - **Nokta:** `/api/gitlab/loc?repoUrl=<repoUrl>`
    - **Açıklama:** Depoları indirir ve Satır Sayısını (LOC) hesaplar.

Bu noktaları kullanarak GitHub ve GitLab API'leri ile etkileşimde bulunabilirsiniz.

# Docker İçin CLI Kullanım Örnekleri

Docker konteynırı içinde CLI komutlarını nasıl kullanılacağına dair örnekler:

## GitHub Noktaları

### Tüm Depoları Al
```bash
docker run app /bin/sh -c "./app cli --github-token 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"
```

```bash
docker run app /bin/sh -c "./app cli --github-token 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 --project-id 621058402"
```

## GitLab Noktaları
```bash
docker run app /bin/sh -c "./app cli --gitlab-host https://gitlab.youandus.net --gitlab-token glpat-FiBYym_JyJPkhsmxVydv"
```
```bash
docker run app /bin/sh -c "./app cli --gitlab-host https://gitlab.youandus.net --gitlab-token glpat-FiBYym_JyJPkhsmxVydv --project-id 3"
```