# Description

This is the work of Andrew G. Morgan and numerous others.

Until, https://go-review.googlesource.com/c/go/+/210639/ is merged,
this can be used as a drop-in replacement which Andrew G. Morgan clearly
explains in [web.go](examples/web.go#L13).

#### Demo build

Install `libcap` and `libpsx`.

To build example example:

```fish
$ env CGO_ENABLED="1"            \
CGO_LDFLAGS_ALLOW="-Wl,-wrap,.+" \
    go build -o web web.go
```

#### Run un-privilege binary

```fish
$ ./web --port=80
2020/03/22 14:58:32 aborting: insufficient privilege to bind to low ports - want "cap_net_bind_service", have "="
$ sudo setcap cap_setpcap,cap_net_bind_service=+p ./web
$ ./web --port=80
2020/03/22 14:59:40 Saying hello from proc: 38671->38671, caps="="
$ watch -d -n1 curl -s localhost:80
Hello from proc: 38671->38671, caps="="
Hello from proc: 38671->38675, caps="="
```

