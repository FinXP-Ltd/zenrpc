package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"time"

	"github.com/semrush/zenrpc/v2/parser"
)

const (
	version = "2.1.1"

	openIssueURL = "https://github.com/semrush/zenrpc/issues/new"
	githubURL    = "https://github.com/semrush/zenrpc"

	logo = `

    _________                  _______________________________  
    /   _____/ ____   ____     |__\______   \______   \_   ___ \ 
    \_____  \_/ __ \ /    \    |  ||       _/|     ___/    \  \/ 
    /        \  ___/|   |  \   |  ||    |   \|    |   \     \____
   /_______  /\___  >___|  /\__|  ||____|_  /|____|    \______  /
		   \/     \/     \/\______|       \/                  \/    

`
)

func main() {
	fmt.Print(logo)

	start := time.Now()
	fmt.Printf("Generator version: %s\n", version)

	var filename string
	if len(os.Args) > 1 {
		filename = os.Args[len(os.Args)-1]
		fmt.Println("Args: ", os.Args)
	} else {
		filename = os.Getenv("GOFILE")
		fmt.Println("ENV GOFILE: ", filename)

	}

	if len(filename) == 0 {
		fmt.Fprintln(os.Stderr, "File path is empty")
		os.Exit(1)
	}

	fmt.Printf("Entrypoint: %s\n", filename)

	// create package info
	pi, err := parser.NewPackageInfo(filename)
	if err != nil {
		printError(err)
		os.Exit(1)
	}

	serverOutFilename := pi.OutputFilename()
	clientOutFilename := pi.OutputFilenameClient("_jsonrpc2_client.go")

	clientOutDir := filepath.Dir(clientOutFilename)
	if _, err := os.Stat(clientOutDir); os.IsNotExist(err) {
		if err := os.MkdirAll(clientOutDir, os.ModePerm); err != nil {
			printError(err)
			os.Exit(1)
		}
		fmt.Printf("Created client dir directory: %s\n", clientOutDir)

	}

	// remove output file if it already exists
	for _, f := range []string{serverOutFilename, clientOutFilename} {
		if _, err := os.Stat(f); err == nil {
			if err := os.Remove(f); err != nil {
				printError(err)
				os.Exit(1)
			}
		}
	}

	dir := filepath.Dir(serverOutFilename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			printError(err)
			os.Exit(1)
		}
	}

	if err := pi.Parse(filename); err != nil {
		printError(err)
		os.Exit(1)
	}

	if len(pi.Services) == 0 {
		fmt.Fprintln(os.Stderr, "Services not found")
		os.Exit(1)
	}

	if err := generateServerFile(serverOutFilename, pi); err != nil {
		printError(err)
		os.Exit(1)
	}

	fmt.Printf("Generated server: %s\n", serverOutFilename)

	if err := generateClientFile(clientOutFilename, pi); err != nil {
		printError(err)
		os.Exit(1)
	}

	fmt.Printf("Generated client: %s\n", serverOutFilename)

	fmt.Printf("Duration: %dms\n", int64(time.Since(start)/time.Millisecond))
	fmt.Println()
	fmt.Print(pi)
	fmt.Println()
}

func printError(err error) {
	// print error to stderr
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)

	// print contact information to stdout
	fmt.Println("\nYou may help us and create issue:")
	fmt.Printf("\t%s\n", openIssueURL)
	fmt.Println("For more information, see:")
	fmt.Printf("\t%s\n\n", githubURL)
}

func generateServerFile(outputFileName string, pi *parser.PackageInfo) error {
	file, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	output := new(bytes.Buffer)
	if err := serviceTemplate.Execute(output, pi); err != nil {
		return err
	}

	source, err := format.Source(output.Bytes())
	if err != nil {
		fmt.Printf("Error formatting generated code: %s %s\n", source, err)
		// return err
	}

	if _, err = file.Write(source); err != nil {
		return err
	}

	return nil
}

func generateClientFile(outputFileName string, pi *parser.PackageInfo) error {
	file, err := os.Create(outputFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	output := new(bytes.Buffer)
	if err := clientTemplate.Execute(output, pi); err != nil {
		return err
	}

	source, err := format.Source(output.Bytes())
	if err != nil {
		fmt.Printf("Error formatting generated code: %s\n %s", source, err)
		// return err
	}

	if _, err = file.Write(source); err != nil {
		return err
	}

	return nil
}
