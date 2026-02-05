# Şablonlar

Statigo, Go'nun `html/template` paketini özel işlevler ve gömülü dosya sistemleri ile hızlı, güvenli HTML oluşturma için kullanır.

## Şablon Yapısı

```
templates/
├── layouts/
│   └── base.html          # Temel düzen
├── pages/
│   ├── index.html         # Ana sayfa
│   ├── about.html         # Hakkında sayfası
│   └── docs.html          # Belgeler sayfası
└── partials/
    ├── header.html        # Üstbilgi bileşeni
    └── footer.html        # Altbilgi bileşeni
```

## Temel Düzen

`templates/layouts/base.html` oluşturun:

```html
<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{if .Title}}{{.Title}} - {{end}}{{t "site.name"}}</title>

    <!-- SEO -->
    <link rel="canonical" href="{{canonicalURL .Canonical .Lang}}">
    {{alternateLinks .Canonical}}

    <!-- Stiller -->
    <link rel="stylesheet" href="/static/styles/main.css">
</head>
<body>
    {{template "header" .}}

    <main>
        {{block "main" .}}{{end}}
    </main>

    {{template "footer" .}}
</body>
</html>
```

## Sayfa Şablonları

Her sayfa temel düzeni genişletir:

`templates/pages/index.html`:
```html
{{define "title"}}Ana Sayfa{{end}}
{{template "base" .}}

{{define "main"}}
<div class="hero">
    <h1>{{t "pages.home.title"}}</h1>
    <p>{{t "pages.home.subtitle"}}</p>
</div>
{{end}}
```

## Parçalar

`templates/partials/` içinde yeniden kullanılabilir bileşenler:

`templates/partials/header.html`:
```html
{{define "header"}}
<header class="site-header">
    <nav>
        <a href="{{localePath "/" .Lang}}">{{t "site.name"}}</a>
        <a href="{{localePath "/about" .Lang}}">{{t "nav.about"}}</a>
        <a href="{{localePath "/docs" .Lang}}">{{t "nav.docs"}}</a>
    </nav>

    <div class="lang-switcher">
        <a href="{{localePath .Canonical "en"}}"{{if eq .Lang "en"}} class="active"{{end}}>EN</a>
        <a href="{{localePath .Canonical "tr"}}"{{if eq .Lang "tr"}} class="active"{{end}}>TR</a>
    </div>
</header>
{{end}}
```

`templates/partials/footer.html`:
```html
{{define "footer"}}
<footer class="site-footer">
    <p>&copy; {{t "site.copyright"}}</p>
</footer>
{{end}}
```

## Şablon İşlevleri

### SEO İşlevleri

```html
<!-- Kanonik URL -->
<link rel="canonical" href="{{canonicalURL "/about" .Lang}}">

<!-- Alternatif bağlantılar (hreflang) -->
{{alternateLinks "/about"}}

<!-- Yerel duyarlı yol -->
<a href="{{localePath "/about" .Lang}}">Hakkında</a>
```

### Çeviri İşlevi

```html
{{t "pages.home.title"}}
{{t "nav.home"}}
{{t "errors.not_found"}}
```

### Matematik İşlevleri

```html
{{add 1 2}}        <!-- 3 -->
{{sub 10 3}}       <!-- 7 -->
{{div 100 5}}      <!-- 20 -->
{{mod 10 3}}       <!-- 1 -->
```

### Dize İşlevleri

```html
{{slugify "Merhaba Dünya"}}    <!-- merhaba-dunya -->
{{safeHTML "<p>Ham HTML</p>"}}  <!-- HTML olarak oluşturur -->
{{safeURL "http://ornek.com"}} <!-- URL için kaçar -->
```

### Tarih İşlevleri

```html
{{formatDate .Date "2006-01-02"}}
{{formatDateTime .Date "2006-01-02T15:04:05Z"}}
```

### Yardımcı İşlevler

```html
{{dict "key" "value" "key2" "value2"}}
{{set . "NewField" "value"}}
{{until 5}}  <!-- [0,1,2,3,4] dilimini döndürür -->
```

## Şablonlara Veri Geçirme

İşleyicinizde:

```go
data := map[string]interface{}{
    "Title":    "Sayfa Başlığı",
    "Lang":     "tr",
    "Canonical": "/about",
    "User": map[string]string{
        "Name":  "Ahmet",
        "Email": "ahmet@ornek.com",
    },
    "Posts": []Post{
        {Title: "İlk Gönderi", Content: "..."},
        {Title: "İkinci Gönderi", Content: "..."},
    },
}
renderer.Render(w, "about.html", data)
```

Şablonunuzda:

```html
<h1>{{.Title}}</h1>
<p>Hoş geldin, {{.User.Name}}</p>

{{range .Posts}}
    <article>
        <h2>{{.Title}}</h2>
        <p>{{.Content}}</p>
    </article>
{{end}}
```

## Bağlam Değişkenleri

Statigo otomatik olarak şu bağlam değişkenlerini sağlar:

| Değişken | Açıklama |
|----------|-------------|
| `{{.Lang}}` | Mevcut dil kodu |
| `{{.Canonical}}` | Kanonik yol |
| `{{.Title}}` | Sayfa başlığı |
| `{{.Env}}` | Ortam değişkenleri (örn. `{{.Env.GTM_ID}}`) |

## Koşullu Oluşturma

```html
{{if .User}}
    <p>Hoş geldin, {{.User.Name}}</p>
{{else}}
    <p>Lütfen giriş yapın</p>
{{end}}

{{if eq .Lang "tr"}}
    <p>Türkçe içerik</p>
{{else if eq .Lang "en"}}
    <p>İngilizce içerik</p>
{{end}}
```

## Döngüler

```html
{{range $index, $item := .Items}}
    <div class="item-{{$index}}">
        {{$item.Name}}
    </div>
{{end}}
```

## Güvenli HTML

Ham HTML oluşturmak için (dikkatli kullanın):

```go
data := map[string]interface{}{
    "Content": "<p><strong>Kalın metin</strong></p>",
}
```

```html
{{safeHTML .Content}}
<!-- Oluşturur: <p><strong>Kalın metin</strong></p> -->
```

## Şablon Mirası

Varsayılan içerik için `block` kullanın (geçersiz kılınabilir):

`base.html` içinde:
```html
{{block "main" .}}
    <p>Varsayılan içerik</p>
{{end}}
```

`index.html` içinde:
```html
{{define "main"}}
    <p>Index sayfası için özel içerik</p>
{{end}}
```

## Ortam Değişkenleri

Şablonlarda ortam değişkenlerine erişin:

```html
{{if .Env.GTM_ID}}
    <!-- Google Etiket Yöneticisi -->
    <script>(function(w,d,s,l,i){...})</script>
{{end}}
```

`.env` içinde ayarlayın:
```
GTM_ID=GTM-XXXXX
```

## En İyi Uygulamalar

1. **Her zaman kullanıcı girdisini kaçırın** - `html/template` otomatik kaçış kullan
2. **`safeHTML`'i dikkatli kullanın** - Sadece güvenilen içerik için
3. **Düzenleri basit tutun** - Çok derine iç içe girmeyin
4. **Tekrar için parçalar kullanın** - Üstbilgi, altbilgi, kartlar
5. **Çevirileri gruplandırın** - i18n anahtarlarında nokta gösterimi kullanın
6. **Kanonik URL'leri doğrulayın** - Doğru SEO yapısı için
