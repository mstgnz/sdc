package main

import (
	"fmt"

	"github.com/mstgnz/gosql/convert"
)

func main() {
	if statement, err := gosql.Cleaner("example/files/mysql.sql"); err != nil {
		fmt.Println("Error: ", err)
	} else {
		for _, stmt := range gosql.Parser(statement) {
			fmt.Println(stmt)
		}
	}
}
