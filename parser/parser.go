package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Navi struct {
	Pname string
	Navis map[string]*Navi
}

// Package type
type Package struct {
	Name  string
	Doc   string
	Files map[string]*File
}

func (p *Package) GetDecls() []string {
	ret := make([]string, 0)
	for _, f := range p.Files {
		for _, d := range f.Decls {
			ret = append(ret, d.GetName())
		}
	}
	return ret
}

func (p *Package) GetFuncs() []string {
	ret := make([]string, 0)
	for _, f := range p.Files {
		for _, fn := range f.Funcs {
			ret = append(ret, fn.GetName())
		}
	}
	return ret
}

func (p *Package) GetDoc() string {
	return p.Doc
}

type Decl interface {
	GetName() string
	GetDoc() string
	GetCode() string
}

type File struct {
	Doc   string
	Funcs []Decl
	Decls []Decl
}

func (f *File) GetDoc() string {
	return f.Doc
}

type Func struct {
	Name string
	Doc  string
	Code string
}

func (fn *Func) GetName() string {
	return fn.Name
}

func (fn *Func) GetDoc() string {
	return fn.Doc
}

func (fn *Func) GetCode() string {
	return fmt.Sprintf("%s\n", fn.Code)
}

type GenDecl struct {
	Name string
	Doc  string
	Code string
}

func (d *GenDecl) GetName() string {
	return d.Name
}

func (d *GenDecl) GetDoc() string {
	return d.Doc
}

func (d *GenDecl) GetCode() string {
	return fmt.Sprintf("%s\n", d.Code)
}

func AstPkgToPkg(fset *token.FileSet, pkg *ast.Package, parentPname string) *Package {
	ret := &Package{
		Files: make(map[string]*File),
		Name:  path.Join(parentPname, pkg.Name),
	}
	for fname, f := range pkg.Files {
		ff, err := os.Open(fname)
		if err != nil {
			log.Print(err)
			continue
		}
		file := &File{
			Doc:   f.Doc.Text(),
			Funcs: make([]Decl, 0),
			Decls: make([]Decl, 0),
		}
		if f.Doc.Text() != "" {
			ret.Doc = f.Doc.Text()
		}
		ret.Files[filepath.Join(ret.Name, filepath.Base(fname))] = file
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				//fmt.Println(d.Name.Name)
				l := fset.Position(d.End()).Offset - fset.Position(d.Pos()).Offset
				b := make([]byte, l)
				if _, err := ff.ReadAt(b, int64(fset.Position(d.Pos()).Offset)); err != nil {
					log.Print(err)
					continue
				}
				file.Funcs = append(file.Funcs, &Func{
					Name: d.Name.Name,
					Doc:  d.Doc.Text(),
					Code: string(b),
				})

			case *ast.GenDecl:
				if d.Tok.String() == "import" {
					continue
				}
				/*
					for _, spec := range d.Specs {
						if spec, ok := spec.(*ast.ValueSpec); ok {
							for _, name := range spec.Names {
								fmt.Println(name.Name)
							}

						}
					}
				*/
				l := fset.Position(d.End()).Offset - fset.Position(d.Pos()).Offset
				b := make([]byte, l)
				if _, err := ff.ReadAt(b, int64(fset.Position(d.Pos()).Offset)); err != nil {
					log.Print(err)
					continue
				}
				names := make([]string, 0)
				for _, spec := range d.Specs {
					if spec, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range spec.Names {
							if name.Name != "" {
								names = append(names, name.Name)
							}
						}
					}
				}
				file.Decls = append(file.Decls, &GenDecl{
					Name: strings.Join(names, ", "),
					Doc:  d.Doc.Text(),
					Code: string(b),
				})
			}
		}
		ff.Close()
	}
	return ret
}

func ParseDir(abspath string, pkgOut map[string]*Package, parent *Navi, excludes map[string]bool) error {
	if _, ok := excludes[abspath]; ok {
		return nil
	}
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, abspath, func(info os.FileInfo) bool {
		if filepath.Ext(info.Name()) == ".go" {
			return true
		}
		return false
	}, parser.ParseComments)
	if err != nil {
		return err
	}

	abspath = filepath.ToSlash(abspath)
	pname := filepath.Base(abspath)
	for _, astPkg := range pkgs {
		pkg := AstPkgToPkg(fset, astPkg, parent.Pname)
		pkgOut[pkg.Name] = pkg
		parent.Navis[filepath.Base(pkg.Name)] = &Navi{
			Pname: path.Join(parent.Pname, filepath.Base(pkg.Name)),
			Navis: make(map[string]*Navi),
		}
	}

	infos, err := ioutil.ReadDir(abspath)
	if err != nil {
		return err
	}

	for _, info := range infos {
		if info.IsDir() {
			next, ok := parent.Navis[pname]
			if !ok {
				next = &Navi{
					Pname: path.Join(parent.Pname, pname),
					Navis: make(map[string]*Navi),
				}
				if pname == "internal" {
					fmt.Println("?????", parent.Pname)
				}
				parent.Navis[pname] = next
			}
			if err := ParseDir(filepath.Join(abspath, info.Name()), pkgOut, next, excludes); err != nil {
				return err
			}
		}
	}

	return nil
}
