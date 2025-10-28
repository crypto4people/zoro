package proxy

import (
	"crypto/ecdsa"
	"errors"
	"net"
	"zoro/crypto"

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

func SolveChallenge(wsConn net.Conn, private *ecdsa.PrivateKey, network, host string) error {
	//todo check timeout

	challenge, err := wsutil.ReadServerBinary(wsConn)
	if err != nil {
		return err
	}

	sig := crypto.SignChallenge(private, challenge, network, host)

	return wsutil.WriteClientMessage(wsConn, ws.OpBinary, sig)

}
