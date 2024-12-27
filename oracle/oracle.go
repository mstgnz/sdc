package oracle

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mstgnz/sqlporter"
)

type Oracle struct {
	schema *sqlporter.Schema
}

// NewOracle creates a new Oracle parser instance
func NewOracle() *Oracle {
	return &Oracle{
		schema: &sqlporter.Schema{},
	}
}

// Parse parses Oracle dump content
func (o *Oracle) Parse(content string) (*sqlporter.Schema, error) {
	if content == "" {
		return nil, errors.New("empty content")
	}

	// SQL ifadelerini ayır
	var statements []string
	var currentStmt strings.Builder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "/" {
			if currentStmt.Len() > 0 {
				statements = append(statements, currentStmt.String())
				currentStmt.Reset()
			}
			continue
		}

		if strings.HasSuffix(line, ";") {
			currentStmt.WriteString(line[:len(line)-1])
			if currentStmt.Len() > 0 {
				statements = append(statements, currentStmt.String())
				currentStmt.Reset()
			}
		} else {
			if currentStmt.Len() > 0 {
				currentStmt.WriteString(" ")
			}
			currentStmt.WriteString(line)
		}
	}

	if currentStmt.Len() > 0 {
		statements = append(statements, currentStmt.String())
	}

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// CREATE TABLE
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE TABLE") {
			table, err := o.parseCreateTable(stmt)
			if err != nil {
				return nil, err
			}
			o.schema.Tables = append(o.schema.Tables, table)
		}

		// CREATE SEQUENCE
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE SEQUENCE") {
			seq, err := o.parseCreateSequence(stmt)
			if err != nil {
				return nil, err
			}
			o.schema.Sequences = append(o.schema.Sequences, seq)
		}

		// CREATE VIEW
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE") && strings.Contains(strings.ToUpper(stmt), "VIEW") {
			view, err := o.parseCreateView(stmt)
			if err != nil {
				return nil, err
			}
			o.schema.Views = append(o.schema.Views, view)
		}

		// CREATE TRIGGER
		if strings.HasPrefix(strings.ToUpper(stmt), "CREATE") && strings.Contains(strings.ToUpper(stmt), "TRIGGER") {
			trigger, err := o.parseCreateTrigger(stmt)
			if err != nil {
				return nil, err
			}
			o.schema.Triggers = append(o.schema.Triggers, trigger)
		}
	}

	return o.schema, nil
}

