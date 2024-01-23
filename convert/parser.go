package gosql

import (
	"bufio"
	"strings"
)

func Parser(sqlDump string) []string {
	var statement []string
	var sqlStatement strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(sqlDump))

	for scanner.Scan() {
		stmt := scanner.Text()

		if strings.HasPrefix(stmt, "CREATE") {
			sqlStatement.Reset()
		}

		sqlStatement.WriteString(stmt)

		if strings.HasSuffix(stmt, ";") {
			statement = append(statement, sqlStatement.String())
			sqlStatement.Reset()
		}
	}
	return statement
}
