package main

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

func main() {
    if len(os.Args) < 2 {
        panic("Insufficient arguments provided")
    }

    switch os.Args[1] {
    case "start":
        initializeParent()
    case "init":
        initializeChild()
    default:
        panic("Unknown command")
    }
}

func initializeParent() {
    args := append([]string{"init"}, os.Args[2:]...)
    cmd := exec.Command("/proc/self/exe", args...)
    cmd.SysProcAttr = &unix.SysProcAttr{
        Cloneflags: unix.CLONE_NEWUTS | unix.CLONE_NEWPID | unix.CLONE_NEWNS,
    }
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        fmt.Printf("Failed to run command: %v\n", err)
        os.Exit(1)
    }
}

func initializeChild() {
    if err := bindMount("rootfs", "rootfs"); err != nil {
        panic(err)
    }
    if err := os.MkdirAll("rootfs/old_rootfs", 0755); err != nil {
        panic(err)
    }
    if err := unix.PivotRoot("rootfs", "rootfs/old_rootfs"); err != nil {
        panic(err)
    }
    if err := os.Chdir("/"); err != nil {
        panic(err)
    }
    if len(os.Args) < 3 {
        panic("No command specified to run")
    }

    command := exec.Command(os.Args[2], os.Args[3:]...)
    command.Stdin = os.Stdin
    command.Stdout = os.Stdout
    command.Stderr = os.Stderr

    if err := command.Run(); err != nil {
        fmt.Printf("Command execution failed: %v\n", err)
        os.Exit(1)
    }
}

func bindMount(source, target string) error {
    return unix.Mount(source, target, "", unix.MS_BIND, "")
}
