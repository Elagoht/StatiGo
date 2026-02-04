package templates

import (
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"time"
)

// PrettyJson formats data as indented JSON.
func PrettyJson(v interface{}) template.JS {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return template.JS("{}")
	}
	return template.JS(bytes)
}

// SafeHTML marks a string as safe HTML.
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// SafeURL marks a string as a safe URL.
func SafeURL(s string) template.URL {
	return template.URL(strings.TrimSpace(s))
}

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b
}

// Sub returns the difference of two integers.
func Sub(a, b int) int {
	return a - b
}

// SubFloat returns the difference of two floats.
func SubFloat(a, b float64) float64 {
	return a - b
}

// FormatPrice formats a price with comma as decimal separator.
func FormatPrice(price float64) string {
	formatted := strings.Replace(
		strings.TrimSpace(strings.Replace(fmt.Sprintf("%.2f", price), ".", ",", 1)),
		"",
		"",
		-1,
	)
	return formatted
}

// PriceWhole returns the whole number part of a price.
func PriceWhole(price float64) string {
	return fmt.Sprintf("%.0f", price)
}

// PriceDecimal returns the decimal part of a price.
func PriceDecimal(price float64) string {
	decimal := fmt.Sprintf("%.2f", price-float64(int(price)))
	parts := strings.Split(decimal, ".")
	if len(parts) > 1 {
		return "," + parts[1]
	}
	return ",00"
}

// CurrencySymbol returns the currency symbol for a given currency code.
func CurrencySymbol(code string) string {
	symbols := map[string]string{
		"TRY": "₺",
		"USD": "$",
		"EUR": "€",
		"GBP": "£",
		"JPY": "¥",
	}

	if symbol, ok := symbols[code]; ok {
		return symbol
	}
	return code
}

// Div returns the integer division of two integers.
func Div(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

// Mod returns the modulo of two integers.
func Mod(a, b int) int {
	if b == 0 {
		return 0
	}
	return a % b
}

// Until returns a slice of integers from 0 to n-1.
func Until(n int) []int {
	if n <= 0 {
		return nil
	}
	out := make([]int, n)
	for i := range n {
		out[i] = i
	}
	return out
}

// Slugify converts a string to a URL-friendly slug.
func Slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Turkish character replacements
	replacements := map[string]string{
		"ı": "i",
		"ğ": "g",
		"ü": "u",
		"ş": "s",
		"ö": "o",
		"ç": "c",
		"İ": "i",
		"Ğ": "g",
		"Ü": "u",
		"Ş": "s",
		"Ö": "o",
		"Ç": "c",
	}

	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")

	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")

	return s
}

// FormatDate formats an ISO 8601 date string to a readable format.
func FormatDate(dateStr, lang string) string {
	if dateStr == "" {
		return ""
	}

	// Parse ISO 8601 format
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr // Return original if parsing fails
	}

	return formatTime(t, lang)
}

// FormatDateTime formats a time.Time object to a readable format.
func FormatDateTime(t time.Time, lang string) string {
	if t.IsZero() {
		return ""
	}
	return formatTime(t, lang)
}

// formatTime is a helper function that formats a time.Time object.
func formatTime(t time.Time, lang string) string {
	// Format based on language
	if lang == "tr" {
		// Turkish format: 25 Aralık 2025
		months := map[time.Month]string{
			time.January:   "Ocak",
			time.February:  "Şubat",
			time.March:     "Mart",
			time.April:     "Nisan",
			time.May:       "Mayıs",
			time.June:      "Haziran",
			time.July:      "Temmuz",
			time.August:    "Ağustos",
			time.September: "Eylül",
			time.October:   "Ekim",
			time.November:  "Kasım",
			time.December:  "Aralık",
		}
		return strings.TrimSpace(t.Format("2") + " " + months[t.Month()] + " " + t.Format("2006"))
	}

	// English format: December 16, 2025
	return t.Format("January 2, 2006")
}

// YouTubeID extracts the video ID from various YouTube URL formats.
func YouTubeID(input string) string {
	if input == "" {
		return ""
	}

	input = strings.TrimSpace(input)

	// If it's already just an ID (11 characters, alphanumeric with hyphens/underscores)
	if regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`).MatchString(input) {
		return input
	}

	// Try to extract from URL patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/)([a-zA-Z0-9_-]{11})`),
		regexp.MustCompile(`youtube\.com/embed/([a-zA-Z0-9_-]{11})`),
		regexp.MustCompile(`youtube\.com/v/([a-zA-Z0-9_-]{11})`),
	}

	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(input); len(matches) > 1 {
			return matches[1]
		}
	}

	return input
}

// Dict creates a map from key-value pairs.
func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires an even number of arguments")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

// Set adds a key-value pair to a map and returns empty string.
func Set(dict map[string]interface{}, key string, value interface{}) string {
	dict[key] = value
	return ""
}

// HasDiscount returns true if oldPrice is greater than price.
func HasDiscount(price float64, oldPrice *float64) bool {
	if oldPrice == nil {
		return false
	}
	return price < *oldPrice
}
