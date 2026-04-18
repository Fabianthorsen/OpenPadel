package store_test

import (
	"database/sql"
	"testing"
)

func TestMigration_TimedAmericano_ColumnsExist(t *testing.T) {
	s := newTestStore(t)
	defer s.Close()

	tests := []struct {
		colName string
	}{
		{"total_duration_minutes"},
		{"buffer_seconds"},
		{"round_duration_seconds"},
		{"round_started_at"},
	}

	for _, tt := range tests {
		var colName string

		row := s.DB().QueryRow(
			"SELECT name FROM pragma_table_info('sessions') WHERE name = ?",
			tt.colName,
		)
		err := row.Scan(&colName)

		if err != nil && err != sql.ErrNoRows {
			t.Fatalf("checking column %s: %v", tt.colName, err)
		}
		if err == sql.ErrNoRows {
			t.Errorf("column %s does not exist in sessions table", tt.colName)
		}
	}
}

func TestMigration_TimedAmericano_ExistingDataPreserved(t *testing.T) {
	s := newTestStore(t)
	defer s.Close()

	sess, err := s.CreateSession(2, 24, "Test Session", "americano", 2, 6, nil, nil, nil, nil, nil, "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	loaded, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}

	if loaded.ID != sess.ID {
		t.Errorf("expected session ID %s, got %s", sess.ID, loaded.ID)
	}
	if loaded.Name != "Test Session" {
		t.Errorf("expected name 'Test Session', got %q", loaded.Name)
	}
}
