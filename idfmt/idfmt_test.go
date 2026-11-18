package idfmt

import (
	"sort"
	"strings"
	"testing"
)

func TestIsValidID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"ABC123", true},
		{"000000", true},
		{"ZZZZZZ", true},
		{"abc123", false}, // lowercase not accepted
		{"ABC12", false},  // too short
		{"ABC1234", false},
		{"ABC12!", false},
		{"", false},
		{"TIKI-ABC123", false}, // legacy form is NOT a valid id
		{"MEM-1", true},
		{"MEM-DB-35", true},
		{"FEAT-FE-15", true},
		{"T2-7", true},
		{"B-55", true},
		{"MEM-0", false},   // numbers start at 1
		{"MEM-01", false},  // no leading zeros
		{"MEM-", false},    // no number
		{"-12", false},     // no prefix
		{"2B-12", false},   // a segment starts with a letter
		{"MEM--12", false}, // empty segment
		{"mem-12", false},  // lowercase prefix
		{"MEM-12A", false}, // number part is digits only
		{"MEM-1234567890", false},
		{"B55", false},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := IsValidID(tt.id); got != tt.want {
				t.Errorf("IsValidID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestIsValidIDRejectsOverlongPrefixedID(t *testing.T) {
	long := strings.Repeat("A", MaxIDLength) + "-1"
	if IsValidID(long) {
		t.Errorf("IsValidID accepted a %d-character id", len(long))
	}
}

func TestLegacyAndPrefixedAreDisjoint(t *testing.T) {
	for _, id := range []string{"ABC123", "MEM-12", "FEAT-FE-15"} {
		if IsLegacyID(id) == IsPrefixedID(id) {
			t.Errorf("%q: legacy=%v prefixed=%v, want exactly one", id, IsLegacyID(id), IsPrefixedID(id))
		}
	}
}

func TestSplitPrefixed(t *testing.T) {
	tests := []struct {
		id         string
		wantPrefix string
		wantNumber int
		wantOK     bool
	}{
		{"MEM-DB-35", "MEM-DB", 35, true},
		{"TIK-1", "TIK", 1, true},
		{"ABC123", "", 0, false},
		{"MEM-0", "", 0, false},
	}
	for _, tt := range tests {
		prefix, number, ok := SplitPrefixed(tt.id)
		if prefix != tt.wantPrefix || number != tt.wantNumber || ok != tt.wantOK {
			t.Errorf("SplitPrefixed(%q) = (%q, %d, %v), want (%q, %d, %v)",
				tt.id, prefix, number, ok, tt.wantPrefix, tt.wantNumber, tt.wantOK)
		}
	}
}

func TestIsValidPrefixSegment(t *testing.T) {
	for _, s := range []string{"MEM", "FE", "T2", "X"} {
		if !IsValidPrefixSegment(s) {
			t.Errorf("IsValidPrefixSegment(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "mem", "2B", "MEM-DB", "ME M"} {
		if IsValidPrefixSegment(s) {
			t.Errorf("IsValidPrefixSegment(%q) = true, want false", s)
		}
	}
}

func TestCompareOrdersNumbersNumerically(t *testing.T) {
	if Compare("MEM-9", "MEM-10") >= 0 {
		t.Error("MEM-9 must sort before MEM-10")
	}
	if Compare("MEM-10", "MEM-9") <= 0 {
		t.Error("MEM-10 must sort after MEM-9")
	}
	if Compare("MEM-10", "MEM-10") != 0 {
		t.Error("equal ids must compare equal")
	}
}

func TestCompareSortsMixedIDs(t *testing.T) {
	ids := []string{"MEM-10", "ZZZZZZ", "FEAT-FE-2", "MEM-9", "ABC123", "FEAT-16", "DB-3"}
	sort.Slice(ids, func(i, j int) bool { return Compare(ids[i], ids[j]) < 0 })
	want := []string{"ABC123", "ZZZZZZ", "DB-3", "FEAT-16", "FEAT-FE-2", "MEM-9", "MEM-10"}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("sorted = %v, want %v", ids, want)
		}
	}
}
