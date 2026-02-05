# Ara Yazılım (Middleware)

Statigo, güvenlik, performans ve işlevsellik için kapsamlı bir ara yazılım boru hattı içerir.

## Mevcut Ara Yazılımlar

### Günlük Kaydı Ara Yazılımı

slog ile yapılandırılmış istek günlüğü kaydı:

```go
r.Use(middleware.StructuredLogger(logger))
```

Çıktı:
```
INFO request method=GET path=/en/about status=200 duration=5ms
```

### Sıkıştırma Ara Yazılımı

Brotli (tercih edilir) ve gzip sıkıştırma:

```go
r.Use(middleware.Compression())
```

Otomatik olarak sıkıştırır: HTML, CSS, JS, JSON, XML, SVG

### Oran Sınırlama Ara Yazılımı

Token bucket oran sınırlama:

```go
r.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
    RPS:   10,  // Saniye başına istek
    Burst: 20,  // Patlama boyutu
}))
```

Ortam değişkeni üzerinden yapılandırın:
```bash
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20
```

### IP Yasaklama Ara Yazılımı

Kalıcı depolama ile yasaklı IP'leri engelleyin:

```go
ipBanList, _ := security.NewIPBanList("data/banned-ips.json", logger)
r.Use(middleware.IPBanMiddleware(ipBanList, logger))
```

Programatik olarak bir IP'yi yasaklayın:
```go
ipBanList.Ban("192.168.1.100", "Kötüye kullanım", r)
```

### Bal Kapanağı Ara Yazılımı

Sahte admin yollarına erişen botları yakalayın:

```go
honeypotPaths := []string{
    "/admin", "/wp-admin", "/wp-login.php",
    "/.env", "/.git/config",
}
r.Use(middleware.HoneypotMiddleware(ipBanList, honeypotPaths, logger))
```

Bu yollara erişen botlar otomatik olarak yasaklanır.

### Güvenlik Başlıkları Ara Yazılımı

Güvenlik başlıkları ekleyin:

```go
// Basit ön ayar
r.Use(middleware.SecurityHeadersSimple())

// Veya özelleştirin
r.Use(middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
    CSP:           "default-src 'self'",
    HSTSEnabled:   true,
    HSTSMaxAge:    31536000,
    FrameOptions:  "DENY",
    PermissionsPolicy: "geolocation=(), camera=()",
}))
```

Eklenen başlıklar:
- `Content-Security-Policy`
- `X-Frame-Options`
- `X-Content-Type-Options`
- `Strict-Transport-Security`
- `Permissions-Policy`

### Dil Ara Yazılımı

URL, çerez veya Accept-Language başlığından dil algılar ve ayarlar:

```go
langConfig := middleware.LanguageConfig{
    SupportedLanguages: []string{"en", "tr"},
    DefaultLanguage:    "en",
    SkipPaths:          []string{"/robots.txt", "/sitemap.xml"},
    SkipPrefixes:       []string{"/static/", "/health/"},
}
r.Use(middleware.Language(i18nInstance, langConfig))
```

Algılama öncelik sırası:
1. URL yolu öneki (`/en/`, `/tr/`)
2. Çerez (`lang`)
3. `Accept-Language` başlığı
4. Varsayılan dil

### Önbellek Başlıkları Ara Yazılımı

Tarayıcı önbellek başlıkları ekleyin:

```go
r.Use(middleware.CachingHeaders(devMode))
```

Önbellek davranışı:
- **Geliştirme modu**: `no-cache`
- **Üretim**: Rota stratejisine dayalı

### Kanonik Yol Ara Yazılımı

Kanonik yolları saklayın ve doğrulayın:

```go
r.Use(router.CanonicalPathMiddleware(routeRegistry))
```

### Önbellek Ara Yazılımı

Otomatik geçersiz kılma ile yanıt önbellekleme:

```go
r.Use(middleware.CacheMiddleware(cacheManager, logger))
```

### Webhook Yetkilendirme Ara Yazılımı

Webhook isteklerini doğrulayın:

```go
r.Use(middleware.WebhookAuthMiddleware("my-secret-key"))
```

## Ara Yazılım Sırası

Önerilen sıra önemlidir:

```go
r.Use(middleware.StructuredLogger(logger))           // 1. Her şeyi günlüğe kaydet
r.Use(chiMiddleware.Recoverer)                        // 2. Panic kurtarma
r.Use(middleware.IPBanMiddleware(ipBanList, logger))  // 3. Yasaklı IP'leri engelle
r.Use(middleware.HoneypotMiddleware(ipBanList, paths, logger)) // 4. Botları yakala
r.Use(middleware.RateLimiter(config))                 // 5. Oran sınırla
r.Use(middleware.Compression())                       // 6. Yanıtları sıkıştır
r.Use(middleware.SecurityHeadersSimple())             // 7. Güvenlik başlıkları
r.Use(middleware.CachingHeaders(devMode))             // 8. Önbellek başlıkları
r.Use(middleware.Language(i18nInstance, config))      // 9. Dil algılama
r.Use(router.CanonicalPathMiddleware(routeRegistry))  // 10. Yolları doğrula
r.Use(middleware.CacheMiddleware(cacheManager, logger)) // 11. Yanıt önbelleği
```

## Özel Ara Yazılım

Kendi ara yazılımınızı oluşturun:

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // İstek öncesi
        start := time.Now()

        // Sonraki işleyiciyi çağır
        next.ServeHTTP(w, r)

        // İstek sonrası
        duration := time.Since(start)
        log.Println("İstek süresi", duration)
    })
}
```

Kullanın:
```go
r.Use(MyMiddleware)
```

## Chi Ara Yazılımları

Statigo chi üzerine kuruludur. Herhangi bir chi ara yazılımını kullanabilirsiniz:

```go
import chiMiddleware "github.com/go-chi/chi/middleware"

r.Use(chiMiddleware.RequestID)
r.Use(chiMiddleware.RealIP)
r.Use(chiMiddleware.Logger)
r.Use(chiMiddleware.Recoverer)
```

Daha fazla bilgi için [chi belgelerine](https://github.com/go-chi/chi) bakın.
