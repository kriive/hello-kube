package http

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/kriive/hello-kube"
)

func handleHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, kube!")
	fmt.Fprintf(w, "You are connecting to %s from: %s using %s\n", r.Host, r.RemoteAddr, r.UserAgent())
	hostname, _ := os.Hostname()
	fmt.Fprintf(w, "hostname: %s, runtime.GOOS: %s, runtime.GOARCH: %s\n", hostname, runtime.GOOS, runtime.GOARCH)
	fmt.Fprintln(w, "------------")
	fmt.Fprintf(w, "hello-kube version: %s, commit hash: %s\n", hello.Version, hello.CommitHash)
}
