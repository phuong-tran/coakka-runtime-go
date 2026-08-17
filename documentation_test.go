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

func TestGoCompatibilityFloorAndCurrentToolchainCI(t *testing.T) {
	module, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !strings.Contains(string(module), "\ngo 1.22\n") {
		t.Fatal("go.mod must retain the tested Go 1.22 compatibility floor")
	}

	workflow, err := os.ReadFile(filepath.Join(".github", "workflows", "go-ci.yml"))
	if err != nil {
		t.Fatalf("read go-ci workflow: %v", err)
	}
	text := string(workflow)
	for _, required := range []string{"\"1.22.12\"", "- stable"} {
		if !strings.Contains(text, required) {
			t.Fatalf("go-ci workflow is missing toolchain matrix entry %q", required)
		}
	}
	if strings.Contains(text, "go-version-file: go.mod") {
		t.Fatal("go-ci must not conflate the module compatibility floor with the current CI toolchain")
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
