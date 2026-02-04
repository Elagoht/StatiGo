// Package context provides shared context keys and helpers for the Statigo framework.
package context

import (
	gocontext "context"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

// Context keys used across the framework.
const (
	LanguageKey      ContextKey = "language"
	CanonicalPathKey ContextKey = "canonicalPath"
	PageTitleKey     ContextKey = "pageTitle"
	StrategyKey      ContextKey = "cacheStrategy"
	LayoutDataKey    ContextKey = "layoutData"
)

// GetLanguage retrieves the language from context.
func GetLanguage(ctx gocontext.Context) string {
	if lang, ok := ctx.Value(LanguageKey).(string); ok {
		return lang
	}
	return "en" // Fallback
}

// SetLanguage creates a new context with the language set.
func SetLanguage(ctx gocontext.Context, lang string) gocontext.Context {
	return gocontext.WithValue(ctx, LanguageKey, lang)
}

// GetCanonicalPath retrieves the canonical path from context.
func GetCanonicalPath(ctx gocontext.Context) string {
	if canonical, ok := ctx.Value(CanonicalPathKey).(string); ok {
		return canonical
	}
	return ""
}

// SetCanonicalPath creates a new context with the canonical path set.
func SetCanonicalPath(ctx gocontext.Context, canonical string) gocontext.Context {
	return gocontext.WithValue(ctx, CanonicalPathKey, canonical)
}

// GetPageTitle retrieves the page title from context.
func GetPageTitle(ctx gocontext.Context) string {
	if title, ok := ctx.Value(PageTitleKey).(string); ok {
		return title
	}
	return ""
}

// SetPageTitle creates a new context with the page title set.
func SetPageTitle(ctx gocontext.Context, title string) gocontext.Context {
	return gocontext.WithValue(ctx, PageTitleKey, title)
}

// GetStrategy retrieves the cache strategy from context.
func GetStrategy(ctx gocontext.Context) string {
	if strategy, ok := ctx.Value(StrategyKey).(string); ok {
		return strategy
	}
	return ""
}

// SetStrategy creates a new context with the cache strategy set.
func SetStrategy(ctx gocontext.Context, strategy string) gocontext.Context {
	return gocontext.WithValue(ctx, StrategyKey, strategy)
}

// GetLayoutData retrieves the layout data from context.
func GetLayoutData(ctx gocontext.Context) interface{} {
	return ctx.Value(LayoutDataKey)
}

// SetLayoutData creates a new context with the layout data set.
func SetLayoutData(ctx gocontext.Context, data interface{}) gocontext.Context {
	return gocontext.WithValue(ctx, LayoutDataKey, data)
}
