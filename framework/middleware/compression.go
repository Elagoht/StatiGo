package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
)

const (
	compressionBrotli = "br"
	compressionGzip   = "gzip"
)

// CompressibleContentTypes defines which content types should be compressed.
var CompressibleContentTypes = map[string]bool{
	"text/html":              true,
	"text/css":               true,
	"text/plain":             true,
	"text/xml":               true,
	"text/javascript":        true,
	"application/javascript": true,
	"application/json":       true,
	"application/xml":        true,
	"application/rss+xml":    true,
	"application/atom+xml":   true,
	"application/xhtml+xml":  true,
	"image/svg+xml":          true,
	"font/woff":              true,
	"font/woff2":             true,
}

var (
	gzipWriterPool = sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(io.Discard)
		},
	}
	brotliWriterPool = sync.Pool{
		New: func() interface{} {
			return brotli.NewWriter(io.Discard)
		},
	}
)

type compressionResponseWriter struct {
	io.Writer
	http.ResponseWriter
	compressionType string
	wroteHeader     bool
}

func (w *compressionResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	// Set appropriate encoding header if compression is used
	if w.compressionType != "" {
		w.ResponseWriter.Header().Set("Content-Encoding", w.compressionType)
		w.ResponseWriter.Header().Del("Content-Length") // Length will change with compression
		w.ResponseWriter.Header().Add("Vary", "Accept-Encoding")
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *compressionResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.Writer.Write(b)
}

// Compression middleware that prefers Brotli over gzip.
func Compression() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip compression for certain conditions
			if r.Header.Get("Upgrade") != "" || // WebSocket or other upgrade
				r.Method == "HEAD" ||
				strings.Contains(r.Header.Get("Content-Encoding"), "identity") {
				next.ServeHTTP(w, r)
				return
			}

			// Determine best compression method from Accept-Encoding header
			acceptEncoding := r.Header.Get("Accept-Encoding")
			compressionType := selectCompression(acceptEncoding)

			// No compression support, serve normally
			if compressionType == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Create a wrapper to intercept the response
			crw := &compressionResponseWriter{
				ResponseWriter:  w,
				Writer:          w,
				compressionType: "", // Will be set if we actually compress
			}

			// We need to wrap again to check content type before compressing
			wrappedWriter := &contentTypeCheckWriter{
				compressionResponseWriter: crw,
				originalWriter:            w,
				compressionType:           compressionType,
			}

			next.ServeHTTP(wrappedWriter, r)

			// Cleanup: close the compression writer if it was created
			if closer, ok := crw.Writer.(io.WriteCloser); ok && crw.Writer != w {
				closer.Close()
			}
		})
	}
}

// contentTypeCheckWriter waits for WriteHeader or first Write to determine if compression should be used.
type contentTypeCheckWriter struct {
	*compressionResponseWriter
	originalWriter  http.ResponseWriter
	compressionType string
	checkedType     bool
}

func (w *contentTypeCheckWriter) WriteHeader(code int) {
	if !w.checkedType {
		w.setupCompression()
	}
	w.compressionResponseWriter.WriteHeader(code)
}

func (w *contentTypeCheckWriter) Write(b []byte) (int, error) {
	if !w.checkedType {
		w.setupCompression()
	}
	return w.compressionResponseWriter.Write(b)
}

func (w *contentTypeCheckWriter) setupCompression() {
	w.checkedType = true

	contentType := w.originalWriter.Header().Get("Content-Type")
	if contentType == "" {
		return
	}

	// Parse content type (remove charset and other parameters)
	ct := strings.Split(contentType, ";")[0]
	ct = strings.TrimSpace(ct)

	// Check if this content type should be compressed
	if !CompressibleContentTypes[ct] {
		return
	}

	// Check if response is already compressed
	if w.originalWriter.Header().Get("Content-Encoding") != "" {
		return
	}

	// Set up compression writer
	switch w.compressionType {
	case compressionBrotli:
		bw := brotliWriterPool.Get().(*brotli.Writer)
		bw.Reset(w.originalWriter)
		w.compressionResponseWriter.Writer = bw
		w.compressionResponseWriter.compressionType = compressionBrotli

	case compressionGzip:
		gw := gzipWriterPool.Get().(*gzip.Writer)
		gw.Reset(w.originalWriter)
		w.compressionResponseWriter.Writer = gw
		w.compressionResponseWriter.compressionType = compressionGzip
	}
}

// selectCompression chooses the best compression method based on Accept-Encoding header.
// Prefers Brotli over gzip.
func selectCompression(acceptEncoding string) string {
	if acceptEncoding == "" {
		return ""
	}

	// Parse Accept-Encoding header
	encodings := strings.Split(acceptEncoding, ",")
	supportsBrotli := false
	supportsGzip := false

	for _, encoding := range encodings {
		encoding = strings.TrimSpace(strings.ToLower(encoding))

		// Handle quality values (e.g., "br;q=0.8")
		parts := strings.Split(encoding, ";")
		encodingType := strings.TrimSpace(parts[0])

		// Check for quality value
		quality := 1.0
		if len(parts) > 1 {
			if strings.HasPrefix(parts[1], "q=") {
				// If quality is 0, skip this encoding
				if strings.TrimSpace(parts[1]) == "q=0" {
					continue
				}
			}
		}

		if quality > 0 {
			switch encodingType {
			case "br":
				supportsBrotli = true
			case "gzip":
				supportsGzip = true
			}
		}
	}

	// Prefer Brotli over gzip
	if supportsBrotli {
		return compressionBrotli
	}
	if supportsGzip {
		return compressionGzip
	}

	return ""
}
