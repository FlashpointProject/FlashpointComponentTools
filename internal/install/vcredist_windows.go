package install

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const vcRedistURL = "https://aka.ms/vs/17/release/vc_redist.x86.exe"

// vcRedistArgs install without UI; the bundle still shows its UAC prompt.
// No restart: the user is told instead.
const vcRedistArgs = "/install /norestart /quiet"

// NeedsVCRedist reports whether the Visual C++ 2015-2022 x86 runtime is
// missing (always false outside Windows). The launcher's native parts are
// 32-bit, hence x86; all runtimes 2015 to 2022 share this key.
func NeedsVCRedist() bool {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\WOW6432Node\Microsoft\VisualStudio\14.0\VC\Runtimes\X86`, registry.QUERY_VALUE)
	if err != nil {
		return true
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue("Installed")
	return err != nil || v != 1
}

// InstallVCRedist downloads, verifies and runs the x86 runtime installer.
// ctx cancels the download only: a running installer is never killed
// half-way.
func InstallVCRedist(ctx context.Context, log *slog.Logger) (rebootNeeded bool, err error) {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	dir, err := os.MkdirTemp("", "fpm-vcredist-")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(dir)
	exe := filepath.Join(dir, "vc_redist.x86.exe")

	log.Info("downloading the Visual C++ runtime", "url", vcRedistURL)
	if err := downloadVCRedist(ctx, exe); err != nil {
		return false, fmt.Errorf("could not download the Visual C++ runtime: %w", err)
	}
	if err := verifySignature(exe); err != nil {
		return false, err
	}
	log.Info("running the Visual C++ runtime installer", "args", vcRedistArgs)
	code, err := runVCRedist(exe)
	if err != nil {
		return false, err
	}
	log.Info("the Visual C++ runtime installer finished", "exitCode", code)
	return vcRedistResult(code)
}

func downloadVCRedist(ctx context.Context, dst string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute) // ~14 MB
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, vcRedistURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the server answered %s", resp.Status)
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// verifySignature checks the file's Authenticode signature. HTTPS from
// Microsoft plus a valid signature is enough, so no signer check. No
// revocation check: it needs more network and fails behind some proxies.
func verifySignature(path string) error {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	data := &windows.WinTrustData{
		Size:             uint32(unsafe.Sizeof(windows.WinTrustData{})),
		UIChoice:         windows.WTD_UI_NONE,
		RevocationChecks: windows.WTD_REVOKE_NONE,
		UnionChoice:      windows.WTD_CHOICE_FILE,
		StateAction:      windows.WTD_STATEACTION_VERIFY,
		ProvFlags:        windows.WTD_REVOCATION_CHECK_NONE,
		FileOrCatalogOrBlobOrSgnrOrCert: unsafe.Pointer(&windows.WinTrustFileInfo{
			Size:     uint32(unsafe.Sizeof(windows.WinTrustFileInfo{})),
			FilePath: p,
		}),
	}
	err = windows.WinVerifyTrustEx(windows.InvalidHWND, &windows.WINTRUST_ACTION_GENERIC_VERIFY_V2, data)
	data.StateAction = windows.WTD_STATEACTION_CLOSE
	windows.WinVerifyTrustEx(windows.InvalidHWND, &windows.WINTRUST_ACTION_GENERIC_VERIFY_V2, data)
	if err != nil {
		return fmt.Errorf("the downloaded Visual C++ runtime installer is not validly signed (%w)", err)
	}
	return nil
}

// runVCRedist runs the installer and returns its exit code. The bundle is
// asInvoker and shows the UAC prompt itself.
func runVCRedist(exe string) (int, error) {
	err := exec.Command(exe, strings.Fields(vcRedistArgs)...).Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return exit.ExitCode(), nil
	}
	if err != nil {
		return 0, fmt.Errorf("could not start the Visual C++ runtime installer: %w", err)
	}
	return 0, nil
}
