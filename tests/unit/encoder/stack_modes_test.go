package encoder_test

import "testing"

// Stack-type aliases map to the addressing modes the ARM ARM defines.
func TestEncodeStackModeAliases(t *testing.T) {
	const (
		// cccc 100P USWL nnnn rrrrrrrrrrrrrrrr with Rn=R1, W=1, list {R2, R3}
		stmIA = 0xE8A1000C
		stmIB = 0xE9A1000C
		stmDA = 0xE821000C
		stmDB = 0xE921000C
		ldmIA = 0xE8B1000C
		ldmIB = 0xE9B1000C
		ldmDA = 0xE831000C
		ldmDB = 0xE931000C
	)
	tests := []struct {
		mnemonic string
		want     uint32
	}{
		{"STMIA", stmIA}, {"STMIB", stmIB}, {"STMDA", stmDA}, {"STMDB", stmDB},
		{"LDMIA", ldmIA}, {"LDMIB", ldmIB}, {"LDMDA", ldmDA}, {"LDMDB", ldmDB},
		{"STMFD", stmDB}, {"LDMFD", ldmIA},
		{"STMFA", stmIB}, {"LDMFA", ldmDA},
		{"STMED", stmDA}, {"LDMED", ldmIB},
		{"STMEA", stmIA}, {"LDMEA", ldmDB},
	}
	for _, tt := range tests {
		t.Run(tt.mnemonic, func(t *testing.T) {
			got := encodeInstruction(t, newTestEncoder(), tt.mnemonic, []string{"R1!", "{R2, R3}"}, 0x8000)
			if got != tt.want {
				t.Errorf("%s R1!, {R2, R3} = 0x%08X, want 0x%08X", tt.mnemonic, got, tt.want)
			}
		})
	}
}
