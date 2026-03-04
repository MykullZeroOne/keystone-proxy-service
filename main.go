package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const servicePort = 51763

func main() {
	certDir := filepath.Join(dataDir(), "certs")
	certFile := filepath.Join(certDir, "localhost.crt")
	keyFile := filepath.Join(certDir, "localhost.key")

	tlsCert, err := loadOrGenerateCert(certDir, certFile, keyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "certificate error: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/GetDeviceInformation", handleDeviceInfo)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		setCORS(w, r)
		w.Write([]byte("okay"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", servicePort),
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		},
	}

	fmt.Printf("Keystone Proxy Service running at https://127.0.0.1:%d\n", servicePort)
	if err := server.ListenAndServeTLS("", ""); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func handleDeviceInfo(w http.ResponseWriter, r *http.Request) {
	setCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	macID := getDeviceID()
	xml := fmt.Sprintf(`<?xml version="1.0"?>
<device type="c" xmlns="http://www.corelationinc.com/deviceLanguage/v1.0" version="2.0.0.0">
  <deviceInformation type="c">
  <identifier>MAC: %s</identifier>
  <userServicePortNumber>%d</userServicePortNumber>
  </deviceInformation>
</device>`, macID, servicePort)

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(xml))
}

func setCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func getDeviceID() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	var macs []string
	seen := make(map[string]bool)
	for _, iface := range ifaces {
		mac := iface.HardwareAddr.String()
		if mac == "" || mac == "00:00:00:00:00:00" {
			continue
		}
		// Convert to dash-separated uppercase-ish format matching the Node version
		mac = strings.ReplaceAll(mac, ":", "-")
		if !seen[mac] {
			seen[mac] = true
			macs = append(macs, mac)
		}
	}
	return strings.Join(macs, " ")
}

// dataDir returns ~/.keystone-proxy-service for storing certs and runtime data.
func dataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".keystone-proxy-service")
}

func loadOrGenerateCert(certDir, certFile, keyFile string) (tls.Certificate, error) {
	// Try loading existing cert
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			cert, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err == nil {
				fmt.Println("Loaded existing certificate")
				return cert, nil
			}
		}
	}

	fmt.Println("Generating self-signed certificate for 127.0.0.1 (valid 10 years)...")

	if err := os.MkdirAll(certDir, 0700); err != nil {
		return tls.Certificate{}, fmt.Errorf("create cert dir: %w", err)
	}

	// Generate ECDSA P-256 key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate key: %w", err)
	}

	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create certificate: %w", err)
	}

	// Write cert PEM
	certOut, err := os.Create(certFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("write cert: %w", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certOut.Close()

	// Write key PEM
	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("write key: %w", err)
	}
	keyBytes, _ := x509.MarshalECPrivateKey(privateKey)
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	keyOut.Close()

	fmt.Println("Certificate generated successfully")

	return tls.LoadX509KeyPair(certFile, keyFile)
}
