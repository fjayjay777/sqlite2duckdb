package translator

import (
	"testing"

	"sql-translator/internal/parser"

	"github.com/antlr4-go/antlr/v4"
)

func TestSimpleSelect(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "select star",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "select columns",
			input:    "SELECT id, name FROM users",
			expected: "SELECT id, name FROM users",
		},
		{
			name:     "select columns with alias",
			input:    "SELECT id AS id, name AS name FROM users",
			expected: "SELECT id AS id, name AS name FROM users",
		},
		{
			name:     "select columns with part alias",
			input:    "SELECT id, name AS name FROM users",
			expected: "SELECT id, name AS name FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := translatorCore{
				BaseSQLiteParserVisitor: &parser.BaseSQLiteParserVisitor{},
			}

			tree := createParseTree(tt.input)
			got := core.Visit(tree).(string)

			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTranslatorCore_SelectWithWhere(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple where equals",
			input:    "SELECT * FROM users WHERE id = 1",
			expected: "SELECT * FROM users WHERE id = 1",
		},
		{
			name:     "where with and",
			input:    "SELECT name, email FROM users WHERE age > 18 AND active = true",
			expected: "SELECT name, email FROM users WHERE age > 18 AND active = true",
		},
		{
			name:     "where with multiple conditions",
			input:    "SELECT id, name FROM employees WHERE department = 'IT' AND salary >= 50000 AND age < 50",
			expected: "SELECT id, name FROM employees WHERE department = 'IT' AND salary >= 50000 AND age < 50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core := translatorCore{
				BaseSQLiteParserVisitor: &parser.BaseSQLiteParserVisitor{},
			}

			tree := createParseTree(tt.input)
			got := core.Visit(tree).(string)

			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func createParseTree(input string) antlr.ParseTree {
	inputStream := antlr.NewInputStream(input)
	lexer := parser.NewSQLiteLexer(inputStream)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewSQLiteParser(stream)
	return p.Parse()
}
