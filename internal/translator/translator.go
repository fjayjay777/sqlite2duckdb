package translator

import (
	"fmt"
	"io"
	"os"

	"sql-translator/internal/parser"

	"github.com/antlr4-go/antlr/v4"
)

type SQLiteTranslator struct {
	input string
	core  translatorCore
}

func NewSQLiteTranslator(input string) *SQLiteTranslator {
	return &SQLiteTranslator{
        input: input,
        core: translatorCore{
            BaseSQLiteParserVisitor: &parser.BaseSQLiteParserVisitor{},
        },
    }	
}

func (t *SQLiteTranslator) Translate() string {
	tree, _ := t.getSyntaxTree()
	return t.core.Visit(tree).(string)
}

func (t *SQLiteTranslator) getSyntaxTree() (antlr.ParseTree, *parser.SQLiteParser) {
	input := antlr.NewInputStream(t.input)
	lexer := parser.NewSQLiteLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewSQLiteParser(stream)
	tree := p.Parse()
	return tree, p
}

func (t *SQLiteTranslator) ShowSyntaxTree() {
	t.WriteSyntaxTree(os.Stdout)
}

func (t *SQLiteTranslator) WriteSyntaxTree(w io.Writer) {
	tree, p := t.getSyntaxTree()
	fmt.Fprintln(w, tree.ToStringTree(nil, p))
}

