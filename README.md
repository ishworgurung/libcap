# Description

This is the original work of Andrew G. Morgan and numerous others. Much thanks.

Until, https://go-review.googlesource.com/c/go/+/210639/ is merged, this can be used as a drop-in CGO replacement to 
achieve many things amongst which one is a *proper* privilege seperation in Go.

It is nicely explained by Andrew G. Morgan in [web.go](examples/web.go#L13).

#### Build

Install `libcap` (which should also include `libpsx`).

Tested on `libcap` version 2.33

The source for `libcap` can be found at https://git.kernel.org/pub/scm/libs/libcap/libcap.git

To build the sample web server:

```fish
$ env CGO_ENABLED="1"            \
CGO_LDFLAGS_ALLOW="-Wl,-wrap,.+" \
    go build -o web web.go
```

#### Run un-privileged binary

```fish
$ ./web --port=80
2020/03/22 14:58:32 aborting: insufficient privilege to bind to low ports - want "cap_net_bind_service", have "="
$ sudo setcap cap_setpcap,cap_net_bind_service=+p ./web
$ ./web --port=80
2020/03/22 15:30:39 Saying hello from proc: 45869->45869, caps="=", euid=1000
$ curl -s localhost:80
Hello from proc: 45869->45869, caps="=", euid=1000
```

#### Skip privilege checks

```fish
$ ./web --port=8080 --skip
2020/03/22 16:05:16 Saying hello from proc: 56453->56455, caps="=", euid=1000
$ curl -s localhost:8080
Hello from proc: 56453->56455, caps="=", euid=1000
```

#### Run as root user
```fish
$ sudo ./web --port=80
2020/03/29 16:10:15 go runtime is running as root - cheating
```
