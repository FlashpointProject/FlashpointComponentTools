package install

import (
	"errors"
	"fmt"
)

// vcRedistResult maps the runtime installer's exit code (Windows Installer
// codes, passed through by its Burn bundle).
func vcRedistResult(code int) (rebootNeeded bool, err error) {
	switch code {
	case 0, 1638: // 1638: this or a newer version is already installed
		return false, nil
	case 3010: // success, a restart completes it
		return true, nil
	case 1602, 1223: // user cancelled (1223 when the UAC prompt is declined)
		return false, errors.New("the Windows permission prompt was declined, so the Visual C++ runtime was not installed")
	case 1618: // ERROR_INSTALL_ALREADY_RUNNING: Windows Installer is busy
		return false, errors.New("another program is installing something right now; wait until it finishes, then try again")
	}
	return false, fmt.Errorf("the Visual C++ runtime installer failed with exit code %d", code)
}
