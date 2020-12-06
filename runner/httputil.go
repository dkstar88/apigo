package runner

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/textproto"
	"strings"
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

func StrToHeaders(str string) http.Header {
	reader := bufio.NewReader(strings.NewReader(str + "\r\n"))
	tp := textproto.NewReader(reader)

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		log.Fatal(err)
	}
	return http.Header(mimeHeader)
}