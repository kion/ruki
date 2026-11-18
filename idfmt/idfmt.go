// Package idfmt holds the canonical document-ID *format* rules.
// It is the single source of truth for "is this string a valid document id".
//
// Two forms are valid:
// a legacy random id of IDLength uppercase letters/digits (`X7F4K2`),
// and a prefixed sequential id (`MEM-DB-35`).
// A prefixed id always contains a hyphen and a legacy id never does,
// so the two forms can never be confused.
//
// ID *generation* lives in the consuming application, not here.
package idfmt

import (
	"regexp"
	"strconv"
	"strings"
)

// IDLength is the number of characters in a legacy random document ID.
const IDLength = 6

// MaxIDLength caps prefixed ids so they stay usable in narrow views.
const MaxIDLength = 64

// Keep the {6} width in sync with IDLength.
var legacyPattern = regexp.MustCompile(`^[A-Z0-9]{6}$`)

// The number is capped at 9 digits so it fits an int on every platform.
var prefixedPattern = regexp.MustCompile(`^([A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*)-([1-9][0-9]{0,8})$`)

var segmentPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*$`)

// IsValidID reports whether id is a legacy or a prefixed document id.
// Legacy TIKI-XXXXXX values are NOT valid and have no compatibility shim.
func IsValidID(id string) bool {
	return IsLegacyID(id) || IsPrefixedID(id)
}

// IsLegacyID reports whether id is a random id of IDLength characters.
func IsLegacyID(id string) bool {
	return legacyPattern.MatchString(id)
}

// IsPrefixedID reports whether id is a prefixed sequential id such as MEM-DB-35.
func IsPrefixedID(id string) bool {
	return len(id) <= MaxIDLength && prefixedPattern.MatchString(id)
}

// SplitPrefixed returns the prefix and number of a prefixed id.
func SplitPrefixed(id string) (prefix string, number int, ok bool) {
	if len(id) > MaxIDLength {
		return "", 0, false
	}
	m := prefixedPattern.FindStringSubmatch(id)
	if m == nil {
		return "", 0, false
	}
	n, err := strconv.Atoi(m[2])
	if err != nil {
		return "", 0, false
	}
	return m[1], n, true
}

// IsValidPrefixSegment reports whether s can be one segment of an id prefix.
func IsValidPrefixSegment(s string) bool {
	return segmentPattern.MatchString(s)
}

// Compare orders ids for sorting: every other id before prefixed ids,
// prefixed ids by prefix and then by number, so MEM-9 precedes MEM-10.
// Grouping by form keeps the order transitive when the two forms are mixed.
func Compare(a, b string) int {
	pa, na, aOK := SplitPrefixed(a)
	pb, nb, bOK := SplitPrefixed(b)
	switch {
	case aOK && bOK:
		if c := strings.Compare(pa, pb); c != 0 {
			return c
		}
		switch {
		case na < nb:
			return -1
		case na > nb:
			return 1
		}
		return 0
	case aOK:
		return 1
	case bOK:
		return -1
	}
	return strings.Compare(a, b)
}
