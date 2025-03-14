# `bbeequeue`

[![PkgGoDev](https://img.shields.io/badge/-reference-blue?logo=go&logoColor=white&labelColor=505050)](https://pkg.go.dev/github.com/thediveo/cpus)
[![GitHub](https://img.shields.io/github/license/thediveo/cpus)](https://img.shields.io/github/license/thediveo/cpus)
![Coverage](https://img.shields.io/badge/Coverage-91.9%25-brightgreen)
![goroutines](https://img.shields.io/badge/go%20routines-not%20leaking-success)

Package `bbeequeue` supports receiving data from type
[`BPF_MAP_TYPE_RINGBUF`](https://docs.ebpf.io/linux/map-type/BPF_MAP_TYPE_RINGBUF/)
eBPF maps using idiomatic Go channels. This package does not attempt to be a
jack-of-all-trades implementation, but instead to cover the basic use case of
providing channel access to eBPF ringbuffer maps.

## Trivia

The package name "bbeequeue" is a terrible pun on eBPF's bee mascot, ringbuffers
or queues, and finally, burning things beyond recognition, also known as "BBQ".

## DevContainer

> [!CAUTION]
>
> Do **not** use VSCode's "~~Dev Containers: Clone Repository in Container
> Volume~~" command, as it is utterly broken by design, ignoring
> `.devcontainer/devcontainer.json`.

1. `git clone https://github.com/thediveo/bbeequeue`
2. in VSCode: Ctrl+Shift+P, "Dev Containers: Open Workspace in Container..."
3. select `bbeequeue.code-workspace` and off you go...

The devcontainer setup includes `bpf2go` and `bpftool`. Run `go generate .` to
(re)generate the eBPF-derived Go source files.

## Supported Go Versions

`bbeequeue` supports versions of Go that are noted by the Go release policy,
that is, major versions _N_ and _N_-1 (where _N_ is the current major version).

## Copyright and License

`bbeequeue` is Copyright 2025 Harald Albrecht, and licensed under the Apache
License, Version 2.0.

The header files in `_headers/` are licensed under a BSD-2-Clause; they
originate from
[`@cilium/ebpf/examples/headers`](https://github.com/cilium/ebpf/tree/main/examples/headers),
with the `libbpf`-originating files having been updated to 1.5.0.

`test.bpf.c` is licensed under a dual MIT/GPL license.
