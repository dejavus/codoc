package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dejavus/godoc-responsive/parser"
	"github.com/dejavus/godoc-responsive/server"
)

//path = ../os
//pname = os

func main() {
	p := flag.String("package", "", "Package directory page")
	exclude := flag.String("exclude", "", "Exclude directories separated by comma")
	s := flag.Bool("server", false, "Server mode")

	flag.Parse()

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	*p, _ = filepath.Abs(*p)
	dir := filepath.Dir(*p)
	//base := filepath.Base(*p)
	os.Chdir(dir)
	defer os.Chdir(pwd)

	_excludes := strings.Split(*exclude, ",")
	excludes := make(map[string]bool)
	for _, e := range _excludes {
		excludes[filepath.Join(*p, e)] = true
	}

	pkgs := make(map[string]*parser.Package)
	navi := &parser.Navi{
		Pname: "",
		Navis: make(map[string]*parser.Navi),
	}

	fmt.Println("====== pkgs ======")
	parser.ParseDir(*p, pkgs, navi, excludes)
	for pname, p := range pkgs {
		fmt.Printf("pname: %s,  p.Name: %s\n", pname, p.Name)
		for fname, _ := range p.Files {
			fmt.Printf("  fname: %s\n", fname)
		}

	}
	fmt.Println("====== navi ======")
	fmt.Printf("%s\n", navi.Pname)
	for pname, navi := range navi.Navis {
		fmt.Printf("%s %s\n", pname, navi.Pname)
	}

	if *s {
		serv := server.NewServer(":8000", pkgs, navi, pwd)
		serv.Listen()
	}

}
