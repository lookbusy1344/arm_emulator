package api

import (
	"os"
	"testing"

	"github.com/lookbusy1344/arm-emulator/tests/internal/sessiondirs"
)

func TestMain(m *testing.M) {
	os.Exit(sessiondirs.Run(m))
}
