package main

import (
	"embed"
	"io/fs"
)

// Embed all static assets, templates, translations, and config files
//
//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed translations
var translationsFS embed.FS

//go:embed config
var configFS embed.FS

// GetTemplatesFS returns the embedded templates filesystem
func GetTemplatesFS() fs.FS {
	// Since we embed ../templates, the path in the embed.FS is "templates"
	sub, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		panic("failed to get templates sub-filesystem: " + err.Error())
	}
	return sub
}

// GetStaticFS returns the embedded static assets filesystem
func GetStaticFS() fs.FS {
	// Since we embed ../static, the path in the embed.FS is "static"
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("failed to get static sub-filesystem: " + err.Error())
	}
	return sub
}

// GetTranslationsFS returns the embedded translations filesystem
func GetTranslationsFS() fs.FS {
	// Since we embed ../translations, the path in the embed.FS is "translations"
	sub, err := fs.Sub(translationsFS, "translations")
	if err != nil {
		panic("failed to get translations sub-filesystem: " + err.Error())
	}
	return sub
}

// GetConfigFS returns the embedded config filesystem
func GetConfigFS() fs.FS {
	// Since we embed ../config, the path in the embed.FS is "config"
	sub, err := fs.Sub(configFS, "config")
	if err != nil {
		panic("failed to get config sub-filesystem: " + err.Error())
	}
	return sub
}