// parseCreateTable parses CREATE TABLE statement
func (o *Oracle) parseCreateTable(stmt string) (sqlporter.Table, error) {
	table := sqlporter.Table{}

	// Tablo adını al
	tableNameRegex := regexp.MustCompile(`CREATE\s+TABLE\s+(\w+)`)
	matches := tableNameRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		table.Name = matches[1]
	}

	// Kolonları parse et
	columnsStr := stmt[strings.Index(stmt, "(")+1 : strings.LastIndex(stmt, ")")]
	columnDefs := strings.Split(columnsStr, ",")

	for _, colDef := range columnDefs {
		colDef = strings.TrimSpace(colDef)
		if strings.HasPrefix(colDef, "CONSTRAINT") {
			constraint := sqlporter.Constraint{}

			// Constraint adını al
			nameRegex := regexp.MustCompile(`CONSTRAINT\s+(\w+)`)
			matches := nameRegex.FindStringSubmatch(colDef)
			if len(matches) > 1 {
				constraint.Name = matches[1]
			}

			if strings.Contains(colDef, "PRIMARY KEY") {
				constraint.Type = "PRIMARY KEY"
				// Kolonları al
				colsRegex := regexp.MustCompile(`PRIMARY\s+KEY\s*\(([^)]+)\)`)
				matches = colsRegex.FindStringSubmatch(colDef)
				if len(matches) > 1 {
					cols := strings.Split(matches[1], ",")
					for i, col := range cols {
						cols[i] = strings.TrimSpace(col)
					}
					constraint.Columns = cols
				}
			} else if strings.Contains(colDef, "FOREIGN KEY") {
				constraint.Type = "FOREIGN KEY"
				// FK kolonlarını al
				fkRegex := regexp.MustCompile(`FOREIGN\s+KEY\s*\(([^)]+)\)`)
				matches = fkRegex.FindStringSubmatch(colDef)
				if len(matches) > 1 {
					cols := strings.Split(matches[1], ",")
					for i, col := range cols {
						cols[i] = strings.TrimSpace(col)
					}
					constraint.Columns = cols
				}
				// Referans tabloyu ve kolonları al
				refRegex := regexp.MustCompile(`REFERENCES\s+(\w+)\s*\(([^)]+)\)`)
				matches = refRegex.FindStringSubmatch(colDef)
				if len(matches) > 2 {
					constraint.RefTable = matches[1]
					refCols := strings.Split(matches[2], ",")
					for i, col := range refCols {
						refCols[i] = strings.TrimSpace(col)
					}
					constraint.RefColumns = refCols
				}
				// ON DELETE kuralını al
				if strings.Contains(colDef, "ON DELETE") {
					if strings.Contains(colDef, "CASCADE") {
						constraint.DeleteRule = "CASCADE"
					}
				}
			} else if strings.Contains(colDef, "UNIQUE") {
				constraint.Type = "UNIQUE"
				// Kolonları al
				colsRegex := regexp.MustCompile(`UNIQUE\s*\(([^)]+)\)`)
				matches = colsRegex.FindStringSubmatch(colDef)
				if len(matches) > 1 {
					cols := strings.Split(matches[1], ",")
					for i, col := range cols {
						cols[i] = strings.TrimSpace(col)
					}
					constraint.Columns = cols
				}
			} else if strings.Contains(colDef, "CHECK") {
				constraint.Type = "CHECK"
				// Check ifadesini al
				checkRegex := regexp.MustCompile(`CHECK\s*\(([^)]+)\)`)
				matches = checkRegex.FindStringSubmatch(colDef)
				if len(matches) > 1 {
					constraint.CheckExpression = strings.TrimSpace(matches[1])
				}
			}
			table.Constraints = append(table.Constraints, constraint)
			continue
		}

		parts := strings.Fields(colDef)
		if len(parts) < 2 {
			continue
		}

		col := sqlporter.Column{
			Name:     parts[0],
			DataType: parts[1],
		}

		if strings.Contains(colDef, "NOT NULL") {
			col.IsNullable = false
		}

		if strings.Contains(colDef, "DEFAULT") {
			defaultIdx := strings.Index(strings.ToUpper(colDef), "DEFAULT")
			restStr := colDef[defaultIdx+7:]
			defaultEnd := strings.Index(restStr, " ")
			if defaultEnd == -1 {
				defaultEnd = len(restStr)
			}
			col.DefaultValue = strings.TrimSpace(restStr[:defaultEnd])
		}

		if strings.Contains(colDef, "PRIMARY KEY") {
			col.IsPrimaryKey = true
			constraint := sqlporter.Constraint{
				Type:    "PRIMARY KEY",
				Columns: []string{col.Name},
			}
			table.Constraints = append(table.Constraints, constraint)
		}

		if strings.Contains(colDef, "UNIQUE") {
			col.IsUnique = true
			constraint := sqlporter.Constraint{
				Type:    "UNIQUE",
				Columns: []string{col.Name},
			}
			table.Constraints = append(table.Constraints, constraint)
		}

		if strings.Contains(colDef, "CHECK") {
			checkRegex := regexp.MustCompile(`CHECK\s*\(([^)]+)\)`)
			matches := checkRegex.FindStringSubmatch(colDef)
			if len(matches) > 1 {
				constraint := sqlporter.Constraint{
					Type:            "CHECK",
					CheckExpression: strings.TrimSpace(matches[1]),
				}
				table.Constraints = append(table.Constraints, constraint)
			}
		}

		table.Columns = append(table.Columns, col)
	}

	return table, nil
}

