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

func TestSelectWithJoin(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "inner join",
			input:    "SELECT users.name, orders.amount FROM users INNER JOIN orders ON users.id = orders.user_id",
			expected: "SELECT users.name, orders.amount FROM users JOIN orders ON users.id = orders.user_id",
		},
		{
			name:     "left outer join",
			input:    "SELECT users.*, orders.order_date FROM users LEFT OUTER JOIN orders ON users.id = orders.user_id",
			expected: "SELECT users.*, orders.order_date FROM users LEFT OUTER JOIN orders ON users.id = orders.user_id",
		},
		{
			name:     "left join without outer keyword",
			input:    "SELECT users.*, orders.order_date FROM users LEFT JOIN orders ON users.id = orders.user_id",
			expected: "SELECT users.*, orders.order_date FROM users LEFT JOIN orders ON users.id = orders.user_id",
		},
		{
			name:     "cross join",
			input:    "SELECT products.*, categories.name FROM products CROSS JOIN categories",
			expected: "SELECT products.*, categories.name FROM products CROSS JOIN categories",
		},
		{
			name:     "natural join",
			input:    "SELECT * FROM employees NATURAL JOIN departments",
			expected: "SELECT * FROM employees NATURAL JOIN departments",
		},
		{
			name:     "join with using clause",
			input:    "SELECT * FROM orders JOIN order_items USING (order_id)",
			expected: "SELECT * FROM orders JOIN order_items USING (order_id)",
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

func TestSelectWithSubqueries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "subquery in FROM",
			input:    "SELECT name FROM (SELECT * FROM users WHERE age > 18) AS adult_users",
			expected: "SELECT name FROM (SELECT * FROM users WHERE age > 18) AS adult_users",
		},
		{
			name:     "subquery in WHERE with IN",
			input:    "SELECT * FROM products WHERE category_id IN (SELECT id FROM categories WHERE active = true)",
			expected: "SELECT * FROM products WHERE category_id IN (SELECT id FROM categories WHERE active = true)",
		},
		{
			name:     "subquery in WHERE with comparison",
			input:    "SELECT * FROM employees WHERE salary > (SELECT AVG(salary) FROM employees)",
			expected: "SELECT * FROM employees WHERE salary > (SELECT AVG(salary) FROM employees)",
		},
		{
			name:     "subquery in SELECT",
			input:    "SELECT name, (SELECT COUNT(*) FROM orders WHERE orders.user_id = users.id) AS order_count FROM users",
			expected: "SELECT name, (SELECT COUNT(*) FROM orders WHERE orders.user_id = users.id) AS order_count FROM users",
		},
		{
			name:     "EXISTS subquery",
			input:    "SELECT * FROM products WHERE EXISTS (SELECT 1 FROM inventory WHERE inventory.product_id = products.id)",
			expected: "SELECT * FROM products WHERE EXISTS (SELECT 1 FROM inventory WHERE inventory.product_id = products.id)",
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
