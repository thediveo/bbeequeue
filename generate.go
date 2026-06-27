//go:generate bpf2go -verbose -go-package bbeequeue_test -output-suffix _test bbq test.bpf.c -- -I./_headers

package bbeequeue
