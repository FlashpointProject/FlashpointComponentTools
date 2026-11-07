package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/launch"
)

var procIsProcessInJob = windows.NewLazySystemDLL("kernel32.dll").NewProc("IsProcessInJob")

// inJob reports whether this process belongs to a job object. The launcher
// (Node/libuv) puts the processes it spawns into a job that kills them when
// the launcher exits, about 2.5 s after it starts us with /update /launcher.
func inJob() bool {
	var r int32
	ok, _, _ := procIsProcessInJob.Call(uintptr(windows.CurrentProcess()), 0, uintptr(unsafe.Pointer(&r)))
	return ok != 0 && r != 0
}

// waitParent waits up to d for the process that started us to exit. After
// a self-update its exe is the backup our recovery deletes, and Windows
// can't delete the file of a running exe.
func waitParent(d time.Duration) {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(os.Getppid()))
	if err != nil {
		return // gone already
	}
	defer windows.CloseHandle(h)
	windows.WaitForSingleObject(h, uint32(d.Milliseconds()))
}

// startOutsideJob starts exe like launch.Start, but outside our job. With
// CREATE_BREAKAWAY_FROM_JOB if the job allows it; else a plain child, which
// is outside a job with silent breakaway (libuv's job has both; the C#
// manager survived the launcher only by relaunching from %TEMP%).
func startOutsideJob(exe string, args []string) error {
	cmd := exec.Command(exe, args...)
	cmd.SysProcAttr = launch.DetachedAttr()
	cmd.SysProcAttr.CreationFlags |= windows.CREATE_BREAKAWAY_FROM_JOB
	if err := cmd.Start(); err == nil {
		return nil
	}
	return launch.Start(exe, args, "")
}

// deleteAfterExit deletes exe, then its folder and root if they are empty,
// once this process has ended: Windows can't delete a running exe. A hidden
// cmd waits two seconds first, as the old manager's did; its working folder
// is root's parent, so it doesn't hold root open.
func deleteAfterExit(exe, root string) error {
	cmd := exec.Command("cmd.exe")
	cmd.Dir = filepath.Dir(root)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NO_WINDOW,
		CmdLine: fmt.Sprintf(`cmd.exe /c ping -n 3 127.0.0.1 >nul & del /f /q "%s" & rmdir "%s" & rmdir "%s"`,
			exe, filepath.Dir(exe), root),
	}
	return cmd.Start()
}
