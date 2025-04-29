package analyze

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"zeusro.com/gotemplate/function/local/gorm/util"
)

// 定义SQL操作类型
type SQLOperation int

const (
	SELECT SQLOperation = iota
	UPDATE
	DELETE
	INSERT
)

// SQL构建器结构体，跟踪SQL生成状态
type SQLBuilder struct {
	Operation SQLOperation
	Table     string
	Columns   string
	Where     []string
	OrderBy   []string
	GroupBy   []string
	Having    []string
	Joins     []string
	Limit     string
	Offset    string
	Updates   string
}

// 初始化一个默认为SELECT的SQLBuilder
func NewSQLBuilder() *SQLBuilder {
	return &SQLBuilder{
		Operation: SELECT,
		Columns:   "*",
	}
}

// 生成SQL字符串
func (sb *SQLBuilder) BuildSQL() string {
	var parts []string

	// 根据操作类型生成SQL前缀
	switch sb.Operation {
	case SELECT:
		if strings.Contains(sb.Columns, "DISTINCT") {
			parts = append(parts, sb.Columns)
		} else {
			parts = append(parts, "SELECT "+sb.Columns)
		}
	case UPDATE:
		parts = append(parts, "UPDATE "+sb.Table+" SET "+sb.Updates)
		// 在UPDATE后不需要添加表名，因为已经包含了
		sb.Table = ""
	case DELETE:
		parts = append(parts, "DELETE")
	case INSERT:
		parts = append(parts, "INSERT INTO "+sb.Table)
		// 在INSERT后不需要添加表名，因为已经包含了
		sb.Table = ""
	}

	// 添加表名（如果有且需要）
	if sb.Table != "" {
		parts = append(parts, "FROM "+sb.Table)
	}

	// 添加其他SQL部分
	if len(sb.Joins) > 0 {
		parts = append(parts, strings.Join(sb.Joins, " "))
	}

	if len(sb.Where) > 0 {
		parts = append(parts, "WHERE "+strings.Join(sb.Where, " AND "))
	}

	if len(sb.GroupBy) > 0 {
		parts = append(parts, "GROUP BY "+strings.Join(sb.GroupBy, ", "))
	}

	if len(sb.Having) > 0 {
		parts = append(parts, "HAVING "+strings.Join(sb.Having, " AND "))
	}

	if len(sb.OrderBy) > 0 {
		parts = append(parts, "ORDER BY "+strings.Join(sb.OrderBy, ", "))
	}

	if sb.Limit != "" {
		parts = append(parts, "LIMIT "+sb.Limit)
	}

	if sb.Offset != "" {
		parts = append(parts, "OFFSET "+sb.Offset)
	}

	return strings.Join(parts, " ")
}

type Meta struct {
}

func NewMeta() *Meta {
	return &Meta{}
}

func (g *Meta) Gofile(filePath string) string {
	return GofileMeta(filePath)
}

func GofileMeta(filePath string) string {
	fmt.Println("分析文件:", filePath)
	srcBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Sprintf("读取文件错误: %v", err)
	}

	src := string(srcBytes)

	// 解析源代码
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, src, parser.AllErrors)
	if err != nil {
		return fmt.Sprintf("解析错误: %v", err)
	}

	results := map[int]string{}

	// 收集SQL分析结果
	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// 初始化SQL构建器
		sqlBuilder := NewSQLBuilder()

		// 分析GORM调用链
		if analyzeGormChain(callExpr, sqlBuilder) {
			sql := sqlBuilder.BuildSQL()
			if sql != "" {
				// 获取源代码位置信息
				pos := fset.Position(n.Pos())
				if v, contains := results[pos.Line]; !contains || len(sql) > len(v) {
					results[pos.Line] = fmt.Sprintf("%s:%d: %s", filePath, pos.Line, sql)
				}
			}
		}

		return true
	})

	// 输出结果
	if len(results) > 0 {
		a := []int{}
		for k, _ := range results {
			a = append(a, k)
		}
		a = util.SortInts(a, true)
		for _, v := range a {
			fmt.Println(results[v])
		}

		return fmt.Sprintf("已分析 %d 条SQL语句", len(results))
	}

	return "未发现GORM查询语句"
}

