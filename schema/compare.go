package schema

import (
	"fmt"
)

// Difference represents a schema difference
type Difference struct {
	Type        string // "table", "column", "index", "constraint"
	Name        string
	ChangeType  string // "add", "modify", "remove"
	Source      interface{}
	Target      interface{}
	Description string
}

// SchemaComparer compares database schemas
type SchemaComparer struct {
	sourceTables map[string]Table
	targetTables map[string]Table
}

// Table represents a database table structure
type Table struct {
	Name        string
	Columns     map[string]Column
	Indexes     map[string]Index
	Constraints map[string]Constraint
}

// Column represents a table column
type Column struct {
	Name       string
	Type       string
	Nullable   bool
	Default    interface{}
	Properties map[string]interface{}
}

// Index represents a table index
type Index struct {
	Name    string
	Columns []string
	Type    string
	Unique  bool
}

// Constraint represents a table constraint
type Constraint struct {
	Name       string
	Type       string
	Columns    []string
	References *Reference
}

// Reference represents a foreign key reference
type Reference struct {
	Table   string
	Columns []string
}

// NewSchemaComparer creates a new schema comparer
func NewSchemaComparer(sourceTables, targetTables map[string]Table) *SchemaComparer {
	return &SchemaComparer{
		sourceTables: sourceTables,
		targetTables: targetTables,
	}
}

// Compare compares two schemas and returns the differences
func (sc *SchemaComparer) Compare() []Difference {
	var differences []Difference

	// Compare tables
	differences = append(differences, sc.compareTables()...)

	// Compare table contents for tables that exist in both schemas
	for tableName, sourceTable := range sc.sourceTables {
		if targetTable, exists := sc.targetTables[tableName]; exists {
			differences = append(differences, sc.compareColumns(tableName, sourceTable.Columns, targetTable.Columns)...)
			differences = append(differences, sc.compareIndexes(tableName, sourceTable.Indexes, targetTable.Indexes)...)
			differences = append(differences, sc.compareConstraints(tableName, sourceTable.Constraints, targetTable.Constraints)...)
		}
	}

	return differences
}

// compareTables compares table existence between schemas
func (sc *SchemaComparer) compareTables() []Difference {
	var differences []Difference

	// Check for removed tables
	for tableName := range sc.sourceTables {
		if _, exists := sc.targetTables[tableName]; !exists {
			differences = append(differences, Difference{
				Type:        "table",
				Name:        tableName,
				ChangeType:  "remove",
				Source:      sc.sourceTables[tableName],
				Description: fmt.Sprintf("Table %s has been removed", tableName),
			})
		}
	}

	// Check for new tables
	for tableName := range sc.targetTables {
		if _, exists := sc.sourceTables[tableName]; !exists {
			differences = append(differences, Difference{
				Type:        "table",
				Name:        tableName,
				ChangeType:  "add",
				Target:      sc.targetTables[tableName],
				Description: fmt.Sprintf("Table %s has been added", tableName),
			})
		}
	}

	return differences
}

// compareColumns compares columns between two tables
func (sc *SchemaComparer) compareColumns(tableName string, sourceColumns, targetColumns map[string]Column) []Difference {
	var differences []Difference

	for colName, sourceCol := range sourceColumns {
		if targetCol, exists := targetColumns[colName]; exists {
			if !columnsEqual(sourceCol, targetCol) {
				differences = append(differences, Difference{
					Type:        "column",
					Name:        fmt.Sprintf("%s.%s", tableName, colName),
					ChangeType:  "modify",
					Source:      sourceCol,
					Target:      targetCol,
					Description: fmt.Sprintf("Column %s in table %s has been modified", colName, tableName),
				})
			}
		} else {
			differences = append(differences, Difference{
				Type:        "column",
				Name:        fmt.Sprintf("%s.%s", tableName, colName),
				ChangeType:  "remove",
				Source:      sourceCol,
				Description: fmt.Sprintf("Column %s has been removed from table %s", colName, tableName),
			})
		}
	}

	for colName, targetCol := range targetColumns {
		if _, exists := sourceColumns[colName]; !exists {
			differences = append(differences, Difference{
				Type:        "column",
				Name:        fmt.Sprintf("%s.%s", tableName, colName),
				ChangeType:  "add",
				Target:      targetCol,
				Description: fmt.Sprintf("Column %s has been added to table %s", colName, tableName),
			})
		}
	}

	return differences
}

