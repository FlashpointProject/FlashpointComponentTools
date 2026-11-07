package install

import "testing"

func TestVCRedistResult(t *testing.T) {
	tests := []struct {
		code   int
		reboot bool
		err    string // "" = nil
	}{
		{0, false, ""},
		{1638, false, ""},
		{3010, true, ""},
		{1602, false, "the Windows permission prompt was declined, so the Visual C++ runtime was not installed"},
		{1223, false, "the Windows permission prompt was declined, so the Visual C++ runtime was not installed"},
		{1618, false, "another program is installing something right now; wait until it finishes, then try again"},
		{1603, false, "the Visual C++ runtime installer failed with exit code 1603"},
	}
	for _, tt := range tests {
		reboot, err := vcRedistResult(tt.code)
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if reboot != tt.reboot || msg != tt.err {
			t.Errorf("vcRedistResult(%d) = %v, %v; want %v, %q", tt.code, reboot, err, tt.reboot, tt.err)
		}
	}
}
