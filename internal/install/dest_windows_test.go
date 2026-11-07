package install

import (
	"strings"
	"testing"
)

func TestCheckDestWindows(t *testing.T) {
	// Fake places on a drive that doesn't exist, so the results don't
	// depend on this PC (a missing Q:\... is simply a new folder).
	env := map[string]string{
		"ProgramFiles":       `Q:\Program Files`,
		"ProgramFiles(x86)":  `Q:\Program Files (x86)`,
		"SystemRoot":         `Q:\Windows`,
		"TMP":                `Q:\Users\u\AppData\Local\Temp\`,
		"TEMP":               `Q:\Users\u\AppData\Local\Temp\`,
		"OneDrive":           `Q:\Users\u\OneDrive`,
		"OneDriveConsumer":   "",
		"OneDriveCommercial": `Q:\Users\u\OneDrive - Work`,
	}
	for k, v := range env {
		t.Setenv(k, v)
	}

	tests := []struct {
		path string
		want string // error substring, "" = ok
	}{
		{`C:foo`, "Enter a full path, such as C:\\Flashpoint."},
		{`\foo`, "Enter a full path"},
		{`Q:\`, "Choose a folder, not a whole drive, such as Q:\\Flashpoint."},
		{`\\server\share`, "not a whole drive, such as \\\\server\\share\\Flashpoint."},
		{`\\server\share\`, "not a whole drive"},
		{`Q:\Program Files`, "Flashpoint can't be installed in Program Files, because Windows protects it. Choose another folder, such as Q:\\Flashpoint."},
		{`q:\program files\Flashpoint`, "in Program Files"},
		{`Q:\Program Files (x86)\Flashpoint`, "in Program Files"},
		{`Q:\Windows\Games`, "in the Windows folder"},
		{`Q:\Users\u\AppData\Local\Temp\Flashpoint`, "in the temporary folder, because Windows may clean it up"},
		{`Q:\Users\u\OneDrive\Games\Flashpoint`, "in OneDrive, because syncing breaks it"},
		{`Q:\Users\u\OneDrive - Work`, "in OneDrive"},
		{`Q:\Program Files2\Flashpoint`, ""},
		{`Q:\Users\u\OneDriveBackup`, ""},
		{`Q:\Users\u\AppData\Local\Flashpoint`, ""},
		{`Q:/Games//Flashpoint/`, ""},
	}
	for _, tt := range tests {
		_, create, err := CheckDest(tt.path)
		switch {
		case tt.want == "" && (err != nil || create == ""):
			t.Errorf("CheckDest(%q) = %q, %v; want ok, new folder", tt.path, create, err)
		case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
			t.Errorf("CheckDest(%q) = %v, want error containing %q", tt.path, err, tt.want)
		}
	}
	// Q: doesn't exist: create stops below the volume root.
	if abs, create, _ := CheckDest(`Q:/Games//Flashpoint/`); abs != `Q:\Games\Flashpoint` || create != `Q:\Games` {
		t.Errorf(`CheckDest = %q, %q; want Q:\Games\Flashpoint, Q:\Games`, abs, create)
	}
}

func TestDefaultDestWindows(t *testing.T) {
	for exe, want := range map[string]string{
		`D:\Downloads\FlashpointInstaller.exe`:       `D:\Flashpoint`,
		`\\nas\games\FlashpointInstaller.exe`:        `\\nas\games\Flashpoint`,
		`c:\Users\u\Desktop\FlashpointInstaller.exe`: `c:\Flashpoint`,
	} {
		if got := DefaultDest(exe); got != want {
			t.Errorf("DefaultDest(%q) = %q, want %q", exe, got, want)
		}
	}
}
