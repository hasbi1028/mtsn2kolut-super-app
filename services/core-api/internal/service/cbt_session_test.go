package service

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func cbtSessionTestUUID(seed byte) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte{seed}, Valid: true}
}

func TestCbtSessionNormalizeScopeMixAndAssignment(t *testing.T) {
	scopeTests := []struct {
		name string
		in   string
		want string
	}{
		{name: "class", in: "class", want: "class"},
		{name: "grade", in: "grade", want: "grade"},
		{name: "school", in: "school", want: "school"},
		{name: "custom", in: "custom", want: "custom"},
		{name: "fallback", in: "invalid", want: "class"},
	}
	for _, tt := range scopeTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeScopeType(tt.in); got != tt.want {
				t.Fatalf("normalizeScopeType(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}

	mixTests := []struct {
		name      string
		value     string
		scopeType string
		want      string
	}{
		{name: "class forced", value: "mixed_scope", scopeType: "class", want: "same_class"},
		{name: "grade forced", value: "same_class", scopeType: "grade", want: "same_grade"},
		{name: "school forced", value: "same_class", scopeType: "school", want: "mixed_scope"},
		{name: "custom forced", value: "same_grade", scopeType: "custom", want: "mixed_scope"},
		{name: "unknown accepts explicit", value: "same_class", scopeType: "unknown", want: "same_class"},
		{name: "unknown fallback", value: "bad", scopeType: "unknown", want: "same_grade"},
	}
	for _, tt := range mixTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeMixPolicy(tt.value, tt.scopeType); got != tt.want {
				t.Fatalf("normalizeMixPolicy(%q, %q) = %q, want %q", tt.value, tt.scopeType, got, tt.want)
			}
		})
	}

	modeTests := []struct {
		name string
		in   string
		want string
	}{
		{name: "manual", in: "manual", want: "manual"},
		{name: "random balanced", in: "random_balanced", want: "random_balanced"},
		{name: "random by gender", in: "random_by_gender", want: "random_by_gender"},
		{name: "random by accommodation", in: "random_by_accommodation", want: "random_by_accommodation"},
		{name: "fallback", in: "bad", want: "random_balanced"},
	}
	for _, tt := range modeTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeAssignmentMode(tt.in); got != tt.want {
				t.Fatalf("normalizeAssignmentMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCbtSessionPgNumeric(t *testing.T) {
	n := pgNumeric(87.5)
	if !n.Valid {
		t.Fatal("pgNumeric() valid = false, want true")
	}
	f, _ := n.Int.Float64()
	scale := new(big.Float).SetInt64(1)
	for i := int32(0); i < -n.Exp; i++ {
		scale.Mul(scale, big.NewFloat(10))
	}
	divisor, _ := scale.Float64()
	if got := f / divisor; got != 87.5 {
		t.Fatalf("pgNumeric(87.5) = %v, want 87.5", got)
	}
}

func TestCbtSessionUUIDHelpers(t *testing.T) {
	ids := []pgtype.UUID{cbtSessionTestUUID(1), cbtSessionTestUUID(2), pgtype.UUID{}}
	jsonBytes, err := UUIDsToJSON(ids)
	if err != nil {
		t.Fatalf("UUIDsToJSON() error = %v", err)
	}
	var got []string
	if err := json.Unmarshal(jsonBytes, &got); err != nil {
		t.Fatalf("UUIDsToJSON() output unmarshal error = %v", err)
	}
	want := []string{
		"01000000-0000-0000-0000-000000000000",
		"02000000-0000-0000-0000-000000000000",
		"",
	}
	if len(got) != len(want) {
		t.Fatalf("UUIDsToJSON() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("UUIDsToJSON()[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	shuffled := shuffleUUIDs(ids)
	if len(shuffled) != len(ids) {
		t.Fatalf("shuffleUUIDs() len = %d, want %d", len(shuffled), len(ids))
	}
	if ids[0] != cbtSessionTestUUID(1) || ids[1] != cbtSessionTestUUID(2) {
		t.Fatalf("shuffleUUIDs() mutated input = %v", ids)
	}
	counts := map[string]int{}
	for _, id := range ids {
		counts[pgUUIDString(id)]++
	}
	for _, id := range shuffled {
		counts[pgUUIDString(id)]--
	}
	for id, count := range counts {
		if count != 0 {
			t.Fatalf("shuffleUUIDs() changed membership for %q: count delta %d", id, count)
		}
	}
}
