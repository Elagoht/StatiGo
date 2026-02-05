# Önbellekleme

Statigo, optimal performans için iki katmanlı önbellekleme sistemi sunar: bellek içi önbellekleme ve disk kalıcılığı.

## Mimari

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   İstek     │────▶│ Bellek Önbellek│────▶│Disk Önbellek│
└─────────────┘     └─────────────┘     └─────────────┘
                          │
                          ▼
                     ┌─────────────┐
                     │  Brotli     │
                     │ Sıkıştırılmış│
                     └─────────────┘
```

## Önbellek Stratejileri

Her rota farklı bir önbellekleme stratejisine sahip olabilir:

| Strateji | Açıklama | Kullanım Durumu |
|----------|-------------|----------|
| `immutable` | Asla sona ermez | Statik varlıklar, sürümlü dosyalar |
| `static` | Uzun önbellek, eski işaretlendiğinde yeniden doğrular | Nadiren değişen sayfalar |
| `incremental` | 24 saat sonra otomatik yeniden doğrular | Blog gönderileri, makaleler |
| `dynamic` | Her zaman eski olduğunda yeniden doğrular | Kullanıcıya özgü içerik |

`config/routes.json` içinde tanımlayın:

```json
{
  "canonical": "/",
  "paths": {"en": "/en"},
  "strategy": "static",
  "template": "index.html",
  "handler": "index"
}
```

## Başlatma

```go
import "statigo/framework/cache"

cacheManager, err := cache.NewManager("data/cache", logger)
if err != nil {
    log.Fatal(err)
}
```

## Önbellek Ara Yazılımı

```go
r.Use(middleware.CacheMiddleware(cacheManager, logger))
```

Ara yazılım otomatik olarak:
1. Mevcut yanıt için bellek önbelleğini kontrol eder
2. Bellek isabeti yoksa disk önbelleğini kontrol eder
3. Her ikisi de isabet yoksa işleyiciyi çalıştırır
4. Yanıtı her iki katmana da depolar

## Önceden Oluşturma

Tüm sayfaları başlangıçta oluşturun:

```bash
# Uygulamayı derle
go build -o statigo

# Tüm sayfaları önceden oluştur
./statigo prerender
```

Veya programatik olarak:

```go
cacheManager.RebuildAll(r, appLogger)
```

## Önbellek Geçersiz Kılma

### Manuel Geçersiz Kılma

Bir rotayı eski olarak işaretleyin:

```go
cacheManager.MarkStale("/en/about")
```

### Webhook Geçersiz Kılma

Webhook uç noktasını yapılandırın:

```go
r.Post("/cache/webhook", middleware.WebhookInvalidate(
    cacheManager,
    os.Getenv("WEBHOOK_SECRET"), // Ortam değişkeninden
    logger,
))
```

Webhook gönderin:

```bash
curl -X POST http://localhost:8080/cache/webhook \
  -H "X-Webhook-Secret: your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"canonical": "/about"}'
```

### Strateji Tabanlı Geçersiz Kılma

Stratejiye göre yeniden oluşturun:

```go
// Tüm statik sayfaları yeniden oluştur
cacheManager.RebuildByStrategy("static", r, logger)

// Belirli bir rotayı yeniden oluştur
cacheManager.RebuildByCanonical("/about", r, logger)
```

## Önbellek Depolama

### Bellek Önbelleği

- Eşzamanlı erişim için `sync.Map` içinde depolanır
- Brotli ile sıkıştırılır
- Otomatik ETag oluşturma

### Disk Önbelleği

- `data/cache/` dizininde depolanır
- Kanonik yolun SHA256 hash'i ile adlandırılır
- Uygulama yeniden başlatmalarından sonra hayatta kalır

## ETag Desteği

Statigo önbellek girdileri için otomatik olarak ETag'ler oluşturur:

```
ETag: "a1b2c3d4e5f6..."
```

`If-None-Match` başlığına sahip istemciler `304 Not Modified` yanıtları alır.

## Yapılandırma

Ortam değişkenleri:

```bash
# Önbellek dizini
CACHE_DIR=./data/cache

# Önbellekleme devre dışı (test için)
DISABLE_CACHE=false
```

## İzleme

Önbellek sağlığını kontrol edin:

```go
stats := cacheManager.GetStats()
fmt.Printf("Bellek girdileri: %d\n", stats.MemoryEntries)
fmt.Printf("Disk girdileri: %d\n", stats.DiskEntries)
```

## En İyi Uygulamalar

1. **Gerçekten statik içerik için `immutable` kullanın**
   - Sürüm hash'li varlıklar: `/style.v1.css`
   - Belge sayfaları

2. **Nadiren değişen sayfalar için `static` kullanın**
   - Ana sayfa
   - Hakkında sayfaları
   - Özellik sayfaları

3. **İçerik sayfaları için `incremental` kullanın**
   - Blog gönderileri
   - Makaleler
   - Haber öğeleri

4. **Kişiselleştirilmiş içerik için `dynamic` kullanın**
   - Kullanıcı panelleri
   - Admin panelleri
   - Hesap ayarları

5. **Dağıtımdan sonra önceden oluşturun**
   ```bash
   go build -o app
   ./app prerender
   ./app serve
   ```

6. **Webhook geçersiz kılma ayarlayın**
   - CMS entegrasyonu için
   - İçerik güncellemeleri için
   - Otomatik dağıtımlar için
