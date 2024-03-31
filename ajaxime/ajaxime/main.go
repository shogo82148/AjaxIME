package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/shogo82148/go-mecab"
	"github.com/shogo82148/ridgenative"
)

const VERSION = "0.0.1"

type Server struct {
	tagger mecab.MeCab
}

func NewServer() (*Server, error) {
	tagger, err := mecab.New(map[string]string{
		"dicdir": "/usr/local/lib/mecab/dic/mecab-as-kkc",
	})
	if err != nil {
		return nil, err
	}
	return &Server{
		tagger: tagger,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lattice, err := mecab.NewLattice()
	if err != nil {
		slog.Error("failed to create lattice", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer lattice.Destroy()

	lattice.SetSentence("こんにちはせかい")
	if err := s.tagger.ParseLattice(lattice); err != nil {
		slog.Error("failed to parse lattice", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	for node := lattice.BOSNode(); !node.IsZero(); node = node.Next() {
		log.Println(node.Surface(), node.Feature())
	}
}

func main() {
	var version bool
	flag.BoolVar(&version, "version", false, "print version")
	flag.Parse()
	if version {
		fmt.Printf("ajaxime version %s\n", VERSION)
		return
	}

	srv, err := NewServer()
	if err != nil {
		slog.Error("failed to create server", slog.String("error", err.Error()))
		os.Exit(1)
	}
	http.Handle("/", srv)
	ridgenative.ListenAndServe(":8080", srv)
}
