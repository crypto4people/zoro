package proxy

import (
	"encoding/json"
	"fmt"
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
	server.httpMux.ServeHTTP(writer, req)
}

func (server *Server) initMux() {
	server.httpMux.HandleFunc("GET /wsc", handlerFuncWithError(func(writer http.ResponseWriter, req *http.Request) (int, any) {
		user := req.URL.Query().Get("user")
		logger := server.logger.With(slog.String("account", user), slog.String("log_key", uuid.NewString()))
		logger.Debug("new request on /wsc", slog.String("address", req.RemoteAddr))

		err := server.serveProxy(writer, logger, req, common.HexToAddress(user))
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return 200, nil
	}))

	server.httpMux.HandleFunc("PUT /", handlerFuncWithError(func(writer http.ResponseWriter, req *http.Request) (int, any) {
		server.logger.Debug("new request on /", slog.String("address", req.RemoteAddr))

		info, err := server.Info(req.Context())
		if err != nil {
			return http.StatusInternalServerError, err
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(ge.Must(json.Marshal(info)))
		return 200, nil
	}))

	server.httpMux.HandleFunc("PUT /subs", handlerFuncWithError(func(writer http.ResponseWriter, req *http.Request) (int, any) {
		ctx := req.Context()
		user := req.URL.Query().Get("user")

		server.logger.Debug("new request on /subs", slog.String("address", req.RemoteAddr), slog.String("account", user))
		subs, err := server.ListSubs(ctx, common.HexToAddress(user))
		if err != nil {
			return http.StatusBadRequest, err
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(ge.Must(json.Marshal(subs)))
		return 200, nil
	}))

	server.httpMux.HandleFunc("PUT /eth/balance", handlerFuncWithError(func(writer http.ResponseWriter, req *http.Request) (int, any) {
		addr, err := parseEthAddress(req.URL.Query().Get("address"))
		if err != nil {
			return http.StatusBadRequest, "failed to parse address"
		}

		bal, err := server.ethClient.BalanceOf(req.Context(), addr)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		writer.Header().Set("Content-Type", "text/plain")
		io.WriteString(writer, bal.Text(10))
		return 200, nil
	}))

	server.httpMux.HandleFunc("PUT /eth/nonce", handlerFuncWithError(func(writer http.ResponseWriter, req *http.Request) (int, any) {
		addr, err := parseEthAddress(req.URL.Query().Get("address"))
		if err != nil {
			return http.StatusBadRequest, "failed to parse address"
		}

		non, err := server.ethClient.NonceOf(req.Context(), addr)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		writer.Header().Set("Content-Type", "text/numb")
		io.WriteString(writer, strconv.FormatUint(non, 10))
		return 200, nil
	}))

	server.httpMux.HandleFunc("PUT /eth/receipt", handlerFuncWithError(func(writer http.ResponseWriter, req *http.Request) (int, any) {
		hash, err := parseEthHash(req.URL.Query().Get("hash"))
		if err != nil {
			return http.StatusBadRequest, "failed to parse hash"
		}

		tx, pending, err := server.ethClient.Receipt(req.Context(), hash)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if tx == nil {
			return http.StatusNotFound, "tx not found"
		}

		writer.Header().Set("Content-Type", "application/json")

		if pending {
			writer.WriteHeader(http.StatusCreated)
		} else {
			writer.WriteHeader(http.StatusAccepted)
		}

		writer.Write(ge.Must(json.Marshal(tx)))
		return 200, nil
	}))

	server.httpMux.HandleFunc("PUT /eth/submit", handlerFuncWithError(func(writer http.ResponseWriter, req *http.Request) (int, any) {
		data, err := io.ReadAll(io.LimitReader(req.Body, 1000))
		if err != nil {
			return http.StatusBadRequest, "failed to read body"
		}

		tx, err := parseEthTransaction(data)
		if err != nil {
			return http.StatusBadRequest, "failed to parse tx"
		}

		if err = server.ethClient.Submit(req.Context(), tx); err != nil {
			return http.StatusInternalServerError, err
		}

		writer.Header().Set("Content-Type", "text/plain")
		io.WriteString(writer, tx.Hash().Hex())
		return 200, nil
	}))

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

func handlerFuncWithError(fn func(writer http.ResponseWriter, req *http.Request) (int, any)) func(writer http.ResponseWriter, req *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		code, err := fn(writer, req)
		if err != nil {
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(code)
			writer.Write(ge.Must(json.Marshal(struct {
				Error string
			}{fmt.Sprint(err)})))
		}
	}
}
