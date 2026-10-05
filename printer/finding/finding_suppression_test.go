package finding_test

import (
	"context"
	// encoding/json (v1) deliberately: it sorts map keys, making the
	// byte-identity comparison in TestEmitSuppressedAcceptedOffIsByteIdentical
	// deterministic — the v2 Marshal used by sibling tests does not sort map
	// keys (the jsonv2 recurrence #5 lesson).
	"encoding/json"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/printer/finding"
	gofinding "github.com/larsartmann/go-finding"
)

// acceptAll and acceptNone are the two predicates used by the suppression
// tests; they stand in for internal/accept's AcceptedSet.IsAccepted.
func acceptAll(domain.ProcessedCloneGroup) bool  { return true }
func acceptNone(domain.ProcessedCloneGroup) bool { return false }

// TestAcceptedGroupFindingsCarryInSourceSuppression pins the emit path of
// EmitSuppressedAccepted: every finding of an accepted group carries an
// in-source suppression naming the accept directive, while a non-accepted
// group (or a nil/declining predicate) emits unsuppressed findings exactly
// as before.
func TestAcceptedGroupFindingsCarryInSourceSuppression(t *testing.T) {
	group := testGroup("accept0123456789ab", 10)

	findings := finding.ToFindings(group, finding.Options{
		Version:                "test",
		EmitSuppressedAccepted: true,
		Accepted:               acceptAll,
	})

	if len(findings) != len(group.Clones) {
		t.Fatalf("ToFindings emitted %d findings, want %d", len(findings), len(group.Clones))
	}

	for _, f := range findings {
		if f.Suppression == nil {
			t.Errorf("finding %q: Suppression = nil, want in-source annotation", f.ID)

			continue
		}

		if f.Suppression.Kind != gofinding.SuppressionInSource {
			t.Errorf("finding %q: Suppression.Kind = %q, want %q",
				f.ID, f.Suppression.Kind, gofinding.SuppressionInSource)
		}

		if f.Suppression.Rule != gofinding.RuleName(finding.RuleCloneDetected) {
			t.Errorf("finding %q: Suppression.Rule = %q, want %q",
				f.ID, f.Suppression.Rule, finding.RuleCloneDetected)
		}

		if f.Suppression.Reason != "//art-dupl:accept directive" {
			t.Errorf("finding %q: Suppression.Reason = %q, want the directive reason", f.ID, f.Suppression.Reason)
		}

		if !f.IsValid() {
			t.Errorf("finding %q with suppression is not valid", f.ID)
		}
	}
}

// TestUnacceptedGroupStaysUnsuppressed: with the flag on, a predicate that
// declines the group must leave the findings unsuppressed — suppression is
// per-group acceptance, never blanket.
func TestUnacceptedGroupStaysUnsuppressed(t *testing.T) {
	group := testGroup("plain0123456789ab", 10)

	for name, opts := range map[string]finding.Options{
		"predicate declines": {EmitSuppressedAccepted: true, Accepted: acceptNone},
		"nil predicate":      {EmitSuppressedAccepted: true},
	} {
		findings := finding.ToFindings(group, opts)

		for _, f := range findings {
			if f.Suppression != nil {
				t.Errorf("%s: finding %q unexpectedly suppressed (kind %q)",
					name, f.ID, f.Suppression.Kind)
			}
		}
	}
}

// TestSuppressionRoundTripsViaSARIFIncludeSuppressed is G7: the suppression
// annotation survives ToSARIFWithOpts(WithIncludeSuppressed()) →
// FindingsFromSARIF with kind and reason intact (SARIF suppression entries
// do not carry the rule name — the rule lives on the result itself — so the
// round-trip pins Kind and Reason, not Rule), while the DEFAULT SARIF export
// drops suppressed findings entirely.
func TestSuppressionRoundTripsViaSARIFIncludeSuppressed(t *testing.T) {
	group := testGroup("round0123456789ab", 10)

	report := finding.ToReport(
		[]domain.ProcessedCloneGroup{group},
		finding.Options{
			Version:                "test",
			EmitSuppressedAccepted: true,
			Accepted:               acceptAll,
		},
	)

	data, err := report.ToSARIFWithOpts(gofinding.WithIncludeSuppressed())
	if err != nil {
		t.Fatalf("ToSARIFWithOpts(WithIncludeSuppressed()) = %v, want nil", err)
	}

	imported, err := gofinding.FindingsFromSARIF(context.Background(), data)
	if err != nil {
		t.Fatalf("FindingsFromSARIF() = %v, want nil", err)
	}

	if len(imported) != len(group.Clones) {
		t.Fatalf("round-trip kept %d findings, want %d", len(imported), len(group.Clones))
	}

	for _, f := range imported {
		if f.GroupID != finding.GroupIDOf(group) {
			t.Errorf("imported finding %q GroupID = %q, want %q", f.ID, f.GroupID, finding.GroupIDOf(group))
		}

		if f.Suppression == nil {
			t.Errorf("imported finding %q lost its suppression", f.ID)

			continue
		}

		if f.Suppression.Kind != gofinding.SuppressionInSource {
			t.Errorf("imported finding %q Suppression.Kind = %q, want %q",
				f.ID, f.Suppression.Kind, gofinding.SuppressionInSource)
		}

		if f.Suppression.Reason != "//art-dupl:accept directive" {
			t.Errorf("imported finding %q Suppression.Reason = %q, want the directive reason",
				f.ID, f.Suppression.Reason)
		}
	}

	defaultData, err := report.ToSARIF()
	if err != nil {
		t.Fatalf("ToSARIF() = %v, want nil", err)
	}

	defaultImported, err := gofinding.FindingsFromSARIF(context.Background(), defaultData)
	if err != nil {
		t.Fatalf("FindingsFromSARIF(default) = %v, want nil", err)
	}

	if len(defaultImported) != 0 {
		t.Errorf("default SARIF export kept %d findings, want 0 (suppressed findings are dropped)",
			len(defaultImported))
	}
}

// TestEmitSuppressedAcceptedOffIsByteIdentical is the M31 guardrail: with the
// flag off (the default), the Accepted predicate is never consulted and the
// emitted findings are byte-identical to the pre-flag output — suppression
// can never leak into the default wire format.
func TestEmitSuppressedAcceptedOffIsByteIdentical(t *testing.T) {
	group := testGroup("guard0123456789ab", 10)

	baseline := finding.ToFindings(group, finding.Options{Version: "test"})

	flagOff := finding.ToFindings(group, finding.Options{
		Version:                "test",
		EmitSuppressedAccepted: false,
		// A predicate that WOULD accept the group: if the flag-off path
		// consulted it, this test would fail on the suppression assert.
		Accepted: acceptAll,
	})

	marshal := func(fs []gofinding.Finding) []byte {
		t.Helper()

		data, err := json.Marshal(fs)
		if err != nil {
			t.Fatalf("marshal findings: %v", err)
		}

		return data
	}

	if string(marshal(baseline)) != string(marshal(flagOff)) {
		t.Error("EmitSuppressedAccepted=false output differs from the zero-value baseline; " +
			"the flag must be byte-identical to the historical output")
	}

	for _, f := range flagOff {
		if f.Suppression != nil {
			t.Errorf("finding %q carries a suppression with the flag off", f.ID)
		}
	}
}