// 分析GORM调用链
func analyzeGormChain(expr ast.Expr, builder *SQLBuilder) bool {
	// 不是函数调用，无需处理
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	// 检查是否是选择器表达式 (例如 db.Find)
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// 获取方法名
	methodName := selector.Sel.Name

	// 处理GORM方法
	switch methodName {
	case "Model":
		if len(call.Args) > 0 {
			tableName := extractTableName(call.Args[0])
			if tableName != "" {
				builder.Table = tableName
			}
		}
	case "Table":
		if len(call.Args) > 0 {
			if strLit, ok := call.Args[0].(*ast.BasicLit); ok {
				builder.Table = strings.Trim(strLit.Value, `"'`)
			}
		}
	case "Where", "And", "Or":
		if len(call.Args) > 0 {
			condition := extractStringArg(call.Args[0])
			if condition != "" {
				builder.Where = append(builder.Where, condition)
			}
		}
	case "Select":
		if len(call.Args) > 0 {
			columns := extractStringArg(call.Args[0])
			if columns != "" {
				builder.Columns = columns
			}
		}
	case "Order":
		if len(call.Args) > 0 {
			order := extractStringArg(call.Args[0])
			if order != "" {
				builder.OrderBy = append(builder.OrderBy, order)
			}
		}
	case "Group":
		if len(call.Args) > 0 {
			group := extractStringArg(call.Args[0])
			if group != "" {
				builder.GroupBy = append(builder.GroupBy, group)
			}
		}
	case "Having":
		if len(call.Args) > 0 {
			having := extractStringArg(call.Args[0])
			if having != "" {
				builder.Having = append(builder.Having, having)
			}
		}
	case "Limit":
		if len(call.Args) > 0 {
			limit := extractLimitOffset(call.Args[0])
			if limit != "" {
				builder.Limit = limit
			}
		}
	case "Offset":
		if len(call.Args) > 0 {
			offset := extractLimitOffset(call.Args[0])
			if offset != "" {
				builder.Offset = offset
			}
		}
	case "Joins", "InnerJoins", "LeftJoins", "RightJoins":
		if len(call.Args) > 0 {
			join := extractStringArg(call.Args[0])
			joinType := "JOIN"
			switch methodName {
			case "InnerJoins":
				joinType = "INNER JOIN"
			case "LeftJoins":
				joinType = "LEFT JOIN"
			case "RightJoins":
				joinType = "RIGHT JOIN"
			}
			if join != "" {
				builder.Joins = append(builder.Joins, joinType+" "+join)
			}
		}
	case "First":
		builder.Limit = "1"
	case "Take":
		builder.Limit = "1"
	case "Last":
		if len(builder.OrderBy) == 0 {
			builder.OrderBy = append(builder.OrderBy, "id DESC")
		}
		builder.Limit = "1"
	case "Find":
		// Find只是执行查询，不影响SQL结构
	case "Count":
		builder.Columns = "COUNT(*)"
		if len(call.Args) > 0 {
			if ident, ok := call.Args[0].(*ast.Ident); ok {
				builder.Columns += fmt.Sprintf(" AS %s", ident.Name)
			}
		}
	case "Distinct":
		if len(call.Args) > 0 {
			distinct := extractStringArg(call.Args[0])
			if distinct != "" {
				builder.Columns = fmt.Sprintf("SELECT DISTINCT %s", distinct)
			} else {
				builder.Columns = "DISTINCT " + builder.Columns
			}
		} else {
			builder.Columns = "DISTINCT " + builder.Columns
		}
	case "Delete":
		builder.Operation = DELETE
	case "Update":
		builder.Operation = UPDATE
		if len(call.Args) >= 2 {
			col := extractStringArg(call.Args[0])
			val := extractStringArg(call.Args[1])
			if col != "" && val != "" {
				builder.Updates = fmt.Sprintf("%s = %s", col, val)
			}
		}
	case "Updates":
		builder.Operation = UPDATE
		builder.Updates = extractUpdatesContent(call.Args)
	case "Create", "CreateInBatches":
		builder.Operation = INSERT
		// 简化处理
	}

	// 递归处理前一个链式调用
	if sel, ok := selector.X.(*ast.CallExpr); ok {
		analyzeGormChain(sel, builder)
	}

	return true
}

