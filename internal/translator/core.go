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

	query := c.Visit(selectCore).(string)

	if orderBy := ctx.Order_by_stmt(); orderBy != nil {
		query = fmt.Sprintf("%s %s", query, c.Visit(orderBy))
	}

	if limit := ctx.Limit_stmt(); limit != nil {
		query = fmt.Sprintf("%s %s", query, c.Visit(limit))
	}

	return query
}

func (c *translatorCore) VisitSelect_core(ctx *parser.Select_coreContext) any {
	fromClause := c.Visit(ctx.Table_or_subquery(0)).(string)
	columnStr := c.buildColumns(ctx)
	whereClause := c.buildWhereClause(ctx)
	groupByClause := c.buildGroupByClause(ctx)

	query := fmt.Sprintf("SELECT %s FROM %s", columnStr, fromClause)

	if whereClause != "" {
		query = fmt.Sprintf("%s %s", query, whereClause)
	}

	if groupByClause != "" {
		query = fmt.Sprintf("%s %s", query, groupByClause)
	}

	return query
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

func (c *translatorCore) VisitExpr(ctx *parser.ExprContext) interface{} {
	if ctx == nil {
		return nil
	}

	if ctx.Select_stmt() != nil {
		return fmt.Sprintf("(%s)", c.Visit(ctx.Select_stmt()))
	}

	exprs := ctx.AllExpr()

	if len(exprs) == 2 {
		leftExpr := c.Visit(exprs[0]).(string)
		rightExpr := c.Visit(exprs[1]).(string)

		operator := ctx.GetChild(1).(antlr.TerminalNode).GetSymbol().GetText()

		return fmt.Sprintf("%s %s %s", leftExpr, operator, rightExpr)
	}

	return ctx.GetText()
}

func (c *translatorCore) VisitOrder_by_stmt(ctx *parser.Order_by_stmtContext) interface{} {
	var orderClauses []string

	for i := 0; i < len(ctx.AllOrdering_term()); i++ {
		term := ctx.Ordering_term(i)
		expr := term.Expr().GetText()

		// if direction exists (ASC/DESC)
		if term.GetChildCount() > 1 {
			direction := term.GetChild(1).(antlr.ParseTree).GetText()
			orderClauses = append(orderClauses, fmt.Sprintf("%s %s", expr, direction))
		} else {
			orderClauses = append(orderClauses, expr)
		}
	}

	return fmt.Sprintf("ORDER BY %s", strings.Join(orderClauses, ", "))
}

func (c *translatorCore) VisitLimit_stmt(ctx *parser.Limit_stmtContext) any {
	if ctx == nil {
		return ""
	}

	exprs := ctx.AllExpr()
	if len(exprs) == 1 {
		return fmt.Sprintf("LIMIT %s", c.Visit(exprs[0]))
	}

	if ctx.OFFSET_() != nil {
		return fmt.Sprintf("LIMIT %s OFFSET %s", c.Visit(exprs[0]), c.Visit(exprs[1]))
	}
	return fmt.Sprintf("LIMIT %s OFFSET %s", c.Visit(exprs[1]), c.Visit(exprs[0]))
}

func (c *translatorCore) buildColumns(ctx *parser.Select_coreContext) string {
	var columns []string
	for i := 0; i < len(ctx.AllResult_column()); i++ {
		col := c.Visit(ctx.Result_column(i)).(string)
		columns = append(columns, col)
	}

	return strings.Join(columns, ", ")
}

func (c *translatorCore) buildWhereClause(ctx *parser.Select_coreContext) string {
	if whereExpr := ctx.GetWhereExpr(); whereExpr != nil {
		return fmt.Sprintf("WHERE %s", c.Visit(whereExpr))
	}
	return ""
}

func (c *translatorCore) buildGroupByClause(ctx *parser.Select_coreContext) string {
	if groupByExprs := ctx.GetGroupByExpr(); len(groupByExprs) > 0 {
		var columns []string
		for _, expr := range groupByExprs {
			columns = append(columns, expr.GetText())
		}
		return fmt.Sprintf("GROUP BY %s", strings.Join(columns, ", "))
	}
	return ""
}
