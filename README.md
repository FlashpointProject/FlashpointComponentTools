# Flashpoint Component Tools

Tools for installing and managing a copy of [Flashpoint Archive](https://flashpointarchive.org/) through its component system.

- `FlashpointInstaller.exe` installs Flashpoint into a new folder, with the components you pick, and creates shortcuts.
- `FlashpointManager.exe` adds, removes, updates and repairs components of an existing copy, or uninstalls it. The [Flashpoint Launcher](https://github.com/FlashpointProject/launcher) opens it.
- `fpm.exe` does the same from the command line (`fpm help`). The launcher reads `fpm list verbose` at startup.

Windows 10 or later, 64-bit. The manager and `fpm.exe` ship together in the `core-manager` component.

## Build

Go 1.27.3 or later.

```sh
scripts/build.sh            # dist/windows-amd64/: the three exes and core-manager.zip
go test ./...
```

The GUI uses [Gio](https://gioui.org) and builds for Windows without cgo; on macOS and Linux it needs cgo (development only).

## Layout

```
cmd/fpm, cmd/manager, cmd/installer   the three programs
internal/engine                       manifest -> plan -> download -> transactional apply
internal/gui                          the window (manager and installer mode)
internal/install                      installer and uninstall parts (destination, shortcuts, VC++ runtime)
docs/go-rewrite                       design notes
```

The design, the compatibility contract with the launcher and the on-disk formats are in `docs/go-rewrite/DESIGN.md`.
