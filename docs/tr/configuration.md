# Yapılandırma

Statigo, esnek uygulama yapılandırması için ortam değişkenleri ve JSON yapılandırma dosyaları kullanır.

## Ortam Değişkenleri

Proje kök dizinine `.env` dosyası oluşturun:

```bash
# Sunucu Yapılandırması
PORT=8080
BASE_URL=http://localhost:8080

# Günlük Kaydı
LOG_LEVEL=INFO

# Önbellek
CACHE_DIR=./data/cache
DISABLE_CACHE=false

# Oran Sınırlama
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20

# Geliştirme
DEV_MODE=false

# Kapatma Zaman Aşımı
SHUTDOWN_TIMEOUT=30

# Webhook Gizli Anahtarı (önbellek geçersiz kılma için)
WEBHOOK_SECRET=your-webhook-secret-key

# Google Etiket Yöneticisi (isteğe bağlı)
GTM_ID=GTM-XXXXX
```

### Sunucu Yapılandırması

| Değişken | Varsayılan | Açıklama |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP sunucusu portu |
| `BASE_URL` | `http://localhost:8080` | Kanonik bağlantılar için temel URL |
| `DEV_MODE` | `false` | Geliştirme modunu etkinleştir |

### Günlük Kaydı

| Değişken | Varsayılan | Açıklama |
|----------|---------|-------------|
| `LOG_LEVEL` | `INFO` | Günlük seviyesi: `DEBUG`, `INFO`, `WARN`, `ERROR` |

### Önbellek

| Değişken | Varsayılan | Açıklama |
|----------|---------|-------------|
| `CACHE_DIR` | `./data/cache` | Önbellek depolama dizini |
| `DISABLE_CACHE` | `false` | Önbellekleme devre dışı (test için) |

### Oran Sınırlama

| Değişken | Varsayılan | Açıklama |
|----------|---------|-------------|
| `RATE_LIMIT_RPS` | `10` | Saniye başına istek |
| `RATE_LIMIT_BURST` | `20` | Patlama boyutu |

### Güvenlik

| Değişken | Varsayılan | Açıklama |
|----------|---------|-------------|
| `WEBHOOK_SECRET` | - | Webhook kimlik doğrulaması için gizli anahtar |

## Rota Yapılandırması

### routes.json

`config/routes.json` içinde rotaları tanımlayın:

```json
{
  "routes": [
    {
      "canonical": "/",
      "paths": {
        "en": "/en"
      },
      "strategy": "static",
      "template": "index.html",
      "handler": "index",
      "title": "pages.home.title"
    }
  ]
}
```

### Rota Alanları

| Alan | Tür | Gerekli | Açıklama |
|-------|------|----------|-------------|
| `canonical` | string | Evet | İsel kanonik yol |
| `paths` | object | Evet | Dile özgü URL yolları |
| `strategy` | string | Evet | Önbellekleme stratejisi |
| `template` | string | Evet | Oluşturulacak şablon dosyası |
| `handler` | string | Evet | `customHandlers` haritasına kayıtlı işleyici adı |
| `title` | string | Hayır | Sayfa başlığı (i18n anahtarı veya harf olabilir) |

### Önbellekleme Stratejileri

- `immutable` - Asla sona ermez (statik varlıklar)
- `static` - Uzun önbellek, eski işaretlendiğinde yeniden doğrular
- `incremental` - 24 saat sonra otomatik yeniden doğrular
- `dynamic` - Her zaman eski olduğunda yeniden doğrular

## Yönlendirme Yapılandırması

### redirects.json

`config/redirects.json` içinde yönlendirmeleri tanımlayın:

```json
{
  "redirects": [
    {
      "from": "/old-page",
      "to": "/new-page",
      "type": 301
    },
    {
      "from": "/blog/*",
      "to": "/articles/*",
      "type": 301,
      "pattern": true
    }
  ]
}
```

### Yönlendirme Alanları

| Alan | Tür | Açıklama |
|-------|------|-------------|
| `from` | string | Kaynak yol (`*` joker karakterini destekler) |
| `to` | string | Hedef yol (eşleşen kısım için `*` kullanın) |
| `type` | number | HTTP durum kodu (301 veya 302) |
| `pattern` | boolean | Joker karakter eşleştirmeyi etkinleştir |

## Çeviri Yapılandırması

Çeviri dosyaları `translations/` dizininde depolanır:

```
translations/
├── en.json    # İngilizce
├── tr.json    # Türkçe
└── de.json    # Almanca
```

### Çeviri Dosyası Biçimi

```json
{
  "site": {
    "name": "Sitem",
    "description": "Bir açıklama"
  },
  "nav": {
    "home": "Ana Sayfa",
    "about": "Hakkında"
  }
}
```

## Go'da Yapılandırmaya Erişme

### Ortam Değişkenleri

```go
import "statigo/framework/utils"

port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}

// Veya yardımcıyı kullan
rateLimitRPS := utils.GetEnvInt("RATE_LIMIT_RPS", 10)
rateLimitBurst := utils.GetEnvInt("RATE_LIMIT_BURST", 20)
```

### Rotaları Yükleme

```go
import "statigo/framework/router"

routeRegistry := router.NewRegistry([]string{"en", "tr"})

err := router.LoadRoutesFromJSON(
    configFS,
    "routes.json",
    routeRegistry,
    renderer,
    customHandlers,
    logger,
)
```

### Yönlendirmeleri Yükleme

```go
import "statigo/framework/middleware"

r.Use(middleware.RedirectMiddleware(
    configFS,
    "redirects.json",
    logger,
))
```

## Yapılandırma En İyi Uygulamaları

1. **Asla `.env` dosyasını işlemeyin** - `.gitignore`'a ekleyin
2. **`.env.example` kullanın** - Gerekli değişkenler için şablon
3. **Yapılandırmayı doğrulayın** - Başlangıçta gerekli değişkenleri kontrol edin
4. **Makul varsayılanlar sağlayın** - İsteğe bağlı ayarlar için varsayılanlar
5. **Değişkenleri belgeleyin** - Her değişkenin ne yaptığını açıklayın

## Örnek Yapılandırma

### Üretim (.env.production)

```bash
PORT=8080
BASE_URL=https://example.com
LOG_LEVEL=WARN
DEV_MODE=false
RATE_LIMIT_RPS=20
RATE_LIMIT_BURST=40
WEBHOOK_SECRET=prod-secret-key
```

### Geliştirme (.env.development)

```bash
PORT=3000
BASE_URL=http://localhost:3000
LOG_LEVEL=DEBUG
DEV_MODE=true
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200
```

### Test (.env.test)

```bash
PORT=8081
BASE_URL=http://localhost:8081
LOG_LEVEL=ERROR
DEV_MODE=true
DISABLE_CACHE=true
RATE_LIMIT_RPS=1000
```

## Ortam Dosyalarını Yükleme

Statigo `godotenv` kullanarak `.env` dosyalarını yükler:

```go
import "github.com/joho/godotenv"

func main() {
    // .env dosyasını yükle
    if err := godotenv.Load(); err != nil {
        log.Println("Uyarı: .env dosyası bulunamadı, varsayılanlar kullanılıyor")
    }

    // Uygulama kodunuz...
}
```

Ortama özgü dosyalar için:

```go
env := os.Getenv("APP_ENV")
if env == "" {
    env = "development"
}

godotenv.Load(".env." + env)
godotenv.Load() // Varsayılan .env için yedek
```
