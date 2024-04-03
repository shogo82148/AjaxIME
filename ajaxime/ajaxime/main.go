package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "POST")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		// preflight request
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := r.Context()
	data, err := s.decode(r)
	if err != nil {
		slog.ErrorContext(ctx, "failed to decode", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		slog.ErrorContext(ctx, "failed to write response", slog.String("error", err.Error()))
		panic(err)
	}
}

func (s *Server) decode(r *http.Request) ([]byte, error) {
	// read the request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var input struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	// parse the query
	result, err := s.parse(input.Query)
	if err != nil {
		return nil, err
	}

	// encode the result
	var output struct {
		Result []string `json:"result"`
	}
	output.Result = result
	return json.Marshal(&output)
}

func (s *Server) parse(query string) ([]string, error) {
	lattice, err := mecab.NewLattice()
	if err != nil {
		return nil, err
	}
	defer lattice.Destroy()

	lattice.SetSentence(query)
	lattice.AddRequestType(mecab.RequestTypeNBest)
	if err := s.tagger.ParseLattice(lattice); err != nil {
		return nil, err
	}

	result := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		var buf strings.Builder
		for node := lattice.BOSNode(); !node.IsZero(); node = node.Next() {
			if node.Stat() == mecab.NormalNode {
				buf.WriteString(node.Feature())
			}
		}
		result = append(result, buf.String())
		if !lattice.Next() {
			break
		}
	}
	return result, nil
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
