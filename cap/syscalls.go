// +build linux

package cap

import (
	"syscall"

	"github.com/ishworgurung/libcap/psx"
)

// multisc provides syscalls overridable for testing purposes that
// support a single kernel security state for all OS threads.
// (Go build tree has no syscall.PerOSThreadSyscall support.)
var multisc = &syscaller{
	w3: psx.Syscall3,
	w6: psx.Syscall6,
	r3: syscall.RawSyscall,
	r6: syscall.RawSyscall6,
}

// singlesc provides a single threaded implementation. Users should
// take care to ensure the thread is OS locked.
var singlesc = &syscaller{
	w3: syscall.RawSyscall,
	w6: syscall.RawSyscall6,
	r3: syscall.RawSyscall,
	r6: syscall.RawSyscall6,
}
