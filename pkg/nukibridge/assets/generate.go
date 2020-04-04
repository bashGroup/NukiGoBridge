// +build ignore

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var cwd, _ = os.Getwd()
	templates := http.Dir(filepath.Join(cwd, "..", "..", "assets"))

	if err := vfsgen.Generate(templates, vfsgen.Options{
		Filename:     filepath.Join(cwd, "assets", "templates", "templates_vfsdata.go"),
		PackageName:  "templates",
		BuildTags:    "!dev",
		VariableName: "Assets",
	}); err != nil {
		log.Fatalln(err)
	}
}
