package design

import ds "github.com/SCKelemen/design-system"

// DesignTokens aliases the design-system token type for local TUI imports.
type DesignTokens = ds.DesignTokens

// DefaultTheme returns the default design token palette.
func DefaultTheme() *DesignTokens { return ds.DefaultTheme() }

// MidnightTheme returns the Midnight design token palette.
func MidnightTheme() *DesignTokens { return ds.MidnightTheme() }

// NordTheme returns the Nord design token palette.
func NordTheme() *DesignTokens { return ds.NordTheme() }

// PaperTheme returns the Paper design token palette.
func PaperTheme() *DesignTokens { return ds.PaperTheme() }

// WrappedTheme returns the Wrapped design token palette.
func WrappedTheme() *DesignTokens { return ds.WrappedTheme() }
