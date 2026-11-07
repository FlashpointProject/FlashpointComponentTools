//go:build darwin || freebsd || netbsd || openbsd || dragonfly

package main

import "golang.org/x/sys/unix"

const ioctlReadTermios = unix.TIOCGETA
