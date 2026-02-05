# Yönlendirme

Statigo'nun yönlendirme sistemi, SEO optimizasyonu ile çok dilli URL eşleme sağlar.

## Rota Yapılandırması

Rotalar `config/routes.json` içinde tanımlanır:

```json
{
  "routes": [
    {
      "canonical": "/about",
      "paths": {
        "en": "/en/about",
        "tr": "/tr/hakkinda"
      },
      "strategy": "static",
      "template": "about.html",
      "handler": "about",
      "title": "pages.about.title"
    }
  ]
}
```

### Rota Alanları

| Alan | Tür | Açıklama |
|-------|------|-------------|
| `canonical` | string | İsel kanonik yol (aramalar için kullanılır) |
| `paths` | object | Dile özgü URL yolları |
| `strategy` | string | Önbellekleme stratejisi: `static`, `incremental`, `dynamic`, `immutable` |
| `template` | string | Oluşturulacak şablon dosyası |
| `handler` | string | İşleyici adı (`customHandlers` haritasına kayıtlı) |
| `title` | string | Sayfa başlığı (i18n anahtarı veya harf olabilir) |

## Çok Dilli Yönlendirme

### Dile Özgü Yolları Tanımlama

Her rota farklı dillerde farklı URL'lere sahip olabilir:

```json
{
  "canonical": "/features",
  "paths": {
    "en": "/en/features",
    "tr": "/tr/ozellikler",
    "de": "/de/funktionen"
  }
}
```

### Mevcut Dile Erişim

İşleyicilerinizde:

```go
import "statigo/framework/middleware"

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    lang := middleware.GetLanguage(r.Context())
    // lang "en", "tr", vb.
}
```

Şablonlarınızda:

```html
<p>Mevcut dil: {{.Lang}}</p>
```

## SEO Özellikleri

### Kanonik URL'ler

Statigo otomatik olarak kanonik URL'ler oluşturur:

```html
<link rel="canonical" href="{{canonicalURL "/about" .Lang}}">
```

### Alternatif Bağlantılar (Hreflang)

SEO için hreflang bağlantıları oluşturun:

```html
{{alternateLinks "/about"}}
```

Çıktı:
```html
<link rel="alternate" hreflang="en" href="https://example.com/en/about">
<link rel="alternate" hreflang="tr" href="https://example.com/tr/hakkinda">
<link rel="alternate" hreflang="x-default" href="https://example.com/en/about">
```

### Yerel Duyarlı Bağlantılar

Çevrilmiş URL'ler için `localePath` kullanın:

```html
<a href="{{localePath "/about" .Lang}}">Hakkında</a>
```

Bu otomatik olarak çözümlenir:
- İngilizce için `/en/about`
- Türkçe için `/tr/hakkinda`

## Kanonik Yol Ara Yazılımı

`CanonicalPathMiddleware`, kullanıcıların doğru dile özgü URL'ye yönlendirildiğinden emin olur:

```go
r.Use(router.CanonicalPathMiddleware(routeRegistry))
```

Örnek yönlendirmeler:
- `/about` → `/en/about` (İngilizce kullanıcılar için)
- `/about` → `/tr/hakkinda` (Türkçe kullanıcılar için)

## Programatik Rota Kaydı

```go
import "statigo/framework/router"

// Kayıt defteri oluştur
routeRegistry := router.NewRegistry([]string{"en", "tr"})

// Programatik olarak bir rota tanımla
routeRegistry.Register(router.RouteDefinition{
    Canonical: "/contact",
    Paths: map[string]string{
        "en": "/en/contact",
        "tr": "/tr/iletisim",
    },
    Strategy: "static",
    Template: "contact.html",
    Handler:  "contact",
    Title:    "pages.contact.title",
})

// Rotaları chi yönlendiricisine kaydet
routeRegistry.RegisterRoutes(r, nil)
```

## Dinamik Rotalar

Dinamik rotalar için (örn. blog gönderileri), chi'nin rota parametrelerini kullanın:

```go
r.Get("/{lang}/blog/{slug}", blogHandler.ServeHTTP)
```

İşleyicinizdeki parametrelere erişin:

```go
slug := chi.URLParam(r, "slug")
lang := middleware.GetLanguage(r.Context())
```

## Yönlendirmeler

`config/redirects.json` içinde statik yönlendirmeleri yapılandırın:

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

Yönlendirme ara yazılımını uygulayın:

```go
r.Use(middleware.RedirectMiddleware(configFS, "redirects.json", logger))
```
