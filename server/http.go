package server

import (
	"encoding/json"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/itsabgr/ge"
)

var _ http.Handler = &Server{}

func (server *Server) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	server.http.mux.ServeHTTP(writer, req)
}

func (server *Server) initMux() {
	server.http.mux.HandleFunc("GET /", func(writer http.ResponseWriter, req *http.Request) {
		info := server.Info()
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(ge.Must(json.Marshal(info)))
	})

	server.http.mux.HandleFunc("PUT /", func(writer http.ResponseWriter, req *http.Request) {
		err := server.serveProxy(writer, req, common.HexToAddress(req.URL.Query().Get("user")))
		if err != nil {
			httpError(writer, err)
			return
		}
	})

	server.http.mux.HandleFunc("GET /subs", func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		subs, err := server.ListSubs(ctx, common.HexToAddress(req.URL.Query().Get("user")))
		if err != nil {
			httpError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(ge.Must(json.Marshal(subs)))
	})

}

func httpError(writer http.ResponseWriter, err any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusBadRequest)
	writer.Write(ge.Must(json.Marshal(struct {
		Error any
	}{err})))
}