// 提取表名
func extractTableName(expr ast.Expr) string {
	switch arg := expr.(type) {
	case *ast.UnaryExpr: // 处理 &User{} 形式
		if starExpr, ok := arg.X.(*ast.CompositeLit); ok {
			return getTypeNameFromCompositeLit(starExpr)
		}
	case *ast.CompositeLit: // 处理 User{} 形式
		return getTypeNameFromCompositeLit(arg)
	case *ast.Ident: // 处理变量引用形式
		return strings.ToLower(arg.Name) + "s" // 简单处理，假定是复数形式
	}
	return ""
}

// 从复合字面量获取类型名
func getTypeNameFromCompositeLit(lit *ast.CompositeLit) string {
	switch typeExpr := lit.Type.(type) {
	case *ast.Ident:
		// 转换为snake_case并复数形式，如User -> users
		return pluralize(camelToSnake(typeExpr.Name))
	case *ast.SelectorExpr:
		// if ident, ok := typeExpr.Sel.(*ast.Ident); ok {
		// 	return pluralize(camelToSnake(ident.Name))
		// }
		return pluralize(camelToSnake(typeExpr.Sel.Name))

	}
	return ""
}

// 提取字符串参数
func extractStringArg(expr ast.Expr) string {
	switch arg := expr.(type) {
	case *ast.BasicLit:
		return strings.Trim(arg.Value, `"'`)
	case *ast.Ident:
		// 如果是标识符，可能是变量名，这里简化处理
		return arg.Name
	}
	return ""
}

// 提取LIMIT或OFFSET值
func extractLimitOffset(expr ast.Expr) string {
	switch arg := expr.(type) {
	case *ast.BasicLit:
		return strings.Trim(arg.Value, `"'`)
	case *ast.Ident:
		return arg.Name
	case *ast.CallExpr: // 处理函数调用，简化返回
		return "?"
	}
	return ""
}

// 提取Updates内容
func extractUpdatesContent(args []ast.Expr) string {
	if len(args) == 0 {
		return "..."
	}

	// 处理map形式的Updates
	if len(args) == 1 {
		if comp, ok := args[0].(*ast.CompositeLit); ok {
			if _, mapType := comp.Type.(*ast.MapType); mapType {
				pairs := []string{}
				for i := 0; i < len(comp.Elts); i += 2 {
					if i+1 < len(comp.Elts) {
						if key, ok := comp.Elts[i].(*ast.KeyValueExpr); ok {
							if keyLit, ok := key.Key.(*ast.BasicLit); ok {
								keyStr := strings.Trim(keyLit.Value, `"'`)
								valStr := "?"
								if valLit, ok := key.Value.(*ast.BasicLit); ok {
									valStr = strings.Trim(valLit.Value, `"'`)
								}
								pairs = append(pairs, fmt.Sprintf("%s = %s", keyStr, valStr))
							}
						}
					}
				}
				if len(pairs) > 0 {
					return strings.Join(pairs, ", ")
				}
			}
		}
	}

	return "..."
}

// 将驼峰命名转换为蛇形命名
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// 简单复数转换
func pluralize(s string) string {
	// 简化处理，只是加s
	return s + "s"
}
