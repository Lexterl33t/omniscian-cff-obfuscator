package obfuscator

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"math/rand"
	"os"
	"time"

	"github.com/Lexterl33t/omniscian-obfuscator-cfg/utils"
	"golang.org/x/tools/go/ast/astutil"
)

func FindForStatement(block *ast.BlockStmt) *ast.ForStmt {

	for _, stmt := range block.List {
		switch x := stmt.(type) {
		case *ast.ForStmt:
			return x
		}
	}

	return nil
}

func CreateForStmt(exit string, statelabel string) ast.ForStmt {

	return ast.ForStmt{
		Cond: &ast.BinaryExpr{
			X: &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("\"%x\"", exit),
			},
			Y: &ast.Ident{
				Name: statelabel,
			},
			Op: token.NEQ,
		},
		Body: &ast.BlockStmt{},
	}
}

func CreateSwitchStmt(state string) ast.SwitchStmt {
	return ast.SwitchStmt{
		Tag:  ast.NewIdent("state"),
		Body: &ast.BlockStmt{},
	}
}

func CreateCaseStmt(exit string) ast.CaseClause {
	return ast.CaseClause{
		List: []ast.Expr{
			&ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("\"%x\"", exit),
			},
		},
		Body: []ast.Stmt{},
	}
}

func CreateAssignVariable(labelName string, value string) ast.AssignStmt {
	return ast.AssignStmt{
		Lhs: []ast.Expr{
			ast.NewIdent(labelName),
		},
		Rhs: []ast.Expr{
			&ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("\"%x\"", value),
			},
		},
		Tok: token.ASSIGN,
	}
}

func CreateBreak() ast.BranchStmt {
	return ast.BranchStmt{
		Tok: token.BREAK,
	}
}

func CreateCondition(cond ast.Expr) ast.IfStmt {
	return ast.IfStmt{
		Cond: cond,
		Body: &ast.BlockStmt{},
	}
}

func CreateDeclStmt(varname string) ast.DeclStmt {
	return ast.DeclStmt{
		Decl: &ast.GenDecl{
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Type: &ast.Ident{
						Name: "string",
					},
					Names: []*ast.Ident{
						ast.NewIdent(varname),
					},
				},
			},
			Tok: token.VAR,
		},
	}
}

func TransformBlock(stmts []ast.Stmt, switchstmt *ast.SwitchStmt, exit *string, funcdecl *ast.FuncDecl) {

	for _, stmt := range stmts {

		switch x := stmt.(type) {
		case *ast.BlockStmt:
			TransformBlock(x.List, switchstmt, exit, funcdecl)
		case *ast.IfStmt:
			TransformIfBlock(stmt, switchstmt, exit, funcdecl)
		case *ast.ExprStmt:
			TransformExpressionStmt(stmt, switchstmt, exit)
		case *ast.AssignStmt:
			TransformAssignStmt(stmt, switchstmt, exit, funcdecl)
		case *ast.ReturnStmt:
			TransformReturnStmt(stmt, switchstmt, exit, funcdecl)
		}
	}
}

func TransformReturnStmt(stmt ast.Stmt, switchstmt *ast.SwitchStmt, exit *string, funcdecl *ast.FuncDecl) {

	var clauseCase ast.CaseClause = CreateCaseStmt(*exit)

	*exit = utils.GenerateStateVariable()

	var breakStatement ast.BranchStmt = ast.BranchStmt{
		Tok: token.BREAK,
	}

	clauseCase.Body = append(clauseCase.Body, stmt, &breakStatement)

	switchstmt.Body.List = append(switchstmt.Body.List, &clauseCase)

}

func TransformAssignStmt(stmt ast.Stmt, switchstmt *ast.SwitchStmt, exit *string, funcdeclBody *ast.FuncDecl) {
	var caseClause ast.CaseClause = CreateCaseStmt(*exit)

	*exit = utils.GenerateStateVariable()

	funcdeclBody.Body.List = append(funcdeclBody.Body.List, stmt)
	var assignVarNext ast.AssignStmt = CreateAssignVariable("state", *exit)

	caseClause.Body = append(caseClause.Body, &ast.AssignStmt{
		Lhs: stmt.(*ast.AssignStmt).Lhs,
		Rhs: stmt.(*ast.AssignStmt).Rhs,
		Tok: token.ASSIGN,
	})
	caseClause.Body = append(caseClause.Body, &assignVarNext)

	switchstmt.Body.List = append(switchstmt.Body.List, &caseClause)
}

