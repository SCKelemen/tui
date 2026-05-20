// Package tui is deprecated. All development has moved to
// github.com/SCKelemen/tui/v2, which is a clean rewrite around
// Bubble Tea, github.com/SCKelemen/text (for UAX #29 grapheme
// handling), and github.com/SCKelemen/layout.
//
// Migration:
//
//	import tui "github.com/SCKelemen/tui/v2"
//
// The last fully-supported v1 release is v1.6.0. The git tag
// remains valid for module proxy fetches if you need to pin to v1.
//
// This module is preserved as a deprecation stub so that go.mod files
// referencing github.com/SCKelemen/tui at semver versions above the
// last v1 tag continue to resolve, but no functionality is exported.
package tui
