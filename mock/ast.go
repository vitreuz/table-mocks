package mock

import (
	"go/ast"
	"go/token"
)

func call(fn ast.Expr, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{Fun: fn, Args: args}
}

func compositeLit(name string, deref bool) *ast.UnaryExpr {
	expr := &ast.UnaryExpr{X: &ast.CompositeLit{Type: ast.NewIdent(name)}}
	if deref {
		expr.Op = token.AND
	}
	return expr
}

func expression(expr ...ast.Expr) []ast.Expr {
	return expr
}

func field(typ ast.Expr, names ...string) *ast.Field {
	expr := &ast.Field{Type: typ}
	if len(names) == 0 {
		return expr
	}
	for _, name := range names {
		expr.Names = append(expr.Names, ast.NewIdent(name))
	}
	return expr
}

func fieldList(fields ...*ast.Field) *ast.FieldList {
	return &ast.FieldList{List: fields}
}

func funcDecl(recv *ast.Field, name string, params, results *ast.FieldList, body *ast.BlockStmt) *ast.FuncDecl {
	decl := &ast.FuncDecl{
		Name: ast.NewIdent(name),
		Type: &ast.FuncType{Params: params, Results: results},
		Body: body,
	}
	if recv != nil {
		decl.Recv = &ast.FieldList{List: []*ast.Field{recv}}
	}

	return decl
}

func selectorExpr(x ast.Expr, sel string) *ast.SelectorExpr {
	return &ast.SelectorExpr{X: x, Sel: ast.NewIdent(sel)}
}

func simpleMap(key, value string) *ast.MapType {
	return &ast.MapType{Key: ast.NewIdent(key), Value: ast.NewIdent(value)}
}

func starExpr(name string) *ast.StarExpr {
	return &ast.StarExpr{X: ast.NewIdent(name)}
}

func blockStmt(stmts ...ast.Stmt) *ast.BlockStmt {
	return &ast.BlockStmt{List: stmts}
}

func valueIndex(x ast.Expr, val string) *ast.IndexExpr {
	return &ast.IndexExpr{X: x, Index: &ast.BasicLit{Value: val}}
}
