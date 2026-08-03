# s3intel

`s3intel`, AWS S3 keşfi ve istihbaratı için yazılmış, Go dilinde bir
komut satırı aracıdır. Bir dosyayı sınıflandırır (isim/uzantı), içeriğinde
regex + Shannon entropy ile olası sızmış secret/token arar ve bulguları
0-100 arası bir risk skoruyla önceliklendirilmiş biçimde raporlar.

**Aktif mod SADECE `config/whitelist.yaml` içindeki kendi test
bucket'larına karşı çalışır. Pasif mod hiçbir zaman gerçek bir bucket'a
bağlanmaz** — sadece grayhatwarfare üzerinden zaten indekslenmiş veriyi
okur. Bu ayrım sadece işlevsel değil, paket seviyesinde de mimari olarak
ayrılmıştır; ayrıntı ve kanıt için [docs/architecture.md](docs/architecture.md)
dosyasına bakın.

## Kurulum

```bash
go mod download
go build -o s3intel .
```

Ya da `make` ile tek adımda:

```bash
make build   # sadece derler
make test    # go test ./...
make web     # derler + tarayıcı arayüzünü başlatır (http://127.0.0.1:8080)
```

## Tarayıcıda çalıştırma (en kolay yol)

```bash
make web
# ya da: ./s3intel serve --port 8080
```

Tarayıcında **http://127.0.0.1:8080** adresini aç. Açılan sayfada mod
(Aktif/Pasif) seç, bucket adı ya da anahtar kelimeyi yaz, "Tara" butonuna
bas — sonuçlar renkli bir tablo olarak aynı sayfada görünür. Flag
ezberlemene gerek yok. Sunucu SADECE `127.0.0.1`'e bağlanır, dışarıdan
erişilemez; kapatmak için terminalde `Ctrl+C`.

Arka planda web arayüzü de tam olarak aynı `internal/scanjobs` katmanını,
dolayısıyla aynı whitelist kuralını ve aynı aktif/pasif ayrımını kullanır —
CLI'den farkı sadece görselleştirmedir.

## Komut satırından çalıştırma

```bash
# AKTIF: whitelist'teki bir bucket'ı tara, terminale renkli tablo bas
./s3intel active --bucket test-lab-level1 --output terminal

# AKTIF: sonucu JSON dosyasına yaz
./s3intel active --bucket test-lab-level1 --output json --output-file results.json

# PASİF: grayhatwarfare'de "backup" anahtar kelimesini ara
./s3intel passive --keyword backup --output terminal

# PASİF: sonucu CSV dosyasına yaz
./s3intel passive --keyword .env --output csv --output-file findings.csv
```

Global flag'ler (`--output`, `--output-file`, `--verbose`) hem `active` hem
`passive` komutlarında geçerlidir.

## Whitelist nasıl güncellenir

Aktif modda taranabilecek bucket'lar `config/whitelist.yaml` dosyasında
listelenir:

```yaml
allowed_buckets:
  - "test-lab-level1"
  - "kadir-s3intel-test-bucket"
```

Buraya **sadece kendi kontrolündeki / test etmek için izinli olduğun**
bucket adlarını ekle. Listede olmayan bir bucket adıyla `active` komutu
çalıştırılırsa, `internal/activeprobe/whitelist.go` bunu daha AWS'ye hiçbir
istek gitmeden reddeder ve hata ile durur.

## grayhatwarfare API key nasıl ayarlanır

```bash
export GHW_API_KEY="gerçek-api-anahtarın"
```

`GHW_API_KEY` ayarlı değilse `passive` komutu otomatik olarak mock (sahte,
sabit) örnek veriyle çalışır — yani gerçek bir API key olmadan da aracı
uçtan uca deneyebilirsin.

## Test

```bash
go test ./...
```

`internal/secretscan` testleri, `testdata/sample_secrets.txt` içindeki
bilinen (sahte) desenlerle aracın ground-truth doğruluğunu doğrular.
`internal/activeprobe` testleri, whitelist'te olmayan bir bucket için
gerçek/mock hiçbir AWS S3 çağrısının tetiklenmediğini kanıtlar.
`internal/webui` testleri, web arayüzünün aynı whitelist reddini ve aynı
skorlama sonuçlarını doğru render ettiğini kanıtlar.

## Mimari

Aktif/pasif ayrımının Mermaid diyagramı ve paket seviyesindeki fiziksel
ayrımın kanıtı için [docs/architecture.md](docs/architecture.md) dosyasına
bakın.
