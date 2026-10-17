package perfdata

import (
	"bytes"
	"testing"
)

func TestSmallMatchesStoredFixture(t *testing.T) {
	stored, err := Fixture("order_small.json")
	if err != nil {
		t.Fatal(err)
	}
	got := ValidOrder(SmallItems)
	if !bytes.Equal(got, stored) {
		t.Fatalf("recipe differs from stored fixture:\n%s\n%s", got, stored)
	}
	if len(got) != 492 {
		t.Fatalf("len = %d, want 492", len(got))
	}
}

func TestDeterministic(t *testing.T) {
	if !bytes.Equal(ValidOrder(MediumItems), ValidOrder(MediumItems)) || !bytes.Equal(ErrorOrder(), ErrorOrder()) {
		t.Fatal("generation not deterministic")
	}
}

func TestErrorOrderBadSkus(t *testing.T) {
	doc := ErrorOrder()
	if n := bytes.Count(doc, []byte(`"sku": "bad"`)); n != ErrorItems/ErrorEvery {
		t.Fatalf("bad skus = %d, want %d", n, ErrorItems/ErrorEvery)
	}
	if n := bytes.Count(doc, []byte(`"sku"`)); n != ErrorItems {
		t.Fatalf("items = %d, want %d", n, ErrorItems)
	}
}

func TestSchemaFixtureReadable(t *testing.T) {
	if b, err := Fixture("order_schema.json"); err != nil || len(b) == 0 {
		t.Fatalf("schema fixture: %v", err)
	}
}
