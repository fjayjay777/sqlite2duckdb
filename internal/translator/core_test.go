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

func TestSelectWithWhere(t *testing.T) {
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

func TestSelectWithOrderBy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple order by",
			input:    "SELECT * FROM users ORDER BY name",
			expected: "SELECT * FROM users ORDER BY name",
		},
		{
			name:     "order by with direction",
			input:    "SELECT * FROM users ORDER BY name DESC",
			expected: "SELECT * FROM users ORDER BY name DESC",
		},
		{
			name:     "order by multiple columns",
			input:    "SELECT id, name, age FROM users ORDER BY age DESC, name ASC",
			expected: "SELECT id, name, age FROM users ORDER BY age DESC, name ASC",
		},
		{
			name:     "order by with where clause",
			input:    "SELECT * FROM users WHERE age > 18 ORDER BY name",
			expected: "SELECT * FROM users WHERE age > 18 ORDER BY name",
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

func TestSelectWithGroupBy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple group by",
			input:    "SELECT department, COUNT(*) FROM employees GROUP BY department",
			expected: "SELECT department, COUNT(*) FROM employees GROUP BY department",
		},
		{
			name:     "group by with multiple columns",
			input:    "SELECT department, role, COUNT(*) FROM employees GROUP BY department, role",
			expected: "SELECT department, role, COUNT(*) FROM employees GROUP BY department, role",
		},
		{
			name:     "group by with where",
			input:    "SELECT department, AVG(salary) FROM employees WHERE salary > 50000 GROUP BY department",
			expected: "SELECT department, AVG(salary) FROM employees WHERE salary > 50000 GROUP BY department",
		},
		{
			name:     "group by with order by",
			input:    "SELECT department, COUNT(*) FROM employees GROUP BY department ORDER BY COUNT(*) DESC",
			expected: "SELECT department, COUNT(*) FROM employees GROUP BY department ORDER BY COUNT(*) DESC",
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

func TestSelectWithLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple limit",
			input:    "SELECT * FROM users LIMIT 5",
			expected: "SELECT * FROM users LIMIT 5",
		},
		{
			name:     "limit with offset clause",
			input:    "SELECT * FROM users LIMIT 5 OFFSET 10",
			expected: "SELECT * FROM users LIMIT 5 OFFSET 10",
		},
		{
			name:     "limit with offset using comma",
			input:    "SELECT * FROM users LIMIT 10, 5",
			expected: "SELECT * FROM users LIMIT 5 OFFSET 10",
		},
		{
			name:     "limit with where and order",
			input:    "SELECT * FROM users WHERE age > 18 ORDER BY name LIMIT 5 OFFSET 10",
			expected: "SELECT * FROM users WHERE age > 18 ORDER BY name LIMIT 5 OFFSET 10",
		},
		{
			name:     "limit with where and order using comma",
			input:    "SELECT * FROM users WHERE age > 18 ORDER BY name LIMIT 10, 5",
			expected: "SELECT * FROM users WHERE age > 18 ORDER BY name LIMIT 5 OFFSET 10",
		},
		{
			name:     "limit with expressions",
			input:    "SELECT * FROM users LIMIT 5 * 2 OFFSET 10 + 5",
			expected: "SELECT * FROM users LIMIT 5 * 2 OFFSET 10 + 5",
		},
		{
			name:     "limit with subquery",
			input:    "SELECT * FROM users LIMIT (SELECT count FROM settings)",
			expected: "SELECT * FROM users LIMIT (SELECT count FROM settings)",
		},
		{
			name:     "limit zero",
			input:    "SELECT * FROM users LIMIT 0",
			expected: "SELECT * FROM users LIMIT 0",
		},
		{
			name:     "limit with parameter",
			input:    "SELECT * FROM users LIMIT ?",
			expected: "SELECT * FROM users LIMIT ?",
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
