package upgrade

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

func verifyFileNonEmpty(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat failed: %w", err)
	}
	if st.Size() <= 0 {
		return fmt.Errorf("file is empty: %s", path)
	}
	return nil
}

func verifySHA256(path, expected string) error {
	expected = strings.ToLower(strings.TrimSpace(expected))
	if len(expected) != sha256.Size*2 {
		return fmt.Errorf("invalid SHA-256 checksum")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, io.LimitReader(file, 512<<20)); err != nil {
		return err
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != expected {
		return fmt.Errorf("SHA-256 mismatch: got %s", actual)
	}
	return nil
}

func checksumForAsset(content []byte, asset string) (string, error) {
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.TrimPrefix(fields[len(fields)-1], "*") == asset {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("asset %s is missing from checksums.txt", asset)
}
