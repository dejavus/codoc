package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/dejavus/godoc-responsive/parser"
	"github.com/gorilla/mux"
)

type Server struct {
	server *http.Server
}

func NewServer(addr string, pkgs map[string]*parser.Package, navi *parser.Navi, pwd string) *Server {
	s := &Server{
		server: &http.Server{
			Addr: addr,
		},
	}

	tmpl := template.Must(template.ParseGlob(filepath.Join(pwd, "templates", "*.html")))
	r := mux.NewRouter().StrictSlash(true)

	r.Handle("/", Index(tmpl, navi)).Methods("GET")
	r.Handle("/pkg/{package:.*}", Package(tmpl, pkgs, navi)).Methods("GET")

	fs := http.FileServer(http.Dir(filepath.Join(pwd, "./assets")))
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", fs))

	s.server.Handler = r

	return s

}

func Index(tmpl *template.Template, navi *parser.Navi) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(navi.Navis)
		tmpl.ExecuteTemplate(w, "index.html", navi.Navis)
	})
}

type data struct {
	Pkg   *parser.Package
	Navi  map[string]*parser.Navi
	Decls []string
	Funcs []string
}

func Package(tmpl *template.Template, pkgs map[string]*parser.Package, navi *parser.Navi) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pname := vars["package"]
		if pkg, ok := pkgs[pname]; ok {
			d := data{
				Pkg:   pkg,
				Navi:  navi.Navis,
				Decls: pkg.GetDecls(),
				Funcs: pkg.GetFuncs(),
			}
			tmpl.ExecuteTemplate(w, "package.html", d)
			return
		}
		http.NotFound(w, r)
	})
}

func (s *Server) Listen() {
	fmt.Printf("Listening at %s\n", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
