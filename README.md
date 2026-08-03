# s3intel

**s3intel**, AWS S3 bucket'larındaki güvenlik açıklarını tespit etmek için yazılmış bir keşif ve istihbarat aracıdır. Go diliyle geliştirilmiştir.

## Ne İşe Yarar?

s3intel, açık bırakılmış veya yanlış yapılandırılmış S3 bucket'larındaki hassas dosyaları (API anahtarları, veritabanı yedekleri, özel anahtarlar, kimlik bilgileri vb.) tespit eder ve her bulguya **0–100 arası risk skoru** verir.

Araç iki farklı modda çalışır:

| Mod | Ne Yapar | Neye Bağlanır |
|-----|----------|---------------|
| **Aktif** | Gerçek AWS S3 API çağrısı yaparak bucket içeriğini tarar | Sadece `config/whitelist.yaml` içindeki bucket'lar |
| **Pasif** | [grayhatwarfare](https://buckets.grayhatwarfare.com) API'si üzerinden zaten indekslenmiş verileri arar | Sadece grayhatwarfare API (gerçek bucket'a bağlanmaz) |

Bu iki mod paket seviyesinde fiziksel olarak birbirinden ayrılmıştır. Pasif mod hiçbir zaman AWS SDK kullanmaz, aktif mod hiçbir zaman grayhatwarfare'e istek atmaz.

## Kurulum

**Gereksinimler:** Go 1.21+

```bash
# Bağımlılıkları indir
go mod download

# Derle
go build -o s3intel .
```

Ya da `make` ile:

```bash
make build   # derle
make test    # testleri çalıştır
make web     # derle + web arayüzünü başlat
```

## Nasıl Çalıştırılır?

### 1. Web Arayüzü (En Kolay Yol)

```bash
./s3intel serve
# ya da: make web
```

Tarayıcında **http://127.0.0.1:8080** adresini aç. Açılan sayfada:

1. **Mod** seç (Pasif veya Aktif)
2. **Anahtar kelime** yaz (pasif modda) veya **bucket adı** yaz (aktif modda)
3. Pasif moddaysan **GHW API Key** alanına API anahtarını gir
4. **Tara** butonuna bas

Sonuçlar renkli bir tablo olarak sayfada görünür. Sunucu sadece `127.0.0.1`'e bağlanır, dışarıdan erişilemez. Kapatmak için terminalde `Ctrl+C`.

### 2. Komut Satırı (CLI)

```bash
# PASİF: grayhatwarfare'de "backup" anahtar kelimesini ara
export GHW_API_KEY="senin-api-anahtarın"
./s3intel passive --keyword backup

# PASİF: sonucu JSON dosyasına kaydet
./s3intel passive --keyword .env --output json --output-file sonuclar.json

# PASİF: sonucu CSV dosyasına kaydet
./s3intel passive --keyword sql --output csv --output-file bulgular.csv

# AKTİF: whitelist'teki bir bucket'ı tara
./s3intel active --bucket test-lab-level1

# AKTİF: sonucu JSON dosyasına kaydet
./s3intel active --bucket test-lab-level1 --output json --output-file sonuclar.json
```

### Çıktı Formatları

`--output` parametresiyle çıktı formatını seçebilirsin:

| Format | Açıklama |
|--------|----------|
| `terminal` | Renkli tablo (varsayılan) |
| `json` | JSON formatı |
| `csv` | CSV formatı |

`--output-file dosya_adı` ile çıktıyı dosyaya kaydedebilirsin. Belirtilmezse ekrana yazdırılır.

## GHW API Key Ayarlama

Pasif mod, [grayhatwarfare](https://buckets.grayhatwarfare.com) API'sini kullanır. API anahtarını iki şekilde verebilirsin:

**1. Ortam değişkeni olarak (CLI için):**
```bash
export GHW_API_KEY="senin-api-anahtarın"
```

**2. Web arayüzünden (tarayıcı için):**

Web panelinde pasif mod seçildiğinde ekranda çıkan **GHW API Key** alanına anahtarını yaz.

> **Not:** API key verilmezse araç mock (sahte) örnek veriyle çalışır. Bu sayede gerçek bir API anahtarın olmadan da aracı deneyebilirsin.

## Whitelist Ayarlama (Aktif Mod)

Aktif modda taranabilecek bucket'lar `config/whitelist.yaml` dosyasında tanımlanır:

```yaml
allowed_buckets:
  - "test-lab-level1"
  - "kadir-s3intel-test-bucket"
```

Buraya **sadece kendi kontrolündeki veya tarama izni olan** bucket adlarını ekle. Listede olmayan bir bucket'ı taramaya çalışırsan, araç AWS'ye hiçbir istek göndermeden hata verir ve durur.

## Risk Skorlama

Araç her dosyayı iki katmanda değerlendirir:

1. **Dosya sınıflandırması** — Dosya adı ve uzantısına göre temel risk puanı atanır:

| Kategori | Örnekler | Temel Risk |
|----------|----------|------------|
| Özel Anahtar | `.pem`, `.key`, `id_rsa` | 85 |
| Kimlik Bilgisi | `.aws/credentials` | 80 |
| Ortam Değişkeni | `.env` | 70 |
| Veritabanı | `.sql`, `.db`, `.sqlite` | 60 |
| Yedek | `.bak`, `.old` | 55 |
| Arşiv | `.zip`, `.tar.gz` | 30 |
| Diğer | — | 10 |

2. **Secret tarama** (sadece aktif mod) — Dosya içeriğinde regex ve Shannon entropy ile gizli anahtar/token aranır, bulunan her secret risk skorunu artırır.

Nihai skor **0–100** aralığında olur; kırmızı (≥70), turuncu (≥40), yeşil (<40) olarak renklendirilir.

## Test

```bash
go test ./...
```

Testler şunları doğrular:

- **secretscan:** `testdata/sample_secrets.txt` içindeki bilinen desenlerle secret tarama doğruluğu
- **activeprobe:** Whitelist'te olmayan bucket'lar için hiçbir AWS çağrısı yapılmadığı
- **passiveintel:** Mock ve HTTP kaynaklarının doğru çalıştığı
- **webui:** Web arayüzünün aynı whitelist ve skorlama kurallarını uyguladığı

## Proje Yapısı

```
s3intel/
├── cmd/                    # CLI komutları (active, passive, serve)
├── config/                 # Whitelist ayarları
├── internal/
│   ├── activeprobe/        # Aktif mod: gerçek AWS S3 tarama
│   ├── passiveintel/       # Pasif mod: grayhatwarfare API sorgusu
│   ├── classifier/         # Dosya adı/uzantı bazlı sınıflandırma
│   ├── secretscan/         # Regex + entropy ile secret tarama
│   ├── riskengine/         # Risk skorlama motoru
│   ├── reporter/           # Çıktı formatları (terminal, JSON, CSV)
│   ├── scanjobs/           # Aktif/pasif orkestrasyon katmanı
│   └── webui/              # Web arayüzü sunucusu
├── testdata/               # Test verileri
├── docs/                   # Mimari belgeler
├── main.go                 # Giriş noktası
├── Makefile                # Derleme ve çalıştırma komutları
└── README.md
```

Mimari detaylar ve Mermaid diyagramı için [docs/architecture.md](docs/architecture.md) dosyasına bakın.
