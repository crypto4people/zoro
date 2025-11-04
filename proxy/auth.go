package proxy

import (
	"errors"
	"net"

	"github.com/crypto4people/zoro/crypto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func (server *Server) authConn(wsConn net.Conn, addr common.Address, network, host string) (err error) {
	//todo check timeout

	challenge := crypto.NewChallenge()

	if err = wsutil.WriteServerMessage(wsConn, ws.OpBinary, challenge); err != nil {
		return err
	}

	sig, err := wsutil.ReadClientBinary(wsConn)
	if err != nil {
		return err
	}

	if !crypto.VerifyChallenge(challenge, addr, network, host, sig) {
		return errors.New("unauthorized")
	}

	return nil
}
