package schema

import (
	"testing"
)

func TestNewSchemaComparer(t *testing.T) {
	sourceTables := map[string]Table{
		"users": {Name: "users"},
	}
	targetTables := map[string]Table{
		"users": {Name: "users"},
	}

	comparer := NewSchemaComparer(sourceTables, targetTables)
	if comparer == nil {
		t.Error("Expected non-nil SchemaComparer")
	}
	if len(comparer.sourceTables) != 1 {
		t.Errorf("Expected 1 source table, got %d", len(comparer.sourceTables))
	}
	if len(comparer.targetTables) != 1 {
		t.Errorf("Expected 1 target table, got %d", len(comparer.targetTables))
	}
}

func TestCompareTables(t *testing.T) {
	tests := []struct {
		name           string
		sourceTables   map[string]Table
		targetTables   map[string]Table
		expectedDiffs  int
		expectedTypes  []string
		expectedChange []string
	}{
		{
			name: "no differences",
			sourceTables: map[string]Table{
				"users": {Name: "users"},
			},
			targetTables: map[string]Table{
				"users": {Name: "users"},
			},
			expectedDiffs: 0,
		},
		{
			name: "added table",
			sourceTables: map[string]Table{
				"users": {Name: "users"},
			},
			targetTables: map[string]Table{
				"users": {Name: "users"},
				"posts": {Name: "posts"},
			},
			expectedDiffs:  1,
			expectedTypes:  []string{"table"},
			expectedChange: []string{"add"},
		},
		{
			name: "removed table",
			sourceTables: map[string]Table{
				"users": {Name: "users"},
				"posts": {Name: "posts"},
			},
			targetTables: map[string]Table{
				"users": {Name: "users"},
			},
			expectedDiffs:  1,
			expectedTypes:  []string{"table"},
			expectedChange: []string{"remove"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparer := NewSchemaComparer(tt.sourceTables, tt.targetTables)
			diffs := comparer.Compare()

			if len(diffs) != tt.expectedDiffs {
				t.Errorf("Expected %d differences, got %d", tt.expectedDiffs, len(diffs))
			}

			if tt.expectedDiffs > 0 {
				for i, diff := range diffs {
					if diff.Type != tt.expectedTypes[i] {
						t.Errorf("Expected difference type %s, got %s", tt.expectedTypes[i], diff.Type)
					}
					if diff.ChangeType != tt.expectedChange[i] {
						t.Errorf("Expected change type %s, got %s", tt.expectedChange[i], diff.ChangeType)
					}
				}
			}
		})
	}
}

func TestCompareColumns(t *testing.T) {
	tests := []struct {
		name           string
		sourceColumns  map[string]Column
		targetColumns  map[string]Column
		expectedDiffs  int
		expectedTypes  []string
		expectedChange []string
	}{
		{
			name: "column type change",
			sourceColumns: map[string]Column{
				"id": {Name: "id", Type: "INT"},
			},
			targetColumns: map[string]Column{
				"id": {Name: "id", Type: "BIGINT"},
			},
			expectedDiffs:  1,
			expectedTypes:  []string{"column"},
			expectedChange: []string{"modify"},
		},
		{
			name: "added column",
			sourceColumns: map[string]Column{
				"id": {Name: "id", Type: "INT"},
			},
			targetColumns: map[string]Column{
				"id":    {Name: "id", Type: "INT"},
				"email": {Name: "email", Type: "VARCHAR"},
			},
			expectedDiffs:  1,
			expectedTypes:  []string{"column"},
			expectedChange: []string{"add"},
		},
		{
			name: "removed column",
			sourceColumns: map[string]Column{
				"id":    {Name: "id", Type: "INT"},
				"email": {Name: "email", Type: "VARCHAR"},
			},
			targetColumns: map[string]Column{
				"id": {Name: "id", Type: "INT"},
			},
			expectedDiffs:  1,
			expectedTypes:  []string{"column"},
			expectedChange: []string{"remove"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparer := NewSchemaComparer(
				map[string]Table{"users": {Name: "users", Columns: tt.sourceColumns}},
				map[string]Table{"users": {Name: "users", Columns: tt.targetColumns}},
			)
			diffs := comparer.Compare()

			if len(diffs) != tt.expectedDiffs {
				t.Errorf("Expected %d differences, got %d", tt.expectedDiffs, len(diffs))
			}

			if tt.expectedDiffs > 0 {
				for i, diff := range diffs {
					if diff.Type != tt.expectedTypes[i] {
						t.Errorf("Expected difference type %s, got %s", tt.expectedTypes[i], diff.Type)
					}
					if diff.ChangeType != tt.expectedChange[i] {
						t.Errorf("Expected change type %s, got %s", tt.expectedChange[i], diff.ChangeType)
					}
				}
			}
		})
	}
}

