package main

import (
	"embed"
	"io"
	"io/fs"
	"sync"
	"text/template"

	"github.com/semrush/zenrpc/v2/parser"
)

var (
	//go:embed embed
	embedFS      embed.FS
	unwrapFSOnce sync.Once
	unwrappedFS  fs.FS
)

func FS() fs.FS {
	unwrapFSOnce.Do(func() {
		fsys, err := fs.Sub(embedFS, "embed")
		if err != nil {
			panic(err)
		}
		unwrappedFS = fsys
	})
	return unwrappedFS
}

var (
	clientTemplate  *template.Template
	serviceTemplate *template.Template
)

func readFile(f fs.File) (string, error) {
	var result []byte
	buf := make([]byte, 1024)

	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			break
		}
		result = append(result, buf[:n]...)
	}

	return string(result), nil
}

func init() {
	cf, err := FS().Open("client.go.template")
	if err != nil {
		panic(err)
	}

	clContent, err := readFile(cf)
	if err != nil {
		panic(err)
	}

	clientTemplate = template.Must(template.New("client").
		Funcs(template.FuncMap{"definitions": parser.Definitions}).
		Parse(string(clContent)),
	)

	sf, err := FS().Open("zenrpc_server.go.template")
	if err != nil {
		panic(err)
	}

	sContent, err := readFile(sf)
	if err != nil {
		panic(err)
	}

	serviceTemplate = template.Must(template.New("service").
		Funcs(template.FuncMap{"definitions": parser.Definitions}).
		Parse(string(sContent)),
	)
}
