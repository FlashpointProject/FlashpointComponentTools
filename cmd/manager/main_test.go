package main

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want cliArgs
	}{
		{"none", nil, cliArgs{}},
		{"launcher open manager", []string{"/update", "/launcher"},
			cliArgs{update: true, launcher: true, restart: []string{"/update", "/launcher"}}},
		{"launcher on-demand install: IDs in one argument", []string{"/notemp", "/download", "supportpack-a supportpack-b\tsupportpack-c"},
			cliArgs{ids: []string{"supportpack-a", "supportpack-b", "supportpack-c"}, notemp: true, restart: []string{"/notemp"}}},
		{"IDs across arguments until the next switch", []string{"/download", "a", "b c", "/LAUNCHER", "d"},
			cliArgs{launcher: true, ids: []string{"a", "b", "c"}, ignored: []string{"d"}, restart: []string{"/LAUNCHER", "d"}}},
		{"mixed case switches and IDs, duplicates dropped", []string{"/Update", "/DownLoad", "Core-Foo core-foo", "/NoTemp"},
			cliArgs{update: true, ids: []string{"core-foo"}, notemp: true, restart: []string{"/Update", "/NoTemp"}}},
		{"/download without IDs", []string{"/download"}, cliArgs{}},
		{"/download with an empty argument", []string{"/download", " ", "/update"},
			cliArgs{update: true, restart: []string{"/update"}}},
		{"--path ends the ID list, its value is not a switch", []string{"/download", "a", "--path", "/fp", "b"},
			cliArgs{ids: []string{"a"}, path: "/fp", ignored: []string{"b"}, restart: []string{"--path", "/fp", "b"}}},
		{"--path=", []string{"--path=C:\\FP", "/launcher"},
			cliArgs{launcher: true, path: `C:\FP`, restart: []string{"--path=C:\\FP", "/launcher"}}},
		{"--path without value", []string{"--path"}, cliArgs{ignored: []string{"--path"}, restart: []string{"--path"}}},
		{"relaunch marker is not passed on", []string{"/update", "/launcher", "/relaunched"},
			cliArgs{update: true, launcher: true, relaunch: true, restart: []string{"/update", "/launcher"}}},
		{"unknown", []string{"/foo", "-x", "bar", "--version"},
			cliArgs{ignored: []string{"/foo", "-x", "bar", "--version"}, restart: []string{"/foo", "-x", "bar", "--version"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseArgs(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseArgs(%q)\n got %+v\nwant %+v", tt.args, got, tt.want)
			}
		})
	}
}