func TestCompareIndexes(t *testing.T) {
	tests := []struct {
		name           string
		sourceIndexes  map[string]Index
		targetIndexes  map[string]Index
		expectedDiffs  int
		expectedTypes  []string
		expectedChange []string
	}{
		{
			name: "index modification",
			sourceIndexes: map[string]Index{
				"idx_email": {Name: "idx_email", Columns: []string{"email"}},
			},
			targetIndexes: map[string]Index{
				"idx_email": {Name: "idx_email", Columns: []string{"email"}, Unique: true},
			},
			expectedDiffs:  1,
			expectedTypes:  []string{"index"},
			expectedChange: []string{"modify"},
		},
		{
			name: "added index",
			sourceIndexes: map[string]Index{
				"idx_email": {Name: "idx_email", Columns: []string{"email"}},
			},
			targetIndexes: map[string]Index{
				"idx_email": {Name: "idx_email", Columns: []string{"email"}},
				"idx_name":  {Name: "idx_name", Columns: []string{"name"}},
			},
			expectedDiffs:  1,
			expectedTypes:  []string{"index"},
			expectedChange: []string{"add"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparer := NewSchemaComparer(
				map[string]Table{"users": {Name: "users", Indexes: tt.sourceIndexes}},
				map[string]Table{"users": {Name: "users", Indexes: tt.targetIndexes}},
			)
			diffs := comparer.Compare()

			if len(diffs) != tt.expectedDiffs {
				t.Errorf("Expected %d differences, got %d", tt.expectedDiffs, len(diffs))
			}

			if tt.expectedDiffs > 0 {
				for i, diff := range diffs {
					if diff.Type != tt.expectedTypes[i] {
						t.Errorf("Expected difference type %s, got %s", tt.expectedTypes[i], diff.Type)
					}
					if diff.ChangeType != tt.expectedChange[i] {
						t.Errorf("Expected change type %s, got %s", tt.expectedChange[i], diff.ChangeType)
					}
				}
			}
		})
	}
}

func TestCompareConstraints(t *testing.T) {
	tests := []struct {
		name              string
		sourceConstraints map[string]Constraint
		targetConstraints map[string]Constraint
		expectedDiffs     int
		expectedTypes     []string
		expectedChange    []string
	}{
		{
			name: "constraint modification",
			sourceConstraints: map[string]Constraint{
				"fk_user": {
					Name:    "fk_user",
					Type:    "FOREIGN KEY",
					Columns: []string{"user_id"},
					References: &Reference{
						Table:   "users",
						Columns: []string{"id"},
					},
				},
			},
			targetConstraints: map[string]Constraint{
				"fk_user": {
					Name:    "fk_user",
					Type:    "FOREIGN KEY",
					Columns: []string{"user_id"},
					References: &Reference{
						Table:   "users",
						Columns: []string{"uuid"},
					},
				},
			},
			expectedDiffs:  1,
			expectedTypes:  []string{"constraint"},
			expectedChange: []string{"modify"},
		},
		{
			name:              "added constraint",
			sourceConstraints: map[string]Constraint{},
			targetConstraints: map[string]Constraint{
				"fk_user": {
					Name:    "fk_user",
					Type:    "FOREIGN KEY",
					Columns: []string{"user_id"},
					References: &Reference{
						Table:   "users",
						Columns: []string{"id"},
					},
				},
			},
			expectedDiffs:  1,
			expectedTypes:  []string{"constraint"},
			expectedChange: []string{"add"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparer := NewSchemaComparer(
				map[string]Table{"posts": {Name: "posts", Constraints: tt.sourceConstraints}},
				map[string]Table{"posts": {Name: "posts", Constraints: tt.targetConstraints}},
			)
			diffs := comparer.Compare()

			if len(diffs) != tt.expectedDiffs {
				t.Errorf("Expected %d differences, got %d", tt.expectedDiffs, len(diffs))
			}

			if tt.expectedDiffs > 0 {
				for i, diff := range diffs {
					if diff.Type != tt.expectedTypes[i] {
						t.Errorf("Expected difference type %s, got %s", tt.expectedTypes[i], diff.Type)
					}
					if diff.ChangeType != tt.expectedChange[i] {
						t.Errorf("Expected change type %s, got %s", tt.expectedChange[i], diff.ChangeType)
					}
				}
			}
		})
	}
}
