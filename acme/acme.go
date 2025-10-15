package acme

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mholt/acmez/v3"
	"github.com/mholt/acmez/v3/acme"
)

const (
	renewBefore = 7 * 24 * time.Hour
)

type HTTP01Solver struct {
	mu               sync.RWMutex
	Token            string
	KeyAuthorization string
}

func (s *HTTP01Solver) Present(ctx context.Context, chal acme.Challenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Token = chal.Token
	s.KeyAuthorization = chal.KeyAuthorization
	return nil
}

func (s *HTTP01Solver) CleanUp(ctx context.Context, chal acme.Challenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Token == chal.Token {
		s.Token = ""
		s.KeyAuthorization = ""
	}
	return nil
}

func (s *HTTP01Solver) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.mu.RLock()
	token := s.Token
	keyAuth := s.KeyAuthorization
	s.mu.RUnlock()

	expectedPath := "/.well-known/acme-challenge/" + token
	if req.URL.Path != expectedPath || token == "" {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(keyAuth))
}

func loadOrCreateAccountKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		return rsa.GenerateKey(rand.Reader, 2048)
	}

	if b, err := os.ReadFile(path); err == nil {
		block, _ := pem.Decode(b)
		if block != nil {
			switch block.Type {
			case "RSA PRIVATE KEY":
				if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
					return key, nil
				}
			default:
				if pk, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
					if rk, ok := pk.(*rsa.PrivateKey); ok {
						return rk, nil
					}
				}
			}
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(path, keyPem, 0600); err != nil {
		return nil, fmt.Errorf("failed to write account key: %w", err)
	}
	return key, nil
}

func certLooksValid(certPath, keyPath, domain string) error {
	now := time.Now()

	// read cert
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("read cert: %w", err)
	}
	var certs []*x509.Certificate
	for {
		var block *pem.Block
		block, certPEM = pem.Decode(certPEM)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			c, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return fmt.Errorf("parse certificate: %w", err)
			}
			certs = append(certs, c)
		}
	}
	if len(certs) == 0 {
		return errors.New("no certificate PEM blocks found")
	}
	leaf := certs[0]

	if leaf.NotAfter.Before(now.Add(renewBefore)) {
		return fmt.Errorf("certificate expires soon or already expired (NotAfter=%v)", leaf.NotAfter)
	}

	ok := false
	for _, n := range leaf.DNSNames {
		if strings.EqualFold(n, domain) {
			ok = true
			break
		}
	}
	if !ok && strings.EqualFold(leaf.Subject.CommonName, domain) {
		ok = true
	}
	if !ok {
		return fmt.Errorf("certificate does not contain domain %q in SANs or CN", domain)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("read key: %w", err)
	}
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return errors.New("invalid private key PEM")
	}

	var privKey interface{}
	switch block.Type {
	case "RSA PRIVATE KEY":
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("parse rsa private key: %w", err)
		}
	default:
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("unsupported private key type %q and parse failed: %w", block.Type, err)
		}
		privKey = parsed
	}

	switch pk := privKey.(type) {
	case *rsa.PrivateKey:
		rsaPub, ok := leaf.PublicKey.(*rsa.PublicKey)
		if !ok {
			return errors.New("certificate public key is not RSA")
		}
		if rsaPub.N.Cmp(pk.PublicKey.N) != 0 || rsaPub.E != pk.PublicKey.E {
			return errors.New("private key does not match certificate public key")
		}
	default:
		return errors.New("private key type not supported for matching (expect RSA)")
	}

	return nil
}

func ObtainCertificateForDomain(domain, datadir string) (certPath, keyPath string, err error) {
	if datadir == "" {
		return "", "", errors.New("datadir required")
	}
	if domain == "" {
		return "", "", errors.New("domain required")
	}

	certPath = filepath.Join(datadir, "cert.pem")
	keyPath = filepath.Join(datadir, "key.pem")
	accountKeyPath := filepath.Join(datadir, "acme_account.pem")

	if _, errStat := os.Stat(certPath); errStat == nil {
		if _, errStat2 := os.Stat(keyPath); errStat2 == nil {
			if err := certLooksValid(certPath, keyPath, domain); err == nil {
				return certPath, keyPath, nil
			} else {
				log.Printf("existing cert/key not usable: %v (will attempt to obtain new)", err)
			}
		}
	}

	if err := os.MkdirAll(datadir, 0700); err != nil {
		return "", "", fmt.Errorf("mkdir datadir: %w", err)
	}

	accountKey, err := loadOrCreateAccountKey(accountKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("load/create account key: %w", err)
	}

	certKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("generate cert key: %w", err)
	}

	solver := &HTTP01Solver{}
	server := &http.Server{
		Addr:    ":80",
		Handler: solver,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
		close(serverErrCh)
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = server.Shutdown(ctx)
		cancel()
	}()

	time.Sleep(200 * time.Millisecond)

	client := &acmez.Client{
		Client: &acme.Client{
			Directory: "https://acme-v02.api.letsencrypt.org/directory",
		},
		ChallengeSolvers: map[string]acmez.Solver{
			acme.ChallengeTypeHTTP01: solver,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	account, err := client.NewAccount(ctx, acme.Account{
		TermsOfServiceAgreed: true,
		PrivateKey:           accountKey,
	})
	if err != nil {
		return "", "", fmt.Errorf("register account: %w", err)
	}

	csrObj, err := acmez.NewCSR(certKey, []string{domain})
	if err != nil {
		return "", "", fmt.Errorf("acmez.NewCSR: %w", err)
	}
	if len(csrObj.DNSNames) == 0 {
		return "", "", fmt.Errorf("CSR has no DNSNames (domain: %s)", domain)
	}

	orderParams := acmez.OrderParameters{
		Account: account,
		CSR:     acmez.StaticCSR(csrObj),
		Identifiers: []acme.Identifier{
			{Type: "dns", Value: domain},
		},
	}

	certs, err := client.ObtainCertificate(ctx, orderParams)
	if err != nil {
		return "", "", fmt.Errorf("obtain certificate: %w", err)
	}
	if len(certs) == 0 || len(certs[0].ChainPEM) == 0 {
		return "", "", errors.New("empty certificate returned")
	}

	if err := os.WriteFile(certPath, certs[0].ChainPEM, 0644); err != nil {
		return "", "", fmt.Errorf("write cert file: %w", err)
	}

	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(certKey)})
	if err := os.WriteFile(keyPath, keyPem, 0600); err != nil {
		return "", "", fmt.Errorf("write key file: %w", err)
	}

	if err := certLooksValid(certPath, keyPath, domain); err != nil {
		return "", "", fmt.Errorf("post-obtain validation failed: %w", err)
	}

	return certPath, keyPath, nil
}
