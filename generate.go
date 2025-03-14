//go:generate bpf2go -go-package bbeequeue_test -output-suffix _test bbq test.bpf.c -- -I./_headers

package bbeequeue
