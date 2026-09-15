package workerd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Onicc/frp-panel/pb"
)

func TestWorkerFilesStayInsideWorkDirectory(t *testing.T) {
	root := t.TempDir()
	workerID := "../../outside"
	entry := "../../entry.js"
	code := "export default {}"
	worker := &pb.Worker{WorkerId: &workerID, CodeEntry: &entry, Code: &code}
	FillWorkerValue(worker, 1)

	if err := WriteWorkerCodeToFile(context.Background(), worker, root); err != nil {
		t.Fatal(err)
	}
	path := CodeFilePath(context.Background(), worker, root)
	relative, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(relative, "..") {
		t.Fatalf("worker escaped root: %s", path)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != code {
		t.Fatalf("worker code was not safely written: %q %v", data, err)
	}
}
