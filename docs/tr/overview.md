# Genel Bakış

Statigo, yüksek performanslı, SEO optimize edilmiş, çok dilli web siteleri oluşturmak için tasarlanmış üretime hazır bir Go web framework'üdür.

## Statigo Nedir?

Statigo, üretime çalışan açılış sayfası sistemlerinden kanıtlanmış kalıpları çıkarır ve bunları kullanımı kolay bir framework olarak sağlar. Aşağıdakilere ihtiyaç duyan geliştiriciler için tasarlanmıştır:

- **Hızlı sayfa yüklemeleri** akıllı önbellekleme ile
- **SEO optimizasyonu** kutudan çıktığı gibi
- **Çok dil desteği** SEO uyumlu URL yönlendirme ile
- **Güvenlik** kapsamlı ara yazılım ile
- **Basit dağıtım** tek bir ikili olarak

## Temel Özellikler

### Statik-Öncelikli Mimari
Statigo, sayfaları önceden oluşturur ve bunları akıllıca önbelleğe alır. Bu şu anlama gelir:
- İstek daha yavaş olabilir (önbellek oluşturur)
- Sonraki istekler son derece hızlıdır (önbellekten sunar)
- Önbellek geçersiz kılma web kancaları veya zaman tabanlı stratejiler ile olur

### Çok Dil Yönlendirme
SEO uyumlu URL'ler ile yerleşik çok dil desteği:
```json
{
  "canonical": "/about",
  "paths": {
    "en": "/en/about",
    "tr": "/tr/hakkinda"
  }
}
```

### Önbellekleme Stratejileri
Her rota için doğru önbellekleme stratejisini seçin:
- **immutable** - Asla sona ermez (örn. statik varlıklar)
- **static** - Uzun önbellek, eski olduğunda yeniden doğrular
- **incremental** - 24 saat sonra otomatik yeniden doğrular
- **dynamic** - Her zaman eski olduğunda yeniden doğrular

### Güvenlik Ara Yazılımı
Kapsamlı güvenlik koruması dahildir:
- Token bucket algoritması ile oran sınırlama
- Kalıcı depolama ile IP yasaklama listesi
- Bot algılama için bal kapanağı tuzakları
- Güvenlik başlıkları (CSP, HSTS, X-Frame-Options)
- Yapılandırılmış çıktı ile istek günlüğü

## Mimari

```
┌─────────────┐
│   İstemci   │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────┐
│      Ara Yazılım Boru Hattı             │
│  ─────────────────────────────────────  │
│  • Yapılandırılmış Günlük Kaydı          │
│  • IP Yasak Listesi                      │
│  • Bal Kapanağı Koruması                │
│  • Oran Sınırlama                        │
│  • Sıkıştırma (Brotli/Gzip)             │
│  • Güvenlik Başlıkları                   │
│  • Dil Algılama                         │
│  • Önbellek Arama                       │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│          Rota İşleyici                  │
│  ─────────────────────────────────────  │
│  • Sayfa İşleyici (index, about, vb.)    │
│  • Şablon Oluşturma                     │
│  • Önbellek Depolama                    │
└─────────────────────────────────────────┘
```

## Sonraki Adımlar

- [Başlarken](getting-started) - Temel bilgileri öğrenin
- [Yönlendirme](routing) - Çok dilli rotaları yapılandırın
- [Ara Yazılım](middleware) - Uygulamanıza ara yazılım ekleyin
- [Önbellekleme](caching) - Önbellekleme stratejilerini anlayın
