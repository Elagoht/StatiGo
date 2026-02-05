# CLI (Komut Satırı Arayüzü)

Statigo, önceden oluşturma ve önbellek yönetimi gibi yaygın işlemler için bir CLI framework içerir.

## Yerleşik Komutlar

### prerender

Önbelleği ısıtmak için tüm sayfaları önceden oluşturun:

```bash
./statigo prerender
```

Bu komut:
1. Yapılandırmadan tüm rotaları yükler
2. Her rotaya istek yapar
3. Yanıtları önbelleğe depolar
4. Her sayfa için başarı/başarısızlık raporlar

Kullanım için:
- İlk dağıtım
- Yeniden başlatmadan sonra önbellek ısıtma
- Tüm sayfaların önbelleğe alındığından emin olma

### clear-cache

Tüm önbelleğe alınan sayfaları temizleyin:

```bash
./statigo clear-cache
```

Bu komut:
1. Bellek önbelleğindeki tüm girdileri siler
2. Disk önbelleğindeki tüm dosyaları siler
3. Temizlenen girdi sayısını raporlar

Kullanım için:
- Önbellek yenilemeye zorlama
- Disk alanını boşaltma
- Önbellek sorunlarını giderme

## CLI Kullanma

### Komutları Kaydetme

`main.go` içinde:

```go
import "statigo/framework/cli"

func main() {
    // Bileşenlerinizi başlat...

    // CLI örneği oluştur
    cliInstance := cli.New()

    // Yerleşik komutları kaydet
    cli.RegisterPrerenderCommand(cliInstance, r, cacheManager, logger)
    cli.RegisterClearCacheCommand(cliInstance, cacheManager, logger)

    // CLI komutu sağlanıp sağlanmadığını kontrol et
    if len(os.Args) > 1 {
        if err := cliInstance.Run(os.Args[1], os.Args[2:]); err == nil {
            // Komut başarıyla yürütüldü
            return
        }
    }

    // CLI komutu yoksa sunucuyu başlat
    // ... sunucu kodunuz ...
}
```

### Özel Komutlar Oluşturma

```go
import "statigo/framework/cli"

// Komut işleyicinizi tanımlayın
func MyCommandHandler(args []string, appCLI *cli.CLI) error {
    fmt.Println("Özel komutum çalışıyor")
    fmt.Println("Argümanlar:", args)
    return nil
}

// Komutu kaydedin
cliInstance.Register("my-command", cli.Command{
    Handler:    MyCommandHandler,
    Usage:      "my-command [args...]",
    Summary:    "Komutumun açıklaması",
    Aliases:    []string{"mc", "my-c"},
})
```

### Bağımlılıkları Olan Komut

Uygulama bileşenlerinizi komut işleyicilerine geçirin:

```go
func ExportDataCommand(args []string, appCLI *cli.CLI) error {
    // appCLI.Context'ten bağımlılıklara erişin
    db := appCLI.Context["db"].(*Database)
    cache := appCLI.Context["cache"].(*cache.Manager)

    // Komut mantığınız...
    return nil
}

// Kaydetmeden önce bağlamı ayarlayın
cliInstance.Context["db"] = database
cliInstance.Context["cache"] = cacheManager

cliInstance.Register("export", cli.Command{
    Handler: ExportDataCommand,
    Usage:   "export [format]",
    Summary: "Verileri çeşitli formatlarda dışa aktar",
})
```

## Komut Takma Adları

Komutlar birden fazla ada (takma ad) sahip olabilir:

```go
cliInstance.Register("version", cli.Command{
    Handler: VersionHandler,
    Usage:   "version",
    Summary: "Sürüm bilgilerini göster",
    Aliases: []string{"v", "ver", "--version"},
})
```

Tümü çalışır:
```bash
./statigo version
./statigo v
./statigo ver
./statigo --version
```

## CLI Bağlamı

Sunucu başlatma ve CLI komutları arasında veri paylaşın:

```go
func main() {
    // Bileşenleri başlat
    cacheManager := cache.NewManager(...)
    database := db.Connect(...)

    // CLI oluştur
    cliInstance := cli.New()

    // Bağlam üzerinden bileşenleri paylaş
    cliInstance.Context["cache"] = cacheManager
    cliInstance.Context["db"] = database

    // Bu bileşenleri kullanan komutları kaydet
    cli.RegisterPrerenderCommand(cliInstance, r, cacheManager, logger)

    // CLI komutunu kontrol et
    if len(os.Args) > 1 {
        command := os.Args[1]
        args := os.Args[2:]

        if err := cliInstance.Run(command, args); err == nil {
            return // Komut yürütüldü
        }
    }

    // Sunucuyu başlat...
}
```

## Örnek: Özel Sağlık Kontrolü Komutu

```go
func HealthCheckCommand(args []string, appCLI *cli.CLI) error {
    logger := appCLI.Context["logger"].(*slog.Logger)

    // Çeşitli bileşenleri kontrol et
    checks := map[string]bool{
        "database": checkDatabase(),
        "cache":    checkCache(),
        "api":      checkAPI(),
    }

    allHealthy := true
    for name, healthy := range checks {
        status := "TAMAM"
        if !healthy {
            status = "BAŞARISIZ"
            allHealthy = false
        }
        fmt.Printf("%s: %s\n", name, status)
    }

    if allHealthy {
        fmt.Println("\nTüm sistemler operasyonel")
        return nil
    }

    return fmt.Errorf("bazı sağlık kontrolleri başarısız")
}

// Kaydet
cliInstance.Register("health", cli.Command{
    Handler: HealthCheckCommand,
    Usage:   "health",
    Summary: "Tüm bileşenlerde sağlık kontrolleri çalıştır",
})
```

## Yardım Metni

Statigo CLI otomatik olarak yardım metni oluşturur:

```bash
./statigo help
```

Çıktı:
```
Mevcut komutlar:
  prerender    Önbelleğe tüm sayfaları önceden oluştur
  clear-cache  Tüm önbelleğe alınan sayfaları temizle
  health       Tüm bileşenlerde sağlık kontrolleri çalıştır
  help         Bu yardım iletisini göster

Bir komut hakkında daha fazla bilgi için "statigo help <komut>" kullanın.
```

## Derleme Süreciyle Entegrasyon

`Makefile`'ınıza ekleyin:

```makefile
.PHONY: build prerender run

build:
	go build -o statigo

prerender: build
	./statigo prerender

run: build
	./statigo

deploy: build prerender
	# Dağıtım komutları...
```

Sonra:
```bash
make prerender  # Derle ve önceden oluştur
make deploy     # Sıcak önbellekle dağıt
```
