// Package i18n provides internationalization support for the Statigo framework.
package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// I18n manages translations for multiple languages.
type I18n struct {
	translations map[string]map[string]interface{}
	defaultLang  string
}

// New creates a new I18n instance by loading translations from the given filesystem.
func New(translationsFS fs.FS, defaultLang string) (*I18n, error) {
	i18n := &I18n{
		translations: make(map[string]map[string]interface{}),
		defaultLang:  defaultLang,
	}

	// Load all JSON files from translations filesystem
	files, err := fs.Glob(translationsFS, "*.json")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		lang := strings.TrimSuffix(path.Base(file), ".json")

		data, err := fs.ReadFile(translationsFS, file)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Parse as nested JSON structure
		var translations map[string]interface{}
		if err := json.Unmarshal(data, &translations); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}

		// Store raw nested translations
		i18n.translations[lang] = translations
	}

	return i18n, nil
}

// GetRaw retrieves raw structured data (arrays, objects) from translations using dot notation.
// Example: GetRaw("en", "features.descriptions") returns []interface{}
func (i *I18n) GetRaw(lang, key string) interface{} {
	// Helper to navigate nested map using dot notation
	getValue := func(data map[string]interface{}, path string) interface{} {
		parts := strings.Split(path, ".")
		var current interface{} = data

		for _, part := range parts {
			if currentMap, ok := current.(map[string]interface{}); ok {
				current = currentMap[part]
			} else {
				return nil
			}
		}
		return current
	}

	// Try requested language
	if trans, ok := i.translations[lang]; ok {
		if value := getValue(trans, key); value != nil {
			return value
		}
	}

	// Fallback to default language
	if trans, ok := i.translations[i.defaultLang]; ok {
		if value := getValue(trans, key); value != nil {
			return value
		}
	}

	// Return nil if not found
	return nil
}

// Get retrieves a string translation for the given key.
// Returns the key itself if translation is not found.
func (i *I18n) Get(lang, key string) string {
	value := i.GetRaw(lang, key)
	if str, ok := value.(string); ok {
		return str
	}
	return key
}

// GetSupportedLanguages returns list of available languages.
func (i *I18n) GetSupportedLanguages() []string {
	langs := make([]string, 0, len(i.translations))
	for lang := range i.translations {
		langs = append(langs, lang)
	}
	return langs
}

// DefaultLanguage returns the default/fallback language.
func (i *I18n) DefaultLanguage() string {
	return i.defaultLang
}
