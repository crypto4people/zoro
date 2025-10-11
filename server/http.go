package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/itsabgr/ge"
)

var _ http.Handler = &Server{}

func (server *Server) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	server.http.mux.ServeHTTP(writer, req)
}

func (server *Server) initMux() {
	server.http.mux.HandleFunc("GET /", func(writer http.ResponseWriter, req *http.Request) {
		server.logger.Debug("new request on /", slog.String("address", req.RemoteAddr))
		info := server.Info()
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(ge.Must(json.Marshal(info)))
	})

	server.http.mux.HandleFunc("GET /wsc", func(writer http.ResponseWriter, req *http.Request) {
		user := req.URL.Query().Get("user")
		logger := server.logger.With(slog.String("account", user), slog.String("log_key", uuid.NewString()))
		logger.Debug("new request on /wsc", slog.String("address", req.RemoteAddr))

		err := server.serveProxy(writer, logger, req, common.HexToAddress(user))
		if err != nil {
			httpError(writer, err)
			return
		}
	})

	server.http.mux.HandleFunc("GET /subs", func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		user := req.URL.Query().Get("user")

		server.logger.Debug("new request on /subs", slog.String("address", req.RemoteAddr), slog.String("account", user))
		subs, err := server.ListSubs(ctx, common.HexToAddress(user))
		if err != nil {
			httpError(writer, err)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(ge.Must(json.Marshal(subs)))
	})

	server.http.mux.HandleFunc("GET /process-block", func(writer http.ResponseWriter, req *http.Request) {
		block, err := strconv.ParseInt(req.URL.Query().Get("block"), 10, 64)
		if err != nil {
			httpError(writer, err)
			return
		}

		server.logger.Debug("new request on /subs", slog.String("address", req.RemoteAddr), slog.Int64("block", block))
		changes, err := server.processBlock(req.Context(), block)
		if err != nil {
			httpError(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(ge.Must(json.Marshal(
			map[string]bool{
				"applied": changes > 0,
			},
		)))
	})
}

func httpError(writer http.ResponseWriter, err any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusBadRequest)
	writer.Write(ge.Must(json.Marshal(struct {
		Error any
	}{err})))
}
