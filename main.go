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
	"os/exec"
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
	mux.HandleFunc("/setup", handleSetup)
	mux.HandleFunc("/", handleRoot)

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", servicePort),
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		},
	}

	fmt.Printf("Keystone Proxy Service running at https://127.0.0.1:%d\n", servicePort)

	// Auto-open browser to setup page if no device ID configured
	if !hasDeviceID() {
		go func() {
			// Brief delay to let the server start
			time.Sleep(500 * time.Millisecond)
			url := fmt.Sprintf("https://127.0.0.1:%d", servicePort)
			fmt.Printf("No device ID configured — opening setup page: %s\n", url)
			exec.Command("open", url).Start()
		}()
	}

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

	deviceID := getDeviceID()
	xml := fmt.Sprintf(`<?xml version="1.0"?>
<device type="c" xmlns="http://www.corelationinc.com/deviceLanguage/v1.0" version="2.0.0.0">
  <deviceInformation type="c">
  <identifier>DEVICE_ID: %s</identifier>
  <userServicePortNumber>%d</userServicePortNumber>
  </deviceInformation>
</device>`, deviceID, servicePort)

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(xml))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	setCORS(w, r)
	if hasDeviceID() {
		deviceID := getDeviceID()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html><head><title>Keystone Proxy Service</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; max-width: 480px; margin: 60px auto; padding: 0 20px; color: #333; }
  .status { background: #e8f5e9; border: 1px solid #a5d6a7; border-radius: 8px; padding: 24px; text-align: center; }
  .status h1 { color: #2e7d32; margin-top: 0; font-size: 20px; }
  .device-id { font-family: monospace; background: #fff; padding: 8px 16px; border-radius: 4px; display: inline-block; margin-top: 8px; font-size: 16px; }
</style></head>
<body><div class="status"><h1>Service Running</h1><p>Device ID:</p><div class="device-id">%s</div></div></body></html>`, deviceID)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(setupPageHTML))
}

func handleSetup(w http.ResponseWriter, r *http.Request) {
	setCORS(w, r)
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	deviceID := strings.TrimSpace(r.FormValue("device_id"))
	if deviceID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	dir := dataDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		http.Error(w, "Failed to create config directory", http.StatusInternalServerError)
		return
	}
	idFile := filepath.Join(dir, "device-id")
	if err := os.WriteFile(idFile, []byte(deviceID+"\n"), 0600); err != nil {
		http.Error(w, "Failed to save device ID", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func hasDeviceID() bool {
	idFile := filepath.Join(dataDir(), "device-id")
	data, err := os.ReadFile(idFile)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) != ""
}

const setupPageHTML = `<!DOCTYPE html>
<html><head><title>Keystone Proxy Service Setup</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; max-width: 480px; margin: 60px auto; padding: 0 20px; color: #333; }
  h1 { font-size: 22px; }
  form { background: #f5f5f5; border-radius: 8px; padding: 24px; }
  label { display: block; font-weight: 600; margin-bottom: 8px; }
  input[type="text"] { width: 100%; padding: 10px; border: 1px solid #ccc; border-radius: 4px; font-size: 16px; font-family: monospace; box-sizing: border-box; }
  button { margin-top: 16px; padding: 10px 24px; background: #1976d2; color: #fff; border: none; border-radius: 4px; font-size: 16px; cursor: pointer; }
  button:hover { background: #1565c0; }
  .help { margin-top: 16px; font-size: 13px; color: #666; line-height: 1.5; }
</style></head>
<body>
<h1>Keystone Proxy Service Setup</h1>
<form method="POST" action="/setup">
  <label for="device_id">Enter your Keystone Device ID</label>
  <input type="text" id="device_id" name="device_id" required placeholder="e.g. my-device-name">
  <button type="submit">Save</button>
</form>
<p class="help">You can find your Device ID in Keystone under<br>
<strong>Configuration Options → Login Information → Identifier</strong><br>
Use the part after <code>DEVICE_ID:</code></p>
</body></html>`

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
	// Check config file first: ~/.keystone-proxy-service/device-id
	idFile := filepath.Join(dataDir(), "device-id")
	if data, err := os.ReadFile(idFile); err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id
		}
	}
	// Use LocalHostName (Bonjour name) — always clean ASCII, no special characters
	out, err := exec.Command("scutil", "--get", "LocalHostName").Output()
	if err == nil {
		name := strings.TrimSpace(string(out))
		if name != "" {
			return name
		}
	}
	// Fall back to hostname
	host, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	if i := strings.IndexByte(host, '.'); i > 0 {
		return host[:i]
	}
	return host
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
