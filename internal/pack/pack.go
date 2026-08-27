package pack

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/German4341374/http-repro-lab/internal/generator"
	"github.com/German4341374/http-repro-lab/internal/model"
)

func Write(path string, request model.RequestSpec, outputs []generator.Output) error {
	files := map[string][]byte{}
	raw, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return err
	}
	files["repro.json"] = append(raw, '\n')
	files[".env.example"] = []byte("AUTH_TOKEN=replace-me\nAPI_KEY=replace-me\n")
	files["README.md"] = []byte("# HTTP reproduction pack\n\nThis sanitized pack was created by HTTP Repro Lab. Review every artifact before sharing. Populate placeholders through a local secret mechanism and never commit real values.\n")
	files["SECURITY.md"] = []byte("# Security notice\n\nSanitization reduces risk but cannot guarantee removal of all sensitive information. Do not run mutating requests against production without independent review.\n")
	for _, output := range outputs {
		files["generated/"+output.Language+"/"+output.FileName] = []byte(output.Content)
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	var checksums strings.Builder
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		fmt.Fprintf(&checksums, "%s  %s\n", hex.EncodeToString(sum[:]), name)
	}
	files["SHA256SUMS"] = []byte(checksums.String())
	names = append(names, "SHA256SUMS")
	sort.Strings(names)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	defer file.Close()
	archive := zip.NewWriter(file)
	for _, name := range names {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.SetMode(0o640)
		writer, createErr := archive.CreateHeader(header)
		if createErr != nil {
			return createErr
		}
		if _, writeErr := writer.Write(files[name]); writeErr != nil {
			return writeErr
		}
	}
	return archive.Close()
}
