package main

import (
	"fmt"

	"sql-translator/internal/translator"
)

func main() {
	query := "SELECT name, id FROM users"
	translator := translator.NewSQLiteTranslator(query)
	translator.ShowSyntaxTree()

	res := translator.Translate()
	fmt.Printf("Original: %s\nTranslated: %s\n", query, res)
}
