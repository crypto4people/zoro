package proxy

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/itsabgr/ge"
)

var _ http.Handler = &Server{}

func (server *Server) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	server.http.mux.ServeHTTP(writer, req)
}

func (server *Server) initMux() {

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

	server.http.mux.HandleFunc("PUT /", func(writer http.ResponseWriter, req *http.Request) {
		server.logger.Debug("new request on /", slog.String("address", req.RemoteAddr))

		info, err := server.Info(req.Context())
		if err != nil {
			httpError(writer, err)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(ge.Must(json.Marshal(info)))
	})

	server.http.mux.HandleFunc("PUT /subs", func(writer http.ResponseWriter, req *http.Request) {
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

	server.http.mux.HandleFunc("PUT /eth/balance", func(writer http.ResponseWriter, req *http.Request) {
		addr, err := parseEthAddress(req.URL.Query().Get("address"))
		if err != nil {
			httpError(writer, "failed to parse address")
			return
		}
		bal, err := server.eth.client.BalanceOf(req.Context(), addr)
		if err != nil {
			httpError(writer, "failed to get balance")
			return
		}

		writer.Header().Set("Content-Type", "text/plain")
		io.WriteString(writer, bal.Text(10))

	})

	server.http.mux.HandleFunc("PUT /eth/nonce", func(writer http.ResponseWriter, req *http.Request) {
		addr, err := parseEthAddress(req.URL.Query().Get("address"))
		if err != nil {
			httpError(writer, "failed to parse address")
			return
		}
		non, err := server.eth.client.NonceOf(req.Context(), addr)
		if err != nil {
			httpError(writer, "failed to get balance")
			return
		}
		writer.Header().Set("Content-Type", "text/numb")
		io.WriteString(writer, strconv.FormatUint(non, 10))
	})

	server.http.mux.HandleFunc("PUT /eth/receipt", func(writer http.ResponseWriter, req *http.Request) {
		hash, err := parseEthHash(req.URL.Query().Get("hash"))
		if err != nil {
			httpError(writer, "failed to parse hash")
			return
		}

		tx, pending, err := server.eth.client.Receipt(req.Context(), hash)
		if err != nil {
			httpError(writer, "failed to get balance")
			return
		}

		if tx == nil || server.filterTransaction(tx) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		writer.Header().Set("Content-Type", "application/json")

		if pending {
			writer.WriteHeader(http.StatusCreated)
		} else {
			writer.WriteHeader(http.StatusAccepted)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(ge.Must(json.Marshal(tx)))

	})

	server.http.mux.HandleFunc("PUT /eth/submit", func(writer http.ResponseWriter, req *http.Request) {
		data, err := io.ReadAll(io.LimitReader(req.Body, 1000))
		if err != nil {
			httpError(writer, "failed to get balance")
			return
		}
		tx, err := parseEthTransaction(data)
		if err != nil {
			httpError(writer, "failed to get balance")
			return
		}
		if server.filterTransaction(tx) {
			httpError(writer, "tx is filtered")
			return
		}
		if err = server.eth.client.Submit(req.Context(), tx); err != nil {
			httpError(writer, "failed to submit tx")
			return
		}
		writer.Header().Set("Content-Type", "text/plain")
		io.WriteString(writer, tx.Hash().Hex())
	})

}

func httpError(writer http.ResponseWriter, err any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusBadRequest)
	writer.Write(ge.Must(json.Marshal(struct {
		Error any
	}{err})))
}

func parseEthAddress(str string) (addr common.Address, err error) {
	err = addr.UnmarshalText([]byte(str))
	return addr, err
}

func parseEthHash(str string) (hash common.Hash, err error) {
	err = hash.UnmarshalText([]byte(str))
	return hash, err
}

func parseEthTransaction(b []byte) (*types.Transaction, error) {
	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(b)
	return tx, err
}
