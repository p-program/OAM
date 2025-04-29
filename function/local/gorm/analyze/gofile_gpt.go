package analyze

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type GPT struct {
}

func NewGPT() *GPT {
	return &GPT{}
}

func (g *GPT) Gofile(filePath string) string {
	return GofileGPT(filePath)
}

// GofileGPT 解析 Go 文件，分析 GORM 调用链并推断 SQL
func GofileGPT(filePath string) string {
	// prompt
	//用go写一个程序，实现一种代码静态分析，在不运行代码的前提下，把gorm的代码转换成实际的 sql。主体函数命名为main，文件名为main.go
	fmt.Println("分析文件:", filePath)
	srcBytes, _ := os.ReadFile(filePath)
	src := string(srcBytes)
	// 解析源代码
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, src, parser.AllErrors)
	if err != nil {
		panic(err)
	}

	// 遍历 AST
	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// 尝试解析 GORM 调用链
		sqlParts := []string{}
		analyzeCall(callExpr, &sqlParts)

		// 如果是 db.Model(...).Where(...) 这样的链式调用
		if len(sqlParts) > 0 {
			sqlParts = append([]string{"SELECT *"}, sqlParts...)
			sql := strings.Join(sqlParts, " ")
			fmt.Println("推断出的SQL:", sql)
		}
		return true
	})
	return ""
}

// 递归分析链式调用，支持更多GORM方法
func analyzeCall(expr ast.Expr, sqlParts *[]string) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return
	}

	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		// 递归处理调用链
		analyzeCall(fun.X, sqlParts)

		method := fun.Sel.Name
		switch method {
		case "Model":
			if len(call.Args) > 0 {
				switch arg := call.Args[0].(type) {
				case *ast.UnaryExpr:
					if starExpr, ok := arg.X.(*ast.CompositeLit); ok {
						if ident, ok := starExpr.Type.(*ast.Ident); ok {
							*sqlParts = append(*sqlParts, fmt.Sprintf("FROM %s", strings.ToLower(ident.Name)+"s"))
						}
					}
				case *ast.CompositeLit:
					if ident, ok := arg.Type.(*ast.Ident); ok {
						*sqlParts = append(*sqlParts, fmt.Sprintf("FROM %s", strings.ToLower(ident.Name)+"s"))
					}
				default:
					*sqlParts = append(*sqlParts, "FROM unknown_table")
				}
			}
		case "Where":
			if len(call.Args) > 0 {
				if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
					condition := strings.Trim(basicLit.Value, `"`)
					*sqlParts = append(*sqlParts, "WHERE", condition)
				}
			}
		case "Find":
			// No-op
		case "First":
			*sqlParts = append(*sqlParts, "LIMIT 1")
		case "Count":
			if len(call.Args) > 0 {
				if ident, ok := call.Args[0].(*ast.Ident); ok {
					// 根据传参生成 COUNT 别名
					*sqlParts = append([]string{"SELECT COUNT(*)"}, *sqlParts...)
					*sqlParts = append(*sqlParts, fmt.Sprintf("AS %s", ident.Name))
				}
			} else {
				*sqlParts = append([]string{"SELECT COUNT(*)"}, *sqlParts...)
			}
		case "Select":
			if len(call.Args) > 0 {
				if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
					columns := strings.Trim(basicLit.Value, `"`)
					if len(*sqlParts) > 0 && strings.HasPrefix((*sqlParts)[0], "SELECT") {
						(*sqlParts)[0] = fmt.Sprintf("SELECT %s", columns)
					} else {
						*sqlParts = append([]string{fmt.Sprintf("SELECT %s", columns)}, *sqlParts...)
					}
				}
			}
		case "Order":
			if len(call.Args) > 0 {
				if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
					order := strings.Trim(basicLit.Value, `"`)
					*sqlParts = append(*sqlParts, "ORDER BY", order)
				}
			}
		case "Limit":
			if len(call.Args) > 0 {
				if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
					limit := strings.Trim(basicLit.Value, `"`)
					*sqlParts = append(*sqlParts, "LIMIT", limit)
				}
			}
		case "Offset":
			if len(call.Args) > 0 {
				if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
					offset := strings.Trim(basicLit.Value, `"`)
					*sqlParts = append(*sqlParts, "OFFSET", offset)
				}
			}
		case "Group":
			if len(call.Args) > 0 {
				if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
					group := strings.Trim(basicLit.Value, `"`)
					*sqlParts = append(*sqlParts, "GROUP BY", group)
				}
			}
		case "Having":
			if len(call.Args) > 0 {
				if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
					having := strings.Trim(basicLit.Value, `"`)
					*sqlParts = append(*sqlParts, "HAVING", having)
				}
			}
		case "Joins":
			if len(call.Args) > 0 {
				if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
					join := strings.Trim(basicLit.Value, `"`)
					*sqlParts = append(*sqlParts, "JOIN", join)
				}
			}
		case "Preload":
			// Preload通常是预加载，静态分析可以忽略
		case "Scopes":
			if len(call.Args) > 0 {
				// 继续递归处理 Scope 内部的表达式（通常是函数调用）
				for _, arg := range call.Args {
					analyzeCall(arg, sqlParts)
				}
			}
		case "Distinct":
			if len(call.Args) > 0 {
				if basicLit, ok := call.Args[0].(*ast.BasicLit); ok {
					distinct := strings.Trim(basicLit.Value, `"`)
					*sqlParts = append([]string{fmt.Sprintf("SELECT DISTINCT %s", distinct)}, *sqlParts...)
				}
			}
		case "Delete":
			if len(call.Args) > 0 {
				*sqlParts = append([]string{"DELETE"}, *sqlParts...)
			} else {
				*sqlParts = append([]string{"DELETE"}, *sqlParts...)
			}
		case "Update":
			if len(call.Args) >= 2 {
				if col, ok1 := call.Args[0].(*ast.BasicLit); ok1 {
					if val, ok2 := call.Args[1].(*ast.BasicLit); ok2 {
						*sqlParts = append([]string{fmt.Sprintf("UPDATE SET %s = %s", strings.Trim(col.Value, `"`), strings.Trim(val.Value, `"`))}, *sqlParts...)
					}
				}
			}
		case "Updates":
			*sqlParts = append([]string{"UPDATE SET ..."}, *sqlParts...) // 复杂情况简化处理
		}
	}
}