// parseCreateSequence parses CREATE SEQUENCE statement
func (o *Oracle) parseCreateSequence(stmt string) (sqlporter.Sequence, error) {
	seq := sqlporter.Sequence{}

	// Sequence adını al
	seqNameRegex := regexp.MustCompile(`CREATE\s+SEQUENCE\s+(\w+)`)
	matches := seqNameRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		seq.Name = matches[1]
	}

	// START WITH değerini al
	startWithRegex := regexp.MustCompile(`START\s+WITH\s+(\d+)`)
	matches = startWithRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		seq.StartValue = 1 // Default değer
	}

	// INCREMENT BY değerini al
	incrementByRegex := regexp.MustCompile(`INCREMENT\s+BY\s+(\d+)`)
	matches = incrementByRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		seq.IncrementBy = 1 // Default değer
	}

	return seq, nil
}

// parseCreateView parses CREATE VIEW statement
func (o *Oracle) parseCreateView(stmt string) (sqlporter.View, error) {
	view := sqlporter.View{}

	// View adını al
	viewNameRegex := regexp.MustCompile(`CREATE\s+(?:OR\s+REPLACE\s+)?VIEW\s+(\w+)`)
	matches := viewNameRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		view.Name = matches[1]
	}

	// View tanımını al
	asIndex := strings.Index(strings.ToUpper(stmt), " AS ")
	if asIndex != -1 {
		view.Definition = strings.TrimSpace(stmt[asIndex+4:])
	}

	return view, nil
}

// parseCreateTrigger parses CREATE TRIGGER statement
func (o *Oracle) parseCreateTrigger(stmt string) (sqlporter.Trigger, error) {
	trigger := sqlporter.Trigger{}

	// Trigger adını al
	triggerNameRegex := regexp.MustCompile(`CREATE\s+(?:OR\s+REPLACE\s+)?TRIGGER\s+(\w+)`)
	matches := triggerNameRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		trigger.Name = matches[1]
	}

	// Trigger zamanlamasını al
	if strings.Contains(strings.ToUpper(stmt), "BEFORE") {
		trigger.Timing = "BEFORE"
	} else if strings.Contains(strings.ToUpper(stmt), "AFTER") {
		trigger.Timing = "AFTER"
	}

	// Trigger olayını al
	if strings.Contains(strings.ToUpper(stmt), "INSERT") {
		trigger.Event = "INSERT"
	} else if strings.Contains(strings.ToUpper(stmt), "UPDATE") {
		trigger.Event = "UPDATE"
	} else if strings.Contains(strings.ToUpper(stmt), "DELETE") {
		trigger.Event = "DELETE"
	}

	// Tablo adını al
	tableRegex := regexp.MustCompile(`ON\s+(\w+)`)
	matches = tableRegex.FindStringSubmatch(stmt)
	if len(matches) > 1 {
		trigger.Table = matches[1]
	}

	// FOR EACH ROW kontrolü
	trigger.ForEachRow = strings.Contains(strings.ToUpper(stmt), "FOR EACH ROW")

	// Trigger gövdesini al
	beginIndex := strings.Index(strings.ToUpper(stmt), "BEGIN")
	endIndex := strings.LastIndex(strings.ToUpper(stmt), "END")
	if beginIndex != -1 && endIndex != -1 {
		trigger.Body = strings.TrimSpace(stmt[beginIndex : endIndex+3])
	}

	return trigger, nil
}

