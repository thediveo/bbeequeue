//go:build ignore

#include "common.h"

char __license[] SEC("license") = "Dual MIT/GPL";

struct event {
    u64 magic;
    u64 inverse_magic;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, u32);
    __type(value, u32);
    __uint(max_entries, 42);
} map_men SEC(".maps");

// see: https://docs.ebpf.io/linux/map-type/BPF_MAP_TYPE_RINGBUF/
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 4096);
    __type(value, struct event);
} events SEC(".maps");

struct emit_prog_args {
    u64 magic;
};

// see: https://docs.ebpf.io/linux/program-type/BPF_PROG_TYPE_SYSCALL/
SEC("syscall")
int emit(struct emit_prog_args *ctx) {
    struct event *ev = bpf_ringbuf_reserve(&events, sizeof(struct event), 0);
    if (ev == NULL) {
        return 42;
    }
    ev->magic = ctx->magic;
    ev->inverse_magic = ~(ctx->magic);
    bpf_ringbuf_submit(ev, 0);

    return 0;
}
