// Progam web provides an example of a webserver using capabilities to
// bind to a privileged port, and then drop all capabilities before
// handling the first web request.
//
// This program cannot work reliably as a pure Go application without
// the equivalent of the Go runtime patch that adds a POSIX semantics
// wrapper around the system calls that change kernel state. A patch
// for the pure Go compiler/runtime to add this support is available
// here [2019-12-14]:
//
//    https://go-review.googlesource.com/c/go/+/210639/
//
// Until that patch, or something like it, is absorbed into the Go
// runtime the only way to get capabilities to work reliably on the Go
// runtime is to use something like libpsx via cgo to do capability
// setting syscalls in C with POSIX semantics. As of this build of the
// Go "libcap/cap" package, courtesy of the "libcap/psx" package, this
// is how things work.
//
// To set this up, compile and empower this binary as follows
// (packages libcap/{cap,psx} should be installed, as must libpsx.a):
//
//   go build web.go
//   sudo setcap cap_setpcap,cap_net_bind_service=p web
//   ./web --port=80
//
// Make requests using wget and observe the log of web:
//
//   wget -o/dev/null -O/dev/stdout localhost:80
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"syscall"

	"github.com/ishworgurung/libcap/cap"
)

var (
	port               = flag.Int("port", 0, "port to listen on")
	skipPrivilegeCheck = flag.Bool("skip", false, "skip raising the effective capability - will fail for low ports")
)

// init aborts the program if it is running setuid something,
// or being invoked by root.  That is, the preparer isn't setting up
// the program correctly.
func init() {
	euid := syscall.Geteuid()
	uid := syscall.Getuid()
	egid := syscall.Getegid()
	gid := syscall.Getgid()
	if uid != euid || gid != egid {
		log.Fatalf("go runtime is setuid uids:(%d vs %d), gids(%d vs %d)", uid, euid, gid, egid)
	}
	if uid == 0 {
		log.Fatalf("go runtime is running as root - cheating")
	}
}

// listen creates a listener by raising effective privilege only to
// bind to address and then lowering that effective privilege.
func (h *HandlerContext) listen(network, address string) (net.Listener, error) {
	if *skipPrivilegeCheck == true {
		l, err := net.Listen(network, address)
		if err != nil {
			return nil, errors.New(err.Error())
		}
		if err := cap.ModeNoPriv.Set(); err != nil {
			return nil, errors.New(err.Error())
		}
		return l, nil
	}
	originalCapabilities := cap.GetProc()
	defer func() {
		if err := originalCapabilities.SetProc(); err != nil {
			log.Fatal(err)
		}
	}() // restore original caps on exit.
	dupCapabilities, err := originalCapabilities.Dup()
	if err != nil {
		return nil, errors.New(err.Error())
	}
	if on, err := dupCapabilities.GetFlag(cap.Permitted, cap.NET_BIND_SERVICE); !on {
		if err != nil {
			return nil, errors.New(err.Error())
		} else {
			return nil, errors.New(
				fmt.Sprintf(
					"insufficient privilege to bind to low ports - want %q, have %q",
					cap.NET_BIND_SERVICE, dupCapabilities))
		}
	}
	if err := dupCapabilities.SetFlag(cap.Effective, true, cap.NET_BIND_SERVICE); err != nil {
		return nil, errors.New(err.Error()) // unable to set capability
	}
	if err := dupCapabilities.SetProc(); err != nil {
		return nil, errors.New(err.Error()) // unable to raise capabilities
	}
	return net.Listen(network, address)
}

// HandlerContext is used to abstract the ServeHTTP function.
type HandlerContext struct {
	skipPrivilegeCheck bool
	port               int
}

func newHandler(skipPrivilegeCheck bool, port int) HandlerContext {
	return HandlerContext{
		skipPrivilegeCheck: skipPrivilegeCheck,
		port:               port,
	}
}

// ServeHTTP says hello from a single Go hardware thread and reveals
// its capabilities.
func (h HandlerContext) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	runtime.LockOSThread()
	// Get some numbers consistent to the current execution, so
	// the returned web page demonstrates that the code execution
	// is bouncing around on different kernel thread ids.
	p := syscall.Getpid()
	t := syscall.Gettid()
	u := syscall.Geteuid()
	c := cap.GetProc()
	runtime.UnlockOSThread()
	log.Printf("Saying hello from proc: %d->%d, caps=%q, euid=%d", p, t, c, u)
	if _, err := fmt.Fprintf(w,
		"Hello from proc: %d->%d, caps=%q, euid=%d\n", p, t, c, u); err != nil {
		log.Printf("Failed to write response")
	}
}

func main() {
	flag.Parse()
	if *port == 0 {
		log.Fatal("please supply --port value")
	}
	h := newHandler(*skipPrivilegeCheck, *port)
	lis, err := h.listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("aborting: %s", err)
	}
	if lis != nil {
		defer func() {
			if err := lis.Close(); err != nil {
				log.Fatalf("unable to close listening socket :%s", err)
			}
		}()
	}
	if err := http.Serve(lis, h); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
