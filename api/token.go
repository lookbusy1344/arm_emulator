package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lookbusy1344/arm-emulator/config"
)

const (
	tokenBytes      = 32
	tokenDirPerm    = 0o700
	bearerPrefix    = "Bearer "
	healthPath      = "/health"
	tokenFilePrefix = "api-token-"
)

// GenerateToken returns a random hex token for one backend launch.
func GenerateToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// TokenPath returns the token file for a backend on port, in the per-user arm-emu
// directory. Clients of that port read the token from it.
func TokenPath(port int) (string, error) {
	dir, err := config.UserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("%s%d", tokenFilePrefix, port)), nil
}

// WriteTokenFile writes token to path, readable only by the owner. Readers see the
// old token or the new one, never a partial file.
func WriteTokenFile(path, token string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, tokenDirPerm); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, filepath.Base(path)+".*") // created with mode 0600
	if err != nil {
		return err
	}
	_, writeErr := f.WriteString(token)
	closeErr := f.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(f.Name())
		return fmt.Errorf("write token file: %w", errors.Join(writeErr, closeErr))
	}
	if err := os.Rename(f.Name(), path); err != nil {
		_ = os.Remove(f.Name())
		return err
	}
	return nil
}

// RequireToken rejects every request except /health that does not carry
// "Authorization: Bearer <token>".
func RequireToken(token string, next http.Handler) http.Handler {
	want := []byte(token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == healthPath {
			next.ServeHTTP(w, r)
			return
		}
		got, ok := strings.CutPrefix(r.Header.Get("Authorization"), bearerPrefix)
		if !ok || subtle.ConstantTimeCompare([]byte(got), want) != 1 {
			w.Header().Set("WWW-Authenticate", strings.TrimSpace(bearerPrefix))
			writeError(w, http.StatusUnauthorized, "Missing or invalid API token")
			return
		}
		next.ServeHTTP(w, r)
	})
}
