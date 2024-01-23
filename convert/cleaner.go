package gosql

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

func Cleaner(fileName string) (string, error) {

	content, err := readFile(fileName)
	if err != nil {
		return "", err
	}

	return removeSQLComments(content), nil
}

func readFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	var content strings.Builder

	for scanner.Scan() {
		content.WriteString(scanner.Text() + "\n")
	}

	return content.String(), scanner.Err()
}

func removeSQLComments(content string) string {
	// Clear comment lines starting with '--'
	content = regexp.MustCompile(`--.*?\n`).ReplaceAllString(content, "")

	// Clear multiple lines in '/* */'
	content = regexp.MustCompile(`/\*(.|\n)*?\*/`).ReplaceAllString(content, "")

	// Clear empty lines
	content = regexp.MustCompile(`(?m)^\s*\n`).ReplaceAllString(content, "")

	// Clear excess gaps
	content = regexp.MustCompile(`\s{2,}`).ReplaceAllString(content, " ")

	return content
}
