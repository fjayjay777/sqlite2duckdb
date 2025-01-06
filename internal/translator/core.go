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
	var fromClause string

	// Handle joins if they exist
	if join := ctx.Join_clause(); join != nil {
		fromClause = c.Visit(join).(string)
	} else {
		// Get the base table/subquery only if no joins
		fromClause = c.Visit(ctx.Table_or_subquery(0)).(string)
	}

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
	if ctx.Table_name() != nil {
		tableName := ctx.Table_name().GetText()
		// Remove any join type keywords that might be concatenated
		tableName = strings.TrimSuffix(tableName, "INNER")
		tableName = strings.TrimSuffix(tableName, "LEFT")
		tableName = strings.TrimSuffix(tableName, "CROSS")
		tableName = strings.TrimSuffix(tableName, "NATURAL")
		return tableName
	}

	if ctx.Select_stmt() != nil {
		subquery := c.Visit(ctx.Select_stmt())
		if ctx.Table_alias() != nil {
			return fmt.Sprintf("(%s) AS %s", subquery, ctx.Table_alias().GetText())
		}
		return fmt.Sprintf("(%s)", subquery)
	}

	return strings.TrimSuffix(ctx.GetText(), "INNER")
}

func (c *translatorCore) VisitResult_column(ctx *parser.Result_columnContext) any {
	if ctx.STAR() != nil {
		if table := ctx.Table_name(); table != nil {
			return fmt.Sprintf("%s.*", table.GetText())
		}
		return "*"
	}

	if ctx.Column_alias() != nil {
		expr := c.Visit(ctx.Expr())
		alias := ctx.Column_alias().GetText()
		return fmt.Sprintf("%s AS %s", expr, alias)
	}

	if ctx.Expr() != nil {
		return c.Visit(ctx.Expr())
	}

	return ctx.GetText()
}

func (c *translatorCore) VisitExpr(ctx *parser.ExprContext) interface{} {
	if ctx == nil {
		return nil
	}

	if ctx.EXISTS_() != nil {
		return fmt.Sprintf("EXISTS (%s)", c.Visit(ctx.Select_stmt()))
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

func (c *translatorCore) VisitJoin_clause(ctx *parser.Join_clauseContext) any {
	if ctx == nil {
		return ""
	}

	result := c.Visit(ctx.Table_or_subquery(0)).(string)

	for i := 0; i < len(ctx.AllJoin_operator()); i++ {
		joinOp := ctx.Join_operator(i)
		table := c.Visit(ctx.Table_or_subquery(i + 1)).(string)

		// Debug join operator details
		fmt.Printf("\nDetailed join operator debug for index %d:\n", i)
		fmt.Printf("Full text: '%s'\n", joinOp.GetText())
		fmt.Printf("Child count: %d\n", joinOp.GetChildCount())
		for j := 0; j < joinOp.GetChildCount(); j++ {
			child := joinOp.GetChild(j).(antlr.ParseTree)
			fmt.Printf("Child %d: '%s' (type: %T)\n", j, child.GetText(), child)
		}
		fmt.Printf("NATURAL_(): %v\n", joinOp.NATURAL_() != nil)
		fmt.Printf("LEFT_(): %v\n", joinOp.LEFT_() != nil)
		fmt.Printf("CROSS_(): %v\n", joinOp.CROSS_() != nil)
		fmt.Printf("OUTER_(): %v\n", joinOp.OUTER_() != nil)

		// Check first table for join type
		firstTable := ctx.Table_or_subquery(0)
		fmt.Printf("First table text: '%s'\n", firstTable.GetText())

		// Determine join type from both the join operator and table text
		var joinType string
		tableText := firstTable.GetText()

		if strings.Contains(tableText, "NATURAL") {
			joinType = "NATURAL JOIN"
		} else if strings.Contains(tableText, "CROSS") {
			joinType = "CROSS JOIN"
		} else if joinOp.LEFT_() != nil || strings.Contains(tableText, "LEFT") {
			if joinOp.OUTER_() != nil {
				joinType = "LEFT OUTER JOIN"
			} else {
				joinType = "LEFT JOIN"
			}
		} else {
			joinType = "JOIN"
		}

		fmt.Printf("Selected join type: '%s'\n", joinType)

		var condition string
		if constraint := ctx.Join_constraint(i); constraint != nil {
			condition = c.Visit(constraint).(string)
		}
		result = fmt.Sprintf("%s %s %s %s", result, joinType, table, strings.TrimSpace(condition))
	}

	return strings.TrimSpace(result)
}

func (c *translatorCore) VisitJoin_constraint(ctx *parser.Join_constraintContext) any {
	if ctx.ON_() != nil {
		return fmt.Sprintf("ON %s", c.Visit(ctx.Expr()))
	}
	if ctx.USING_() != nil {
		return fmt.Sprintf("USING %s", ctx.GetText()[5:]) // Skip "USING" keyword
	}
	return ""
}
