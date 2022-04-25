package proxy

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer(t *testing.T) {
	var backendCount = 8
	var addressTable []string
	for i := 0; i < backendCount; i++ {
		addr := httptest.NewServer(NoopHandler{}).Listener.Addr().(*net.TCPAddr)
		addressTable = append(addressTable, "http://"+addr.String())
	}
	app := Application{}
	app.Init(addressTable)
	if err := app.Start(9001); err != nil {
		log.Fatal(err)
	}
}

type NoopHandler struct{}

func (np NoopHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}
