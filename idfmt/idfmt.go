// Package idfmt holds the canonical document-ID *format* rules.
// It is the single source of truth for "is this string a valid document id".
//
// Two forms are valid:
// a legacy random id of IDLength uppercase letters/digits (`X7F4K2`),
// and a prefixed sequential id: prefix segments and a number,
// joined by one separator, "-" (`MEM-DB-35`), "_" (`MEM_DB_35`) or none (`B95`).
// Without a separator a prefixed id can look like a legacy one (`BUG123`),
// so telling them apart takes the consuming application's own prefix scheme.
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
// Without a separator the prefix must end in a letter, so the number is exactly the trailing digits.
var prefixedPatterns = map[string]*regexp.Regexp{
	"-": regexp.MustCompile(`^([A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*)-([1-9][0-9]{0,8})$`),
	"_": regexp.MustCompile(`^([A-Z][A-Z0-9]*(?:_[A-Z][A-Z0-9]*)*)_([1-9][0-9]{0,8})$`),
	"":  regexp.MustCompile(`^([A-Z](?:[A-Z0-9]*[A-Z])?)([1-9][0-9]{0,8})$`),
}

var segmentPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*$`)

// IsValidSeparator reports whether sep may join an id's prefix segments and number.
func IsValidSeparator(sep string) bool {
	_, ok := prefixedPatterns[sep]
	return ok
}

// IsValidID reports whether id is a legacy or a prefixed document id.
// Legacy TIKI-XXXXXX values are NOT valid and have no compatibility shim.
func IsValidID(id string) bool {
	return IsLegacyID(id) || IsPrefixedID(id)
}

// IsLegacyID reports whether id is a random id of IDLength characters.
func IsLegacyID(id string) bool {
	return legacyPattern.MatchString(id)
}

// IsPrefixedID reports whether id has the shape of a prefixed id with any allowed separator.
func IsPrefixedID(id string) bool {
	_, _, ok := splitAny(id)
	return ok
}

// SplitID returns the prefix and number of id, prefixed with separator sep.
func SplitID(id, sep string) (prefix string, number int, ok bool) {
	re, known := prefixedPatterns[sep]
	if !known || len(id) > MaxIDLength {
		return "", 0, false
	}
	m := re.FindStringSubmatch(id)
	if m == nil {
		return "", 0, false
	}
	n, err := strconv.Atoi(m[2])
	if err != nil {
		return "", 0, false
	}
	return m[1], n, true
}

// SplitPrefixed returns the prefix and number of an id prefixed with "-".
func SplitPrefixed(id string) (prefix string, number int, ok bool) {
	return SplitID(id, "-")
}

func splitAny(id string) (prefix string, number int, ok bool) {
	for _, s := range []string{"-", "_", ""} {
		if p, n, found := SplitID(id, s); found {
			return p, n, true
		}
	}
	return "", 0, false
}

// IsValidPrefixSegment reports whether s can be one segment of an id prefix.
func IsValidPrefixSegment(s string) bool {
	return segmentPattern.MatchString(s)
}

// Compare orders ids for sorting: legacy and other ids first,
// prefixed ids by prefix and then by number, so MEM-9 precedes MEM-10 and B9 precedes B10.
// Grouping by form keeps the order transitive when the forms are mixed;
// a legacy-shaped id counts as legacy even when it could also read as prefixed.
func Compare(a, b string) int {
	pa, na, aOK := sortKey(a)
	pb, nb, bOK := sortKey(b)
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
		return strings.Compare(a, b)
	case aOK:
		return 1
	case bOK:
		return -1
	}
	return strings.Compare(a, b)
}

func sortKey(id string) (string, int, bool) {
	if IsLegacyID(id) {
		return "", 0, false
	}
	return splitAny(id)
}
