package sql

import (
	"context"
	"testing"
)

func TestQuery(t *testing.T) {
	rows, err := client.Query(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Errorf("close rows: %v", err)
		}
	}()
	if !rows.Next() {
		t.Fatal("SELECT 1 returned no row")
	}
	var got int
	if err := rows.Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("SELECT 1 = %d, want 1", got)
	}
}

func TestLegacyQureyRow(t *testing.T) {
	var got int
	if err := client.QureyRow(context.Background(), "SELECT 1").Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("SELECT 1 = %d, want 1", got)
	}
}
