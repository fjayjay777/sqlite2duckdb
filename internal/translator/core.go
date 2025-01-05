package translator

import (
	"fmt"
	"strings"

	"sql-translator/internal/parser"

	"github.com/antlr4-go/antlr/v4"
)

type translatorCore struct {
	*parser.BaseSQLiteParserVisitor
}

func (c *translatorCore) Visit(tree antlr.ParseTree) any {
	if tree == nil {
		return ""
	}
	return tree.Accept(c)
}

func (c *translatorCore) VisitParse(ctx *parser.ParseContext) any {
	return c.Visit(ctx.Sql_stmt_list(0))
}

func (c *translatorCore) VisitSql_stmt_list(ctx *parser.Sql_stmt_listContext) any {
	return c.Visit(ctx.Sql_stmt(0))
}

func (c *translatorCore) VisitSql_stmt(ctx *parser.Sql_stmtContext) any {
	return c.Visit(ctx.Select_stmt())
}

func (c *translatorCore) VisitSelect_stmt(ctx *parser.Select_stmtContext) any {
	selectCore := ctx.Select_core(0)
	if selectCore == nil {
		return ""
	}
	return c.Visit(selectCore)
}

func (c *translatorCore) VisitSelect_core(ctx *parser.Select_coreContext) any {
	var columns []string
	for i := 0; i < len(ctx.AllResult_column()); i++ {
		col := c.Visit(ctx.Result_column(i)).(string)
		columns = append(columns, col)
	}

	columnStr := strings.Join(columns, ", ")

	fromClause := c.Visit(ctx.Table_or_subquery(0)).(string)
	return fmt.Sprintf("SELECT %s FROM %s", columnStr, fromClause)
}

func (c *translatorCore) VisitTable_or_subquery(ctx *parser.Table_or_subqueryContext) any {
	tableName := ctx.Table_name().GetText()
	return tableName
}

func (c *translatorCore) VisitResult_column(ctx *parser.Result_columnContext) any {
	if ctx.STAR() != nil {
		return "*"
	}

	if ctx.Column_alias() != nil {
		expr := ctx.Expr().GetText()
		alias := ctx.Column_alias().GetText()
		return fmt.Sprintf("%s AS %s", expr, alias)
	}

	return ctx.GetText()
}
