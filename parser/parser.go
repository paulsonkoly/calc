package parser

import "github.com/phaul/calc/lexer"

type ASTNode struct {}

func Parse(input string) (*ASTNode, error) {
  l := lexer.NewLexer(input)
  r, err := statement(l)
  return r, err
}

func statement(l lexer.Lexer) (*ASTNode, error) {
  if !l.Next() {
    return nil, errors.New("Parser: expression expected")
  }
  if l.Err != nil {
    return nil, l.Err
  } 
  if !l.Next()
  if.
  if l.Err != nil {
    return &ASTNode{}, l.Err
  }
  return &ASTNode{}, nil
