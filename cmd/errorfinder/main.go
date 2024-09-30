// Binary errorfinder extracts error sentinel and structured error value types
// from source code at the named import paths or directories (absolute
// file system paths).
package main

import (
	"cmp"
	"encoding/csv"
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"iter"
	"log"
	"os"
	"slices"

	"golang.org/x/tools/go/packages"
)

var errorInterface = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

func isErrorType(t types.Type) bool {
	return types.Implements(t, errorInterface)
}

//go:generate stringer -type=ErrorType
type errorType int

const (
	errorTypeUnknown errorType = iota
	errorTypeSentinel
	errorTypeStructured
)

//go:generate stringer -type=ExportType
type exportType int

const (
	exportTypeUnknown exportType = iota
	exportTypeExported
	exportTypeUnexported
)

type def struct {
	errorType
	exportType
	ImportPath      string
	PackageName     string
	Name            string
	BackingTypeName string
}

const escapes = "" // Convenient code formatting with Markdown.

func (d def) Write(enc *csv.Writer) error {
	data := []string{
		d.errorType.String(),
		d.exportType.String(),
		escapes + d.ImportPath + escapes,
		d.PackageName,
		escapes + d.Name + escapes,
		d.BackingTypeName,
	}
	return enc.Write(data)
}

func compareDef(a, b def) int {
	switch v := cmp.Compare(a.errorType, b.errorType); v {
	case -1, 1:
		return v
	}
	switch v := cmp.Compare(a.exportType, b.exportType); v {
	case -1, 1:
		return v
	}
	switch v := cmp.Compare(a.ImportPath, b.ImportPath); v {
	case -1, 1:
		return v
	}
	switch v := cmp.Compare(a.PackageName, b.PackageName); v {
	case -1, 1:
		return v
	}
	switch v := cmp.Compare(a.Name, b.Name); v {
	case -1, 1:
		return v
	}
	return cmp.Compare(a.BackingTypeName, b.BackingTypeName)
}

type searchTree struct {
	Decl ast.Decl
	Info *types.Info
	Pkg  *packages.Package
}

func topLevelDecls(pkgs []*packages.Package) iter.Seq[searchTree] {
	return func(yield func(searchTree) bool) {
		for _, pkg := range pkgs {
			for _, file := range pkg.Syntax {
				for _, decl := range file.Decls {
					if !yield(searchTree{decl, pkg.TypesInfo, pkg}) {
						return
					}
				}
			}
		}
	}
}

func expType(id *ast.Ident) exportType {
	if ast.IsExported(id.Name) {
		return exportTypeExported
	}
	return exportTypeUnexported
}

func extractSentinels(tree searchTree) iter.Seq[def] {
	return func(yield func(def) bool) {
		genDecl, ok := tree.Decl.(*ast.GenDecl)
		if !ok {
			return
		}
		for _, s := range genDecl.Specs {
			valueSpec, ok := s.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, n := range valueSpec.Names {
				if !isErrorType(tree.Info.TypeOf(n)) {
					continue
				}
				def := def{
					errorType:       errorTypeSentinel,
					exportType:      expType(n),
					ImportPath:      tree.Pkg.PkgPath,
					PackageName:     tree.Pkg.Name,
					Name:            n.Name,
					BackingTypeName: tree.Info.Defs[n].Type().String(),
				}
				if !yield(def) {
					return
				}
			}
		}
	}
}

func extractStructured(tree searchTree) iter.Seq[def] {
	return func(yield func(def) bool) {
		genDecl, ok := tree.Decl.(*ast.GenDecl)
		if !ok {
			return
		}
		for _, s := range genDecl.Specs {
			typeSpec, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if !isErrorType(tree.Info.TypeOf(typeSpec.Name)) {
				continue
			}
			def := def{
				errorType:       errorTypeStructured,
				exportType:      expType(typeSpec.Name),
				ImportPath:      tree.Pkg.PkgPath,
				PackageName:     tree.Pkg.Name,
				Name:            typeSpec.Name.Name,
				BackingTypeName: tree.Info.Defs[typeSpec.Name].Type().String(),
			}
			if !yield(def) {
				return
			}
		}
	}
}

func run(args []string, out io.Writer) (err error) {
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, args...)
	if err != nil {
		return fmt.Errorf("loading packages: %v", err)
	}
	enc := csv.NewWriter(out)
	defer enc.Flush()
	defer func() {
		if encErr := enc.Error(); encErr != nil && err == nil {
			err = fmt.Errorf("writing CSV: %v", encErr)
		}
	}()
	var defs []def
	for tree := range topLevelDecls(pkgs) {
		for def := range extractSentinels(tree) {
			defs = append(defs, def)
		}
		for def := range extractStructured(tree) {
			defs = append(defs, def)
		}
	}
	slices.SortFunc(defs, compareDef)
	for _, def := range defs {
		if err := def.Write(enc); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()
	if err := run(flag.Args(), os.Stdout); err != nil {
		log.Fatalln(err)
	}
}
