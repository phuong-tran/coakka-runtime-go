package coakka_v2_connector

import (
	"os"
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
