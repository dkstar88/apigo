package runner

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func CountRequestSize(r *http.Request) int {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatal("Cannot count request size, ", fmt.Sprint(err))
		return 0
	}
	return len(dump)
}

func CountResponseSize(r *http.Response) int {
	dump, err := httputil.DumpResponse(r, true)
	if err != nil {
		log.Fatal("Cannot count response size, ", fmt.Sprint(err))
		return 0
	}
	return len(dump)
}
