// todo: refactor
package main

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
	"time"

	"github.com/mholt/acmez/v3"
	"github.com/mholt/acmez/v3/acme"
)

type HTTP01Solver struct {
	Token            string
	KeyAuthorization string
}

func (solver *HTTP01Solver) Present(ctx context.Context, chal acme.Challenge) error {
	solver.Token = chal.Token
	solver.KeyAuthorization = chal.KeyAuthorization
	return nil
}

func (solver *HTTP01Solver) CleanUp(ctx context.Context, chal acme.Challenge) error {
	solver.Token = ""
	solver.KeyAuthorization = ""
	return nil
}

func (solver *HTTP01Solver) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/.well-known/acme-challenge/"+solver.Token {
		http.NotFound(writer, req)
		return
	}
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(solver.KeyAuthorization))
}

func loadOrCreateAccountKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		return rsa.GenerateKey(rand.Reader, 2048)
	}

	b, err := os.ReadFile(path)
	if err == nil {
		block, _ := pem.Decode(b)
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			return nil, fmt.Errorf("invalid PEM in %s", path)
		}

		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err == nil {
			return key, nil
		}

		pk, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 == nil {
			if rk, ok := pk.(*rsa.PrivateKey); ok {
				return rk, nil
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

func obtainCertificateForDomain(domain, datadir string) (string, string, error) {
	certPath := filepath.Join(datadir, "cert.pem")
	keyPath := filepath.Join(datadir, "key.pem")

	if _, err := os.Stat(certPath); err == nil {
		if _, err := os.Stat(certPath); err == nil {
			return certPath, keyPath, nil
		}
	}

	accountKey, err := loadOrCreateAccountKey(filepath.Join(datadir, "acme_account.pem"))
	if err != nil {
		return "", "", fmt.Errorf("load/create account key: %w", err)
	}

	certKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("generate cert key: %w", err)
	}

	solver := &HTTP01Solver{}
	httpServer := &http.Server{
		Addr:    ":80",
		Handler: solver,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()
	time.Sleep(200 * time.Millisecond)
	defer httpServer.Close()

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

	if len(certs[0].ChainPEM) == 0 {
		return "", "", errors.New("empty certificate returned")
	}

	keyPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(certKey)})

	if certPath != "" {
		if err := os.WriteFile(certPath, certs[0].ChainPEM, 0644); err != nil {
			return "", "", fmt.Errorf("write cert file: %w", err)
		}
	}

	if keyPath != "" {
		if err := os.WriteFile(keyPath, keyPem, 0600); err != nil {
			return "", "", fmt.Errorf("write key file: %w", err)
		}
	}

	return "", "", nil
}