// Generate generates Oracle dump from schema
func (o *Oracle) Generate(schema *sqlporter.Schema) (string, error) {
	if schema == nil {
		return "", errors.New("empty schema")
	}

	var result strings.Builder

	// Sequence'leri oluştur
	for _, seq := range schema.Sequences {
		result.WriteString(fmt.Sprintf("CREATE SEQUENCE %s START WITH %d INCREMENT BY %d;\n\n",
			seq.Name, seq.StartValue, seq.IncrementBy))
	}

	// Tabloları oluştur
	for _, table := range schema.Tables {
		result.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table.Name))

		// Kolonları ekle
		for i, col := range table.Columns {
			result.WriteString(fmt.Sprintf("    %s %s", col.Name, col.DataType))
			if col.Length > 0 {
				if col.Scale > 0 {
					result.WriteString(fmt.Sprintf("(%d,%d)", col.Length, col.Scale))
				} else {
					result.WriteString(fmt.Sprintf("(%d)", col.Length))
				}
			}
			if col.IsPrimaryKey {
				result.WriteString(" PRIMARY KEY")
			} else if col.IsNullable == false {
				result.WriteString(" NOT NULL")
			}
			if col.DefaultValue != "" {
				// String tipindeki default değerler için tırnak işareti ekle
				if strings.HasPrefix(col.DataType, "VARCHAR") || strings.HasPrefix(col.DataType, "CHAR") {
					result.WriteString(fmt.Sprintf(" DEFAULT '%s'", col.DefaultValue))
				} else {
					result.WriteString(fmt.Sprintf(" DEFAULT %s", col.DefaultValue))
				}
			}
			if col.IsUnique && !col.IsPrimaryKey {
				result.WriteString(" UNIQUE")
			}
			if i < len(table.Columns)-1 || len(table.Constraints) > 0 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}

		// Constraint'leri ekle
		for i, constraint := range table.Constraints {
			if constraint.Name == "" {
				continue // Skip unnamed constraints as they are handled with column definitions
			}
			result.WriteString(fmt.Sprintf("    CONSTRAINT %s %s", constraint.Name, constraint.Type))
			if len(constraint.Columns) > 0 {
				result.WriteString(fmt.Sprintf(" (%s)", strings.Join(constraint.Columns, ", ")))
			}
			if constraint.Type == "FOREIGN KEY" && constraint.RefTable != "" {
				result.WriteString(fmt.Sprintf(" REFERENCES %s", constraint.RefTable))
				if len(constraint.RefColumns) > 0 {
					result.WriteString(fmt.Sprintf("(%s)", strings.Join(constraint.RefColumns, ", ")))
				}
				if constraint.DeleteRule != "" {
					result.WriteString(fmt.Sprintf(" ON DELETE %s", constraint.DeleteRule))
				}
			}
			if i < len(table.Constraints)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}

		result.WriteString(");\n")

		// Index'leri oluştur
		for _, index := range table.Indexes {
			if index.IsUnique {
				result.WriteString(fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s(%s);\n",
					index.Name, table.Name, strings.Join(index.Columns, ", ")))
			} else {
				result.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s(%s);\n",
					index.Name, table.Name, strings.Join(index.Columns, ", ")))
			}
		}

		result.WriteString("\n")
	}

	// View'ları oluştur
	for _, view := range schema.Views {
		result.WriteString(fmt.Sprintf("CREATE OR REPLACE VIEW %s AS\n%s;\n\n",
			view.Name, view.Definition))
	}

	// Trigger'ları oluştur
	for _, trigger := range schema.Triggers {
		result.WriteString(fmt.Sprintf("CREATE OR REPLACE TRIGGER %s\n", trigger.Name))
		if trigger.Timing != "" {
			result.WriteString(trigger.Timing + " ")
		}
		if trigger.Event != "" {
			result.WriteString(trigger.Event + " ")
		}
		if trigger.Table != "" {
			result.WriteString("ON " + trigger.Table + " ")
		}
		if trigger.ForEachRow {
			result.WriteString("FOR EACH ROW\n")
		}
		if trigger.Body != "" {
			result.WriteString(trigger.Body)
		}
		result.WriteString("\n/\n\n")
	}

	return result.String(), nil
}