func TransformExpressionStmt(stmt ast.Stmt, switchstmt *ast.SwitchStmt, exit *string) {

	var clauseCase ast.CaseClause = CreateCaseStmt(*exit)

	var currstmt *ast.ExprStmt = stmt.(*ast.ExprStmt)

	clauseCase.Body = append(clauseCase.Body, currstmt)

	*exit = utils.GenerateStateVariable()

	var nextVarAssignState ast.AssignStmt = CreateAssignVariable("state", *exit)

	clauseCase.Body = append(clauseCase.Body, &nextVarAssignState)

	switchstmt.Body.List = append(switchstmt.Body.List, &clauseCase)
}

func TransformIfBlock(stmt ast.Stmt, switchstmt *ast.SwitchStmt, exit *string, funcdecl *ast.FuncDecl) {

	var currstmt *ast.IfStmt = stmt.(*ast.IfStmt)
	var condition ast.IfStmt = CreateCondition(currstmt.Cond)

	var clauseCase ast.CaseClause = CreateCaseStmt(*exit)

	if currstmt.Else != nil {
		*exit = utils.GenerateStateVariable()
		var assignVar ast.AssignStmt = CreateAssignVariable("state", *exit)
		condition.Else = &assignVar
	}

	*exit = utils.GenerateStateVariable()

	var assignVarBodyCond ast.AssignStmt = CreateAssignVariable("state", *exit)
	condition.Body.List = append(condition.Body.List, &assignVarBodyCond)

	clauseCase.Body = append(clauseCase.Body, &condition)

	switchstmt.Body.List = append(switchstmt.Body.List, &clauseCase)
	TransformBlock(currstmt.Body.List, switchstmt, exit, funcdecl)

	if currstmt.Else != nil {
		TransformBlock(currstmt.Else.(*ast.BlockStmt).List, switchstmt, exit, funcdecl)
	}
}

func FlattenBlockFuncDecl(funcDecl *ast.FuncDecl) *ast.BlockStmt {
	var funcDeclRet *ast.FuncDecl = &ast.FuncDecl{
		Body: &ast.BlockStmt{},
	}

	var exit string = utils.GenerateStateVariable()

	var entryVarDec ast.DeclStmt = CreateDeclStmt("state")
	var entryStmt ast.AssignStmt = CreateAssignVariable("state", exit)

	funcDeclRet.Body.List = append(funcDeclRet.Body.List, &entryVarDec, &entryStmt)

	var forStmt ast.ForStmt = CreateForStmt(exit, "state")

	var switchstmt ast.SwitchStmt = CreateSwitchStmt("state")

	TransformBlock(funcDecl.Body.List, &switchstmt, &exit, funcDeclRet)

	forStmt.Body.List = append(forStmt.Body.List, &switchstmt)

	forStmt.Cond.(*ast.BinaryExpr).X.(*ast.BasicLit).Value = fmt.Sprintf("\"%x\"", exit)

	funcDeclRet.Body.List = append(funcDeclRet.Body.List, &forStmt)

	return funcDeclRet.Body

}

func FindSwitchByFlattenBlock(flattenblock *ast.BlockStmt) (*ast.SwitchStmt, bool) {

	for _, stmt := range flattenblock.List {
		switch x := stmt.(type) {
		case *ast.ForStmt:
			for _, st := range x.Body.List {
				switch c := st.(type) {
				case *ast.SwitchStmt:
					return c, true
				}
			}
		}
	}

	return nil, false
}

func CFG(content string) error {
	prog_ast, fs, err := utils.StringToAst(content)
	if err != nil {
		return err
	}

	astutil.Apply(prog_ast, nil, func(c *astutil.Cursor) bool {
		n := c.Node()

		switch x := n.(type) {
		case *ast.FuncDecl:

			var flattenBlock *ast.BlockStmt = FlattenBlockFuncDecl(x)
			rand.Seed(time.Now().Unix())

			switchstmt, ok := FindSwitchByFlattenBlock(flattenBlock)
			if ok {
				rand.Shuffle(len(switchstmt.Body.List), func(i, j int) {
					switchstmt.Body.List[i], switchstmt.Body.List[j] = switchstmt.Body.List[j], switchstmt.Body.List[i]
				})

				x.Body = flattenBlock
			}

			if x.Type.Results != nil {
				if len(x.Type.Results.List) > 0 {
					flattenBlock.List = append(flattenBlock.List, &ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{
								Name: "n",
							},
						},
					})
				}
			}
			x.Body = flattenBlock

		}

		return true
	})

	printer.Fprint(os.Stdout, fs, prog_ast)
	return nil
}