// compareIndexes compares indexes between two tables
func (sc *SchemaComparer) compareIndexes(tableName string, sourceIndexes, targetIndexes map[string]Index) []Difference {
	var differences []Difference

	for idxName, sourceIdx := range sourceIndexes {
		if targetIdx, exists := targetIndexes[idxName]; exists {
			if !indexesEqual(sourceIdx, targetIdx) {
				differences = append(differences, Difference{
					Type:        "index",
					Name:        fmt.Sprintf("%s.%s", tableName, idxName),
					ChangeType:  "modify",
					Source:      sourceIdx,
					Target:      targetIdx,
					Description: fmt.Sprintf("Index %s in table %s has been modified", idxName, tableName),
				})
			}
		} else {
			differences = append(differences, Difference{
				Type:        "index",
				Name:        fmt.Sprintf("%s.%s", tableName, idxName),
				ChangeType:  "remove",
				Source:      sourceIdx,
				Description: fmt.Sprintf("Index %s has been removed from table %s", idxName, tableName),
			})
		}
	}

	for idxName, targetIdx := range targetIndexes {
		if _, exists := sourceIndexes[idxName]; !exists {
			differences = append(differences, Difference{
				Type:        "index",
				Name:        fmt.Sprintf("%s.%s", tableName, idxName),
				ChangeType:  "add",
				Target:      targetIdx,
				Description: fmt.Sprintf("Index %s has been added to table %s", idxName, tableName),
			})
		}
	}

	return differences
}

// compareConstraints compares constraints between two tables
func (sc *SchemaComparer) compareConstraints(tableName string, sourceConstraints, targetConstraints map[string]Constraint) []Difference {
	var differences []Difference

	for constName, sourceCons := range sourceConstraints {
		if targetCons, exists := targetConstraints[constName]; exists {
			if !constraintsEqual(sourceCons, targetCons) {
				differences = append(differences, Difference{
					Type:        "constraint",
					Name:        fmt.Sprintf("%s.%s", tableName, constName),
					ChangeType:  "modify",
					Source:      sourceCons,
					Target:      targetCons,
					Description: fmt.Sprintf("Constraint %s in table %s has been modified", constName, tableName),
				})
			}
		} else {
			differences = append(differences, Difference{
				Type:        "constraint",
				Name:        fmt.Sprintf("%s.%s", tableName, constName),
				ChangeType:  "remove",
				Source:      sourceCons,
				Description: fmt.Sprintf("Constraint %s has been removed from table %s", constName, tableName),
			})
		}
	}

	for constName, targetCons := range targetConstraints {
		if _, exists := sourceConstraints[constName]; !exists {
			differences = append(differences, Difference{
				Type:        "constraint",
				Name:        fmt.Sprintf("%s.%s", tableName, constName),
				ChangeType:  "add",
				Target:      targetCons,
				Description: fmt.Sprintf("Constraint %s has been added to table %s", constName, tableName),
			})
		}
	}

	return differences
}

// Helper functions for comparing schema objects
func columnsEqual(a, b Column) bool {
	return a.Type == b.Type &&
		a.Nullable == b.Nullable &&
		a.Default == b.Default
}

func indexesEqual(a, b Index) bool {
	if len(a.Columns) != len(b.Columns) {
		return false
	}
	for i := range a.Columns {
		if a.Columns[i] != b.Columns[i] {
			return false
		}
	}
	return a.Type == b.Type && a.Unique == b.Unique
}

func constraintsEqual(a, b Constraint) bool {
	if a.Type != b.Type || len(a.Columns) != len(b.Columns) {
		return false
	}
	for i := range a.Columns {
		if a.Columns[i] != b.Columns[i] {
			return false
		}
	}

	// Check references
	if (a.References == nil) != (b.References == nil) {
		return false
	}
	if a.References != nil {
		if a.References.Table != b.References.Table {
			return false
		}
		if len(a.References.Columns) != len(b.References.Columns) {
			return false
		}
		for i := range a.References.Columns {
			if a.References.Columns[i] != b.References.Columns[i] {
				return false
			}
		}
	}

	return true
}
