# Başlarken

Bu kılavuz, sıfırdan bir Statigo projesi kurmanıza yardımcı olacaktır.

## Önkoşullar

- Go 1.25 veya üzeri
- Go ve HTTP hakkında temel bilgi

## Kurulum

### Yeni Proje Oluştur

```bash
mkdir my-project
cd my-project
go mod init my-project
```

### Statigo Bağımlılığını Ekleyin

Statigo'yu bir modül olarak kullanıyorsanız:

```bash
go get github.com/yourusername/statigo
```

Veya framework dosyalarını doğrudan projenize kopyalayın.

## Temel Kurulum

### 1. Dizin Yapısını Oluştur

```bash
mkdir -p templates/{layouts,pages,partials}
mkdir -p static/{styles,scripts}
mkdir -p translations
mkdir -p config
```

### 2. Basit Bir Rota Oluştur

`config/routes.json` oluşturun:

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
      "title": "Home"
    }
  ]
}
```

### 3. Temel Şablon Oluştur

`templates/layouts/base.html` oluşturun:

```html
<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{if .Title}}{{.Title}}{{else}}Welcome{{end}}</title>
</head>
<body>
    {{block "main" .}}{{end}}
</body>
</html>
```

### 4. Sayfa Şablonu Oluştur

`templates/pages/index.html` oluşturun:

```html
{{define "title"}}Home{{end}}
{{template "base" .}}

{{define "main"}}
<h1>Statigo'ya Hoş Geldiniz!</h1>
<p>Bu ilk sayfanız.</p>
{{end}}
```

## Uygulamayı Çalıştırma

```bash
go run .
```

http://localhost:8080/en adresini ziyaret edin.

## Sonraki Adımlar

- [Yönlendirme](routing) - Çok dilli yönlendirme hakkında bilgi edinin
- [Ara Yazılım](middleware) - Güvenlik ve performans özellikleri ekleyin
