// Package utils provides utility functions for the Statigo framework.
package utils

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

// Minifier handles minification of CSS, JS, and HTML.
type Minifier struct {
	m *minify.M
}

// NewMinifier creates a new Minifier instance.
func NewMinifier() *Minifier {
	m := minify.New()

	// CSS and JS minifiers
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("application/javascript", js.Minify)

	// HTML minifier with custom settings to preserve structure
	htmlMinifier := &html.Minifier{
		KeepEndTags:         true,
		KeepSpecialComments: true,
		KeepDefaultAttrVals: false,
		KeepDocumentTags:    true,
		KeepWhitespace:      false,
	}
	m.AddFunc("text/html", htmlMinifier.Minify)
	m.AddFunc("application/html", htmlMinifier.Minify)

	return &Minifier{m: m}
}

// MinifyFile minifies a file and returns the minified content.
func (m *Minifier) MinifyFile(contentType string, filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buf bytes.Buffer
	err = m.m.Minify(contentType, &buf, file)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MinifyBytes minifies byte content.
func (m *Minifier) MinifyBytes(contentType string, data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	var buf bytes.Buffer

	err := m.m.Minify(contentType, &buf, reader)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MinifyString minifies string content.
func (m *Minifier) MinifyString(contentType string, data string) (string, error) {
	minBytes, err := m.MinifyBytes(contentType, []byte(data))
	if err != nil {
		return "", err
	}
	return string(minBytes), nil
}

// ServeMinifiedFile serves a minified file.
func (m *Minifier) ServeMinifiedFile(w http.ResponseWriter, r *http.Request, filePath string) {
	ext := strings.ToLower(filepath.Ext(filePath))

	var contentType string
	switch ext {
	case ".css":
		contentType = "text/css"
	case ".js":
		contentType = "application/javascript"
	case ".html":
		contentType = "text/html"
	default:
		http.ServeFile(w, r, filePath)
		return
	}

	// Get file info for modification time
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		http.ServeFile(w, r, filePath)
		return
	}

	minifiedData, err := m.MinifyFile(contentType, filePath)
	if err != nil {
		// Fallback to serving original file if minification fails
		http.ServeFile(w, r, filePath)
		return
	}

	// Use ServeContent for proper Content-Length and caching headers
	w.Header().Set("Content-Type", contentType+"; charset=utf-8")
	http.ServeContent(w, r, filepath.Base(filePath), fileInfo.ModTime(), bytes.NewReader(minifiedData))
}
