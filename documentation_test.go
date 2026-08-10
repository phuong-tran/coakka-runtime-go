package coakka_v2_connector

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestReadmeUsesPublicFileLaneContract(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README: %v", err)
	}

	text := string(readme)
	const publicContract = "https://github.com/phuong-tran/coakka-publish/blob/main/docs/runtime-file-transfer.md"
	if !strings.Contains(text, publicContract) {
		t.Fatalf("README is missing public file-lane contract: %s", publicContract)
	}
	for _, privateRepository := range []string{"coakkaJVMConnector", "coakkaCoreNativeDev"} {
		if strings.Contains(text, privateRepository) {
			t.Fatalf("README contains private repository reference: %s", privateRepository)
		}
	}
}

func TestRootAPIDoesNotExposeWireCodecTypes(t *testing.T) {
	if _, err := os.Stat(filepath.Join("coakka", "v2")); !os.IsNotExist(err) {
		t.Fatalf("generated wire package must not be public: %v", err)
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list root Go files: %v", err)
	}
	forbiddenImports := map[string]bool{
		"github.com/phuong-tran/coakka-runtime-go/coakka/v2":       true,
		"github.com/phuong-tran/coakka-runtime-go/internal/wirepb": true,
		"google.golang.org/protobuf/proto":                         true,
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", path, parseErr)
		}
		for _, imported := range parsed.Imports {
			importPath, unquoteErr := strconv.Unquote(imported.Path.Value)
			if unquoteErr != nil {
				t.Fatalf("parse import in %s: %v", path, unquoteErr)
			}
			if forbiddenImports[importPath] {
				t.Fatalf("root package imports private wire implementation %q in %s", importPath, path)
			}
		}
		for _, declaration := range parsed.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, spec := range general.Specs {
				typeSpec := spec.(*ast.TypeSpec)
				if typeSpec.Assign.IsValid() && (typeSpec.Name.Name == "Envelope" || typeSpec.Name.Name == "Deadletter") {
					t.Fatalf("%s must be a connector-owned concrete type, not an alias", typeSpec.Name.Name)
				}
			}
		}
	}
}
