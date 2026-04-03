package mqtt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientConfigDefaults(t *testing.T) {
	cfg := ClientConfig{
		Broker:      "tcp://localhost:1883",
		ClientID:    "test-client",
		TopicPrefix: "test-prefix",
	}

	assert.Equal(t, "tcp://localhost:1883", cfg.Broker)
	assert.Equal(t, "test-client", cfg.ClientID)
	assert.Equal(t, "test-prefix", cfg.TopicPrefix)
}

func TestMockPublisher(t *testing.T) {
	mock := NewMockPublisher()
	assert.True(t, mock.IsConnected())

	err := mock.Publish("test/topic", 1, true, []byte("hello"))
	assert.NoError(t, err)
	assert.Len(t, mock.Published, 1)
	assert.Equal(t, "test/topic", mock.Published[0].Topic)
	assert.Equal(t, []byte("hello"), mock.Published[0].Payload)

	err = mock.Subscribe("test/sub", 1, func(topic string, payload []byte) {})
	assert.NoError(t, err)
	assert.Len(t, mock.Subscribed, 1)

	mock.Disconnect()
	assert.False(t, mock.IsConnected())
}

func TestBuildTLSConfigSkipVerify(t *testing.T) {
	cfg := TLSConfig{
		Enabled:    true,
		SkipVerify: true,
	}

	tlsCfg, err := buildTLSConfig(cfg)
	require.NoError(t, err)
	assert.True(t, tlsCfg.InsecureSkipVerify)
	assert.Nil(t, tlsCfg.RootCAs)
	assert.Empty(t, tlsCfg.Certificates)
}

func TestBuildTLSConfigWithCACert(t *testing.T) {
	// Generate a self-signed CA cert for testing
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	generateTestCACert(t, caPath)

	cfg := TLSConfig{
		Enabled: true,
		CACert:  caPath,
	}

	tlsCfg, err := buildTLSConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, tlsCfg.RootCAs)
	assert.False(t, tlsCfg.InsecureSkipVerify)
}

func TestBuildTLSConfigWithClientCert(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "client.pem")
	keyPath := filepath.Join(dir, "client-key.pem")
	generateTestClientCert(t, certPath, keyPath)

	cfg := TLSConfig{
		Enabled:    true,
		ClientCert: certPath,
		ClientKey:  keyPath,
	}

	tlsCfg, err := buildTLSConfig(cfg)
	require.NoError(t, err)
	assert.Len(t, tlsCfg.Certificates, 1)
}

func TestBuildTLSConfigInvalidCACert(t *testing.T) {
	cfg := TLSConfig{
		Enabled: true,
		CACert:  "/nonexistent/ca.pem",
	}

	_, err := buildTLSConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading CA cert")
}

func TestBuildTLSConfigInvalidCACertContent(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "bad-ca.pem")
	require.NoError(t, os.WriteFile(caPath, []byte("not a cert"), 0o644))

	cfg := TLSConfig{
		Enabled: true,
		CACert:  caPath,
	}

	_, err := buildTLSConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse CA cert")
}

func TestBuildTLSConfigInvalidClientCert(t *testing.T) {
	cfg := TLSConfig{
		Enabled:    true,
		ClientCert: "/nonexistent/client.pem",
		ClientKey:  "/nonexistent/client-key.pem",
	}

	_, err := buildTLSConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading client cert/key")
}

func TestBuildTLSConfigFullSetup(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	certPath := filepath.Join(dir, "client.pem")
	keyPath := filepath.Join(dir, "client-key.pem")

	generateTestCACert(t, caPath)
	generateTestClientCert(t, certPath, keyPath)

	cfg := TLSConfig{
		Enabled:    true,
		CACert:     caPath,
		ClientCert: certPath,
		ClientKey:  keyPath,
		SkipVerify: false,
	}

	tlsCfg, err := buildTLSConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, tlsCfg.RootCAs)
	assert.Len(t, tlsCfg.Certificates, 1)
	assert.False(t, tlsCfg.InsecureSkipVerify)
}

// --- test helpers for generating certs ---

func generateTestCACert(t *testing.T, path string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "Test CA"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		IsCA:         true,
		KeyUsage:     x509.KeyUsageCertSign,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	f, err := os.Create(path)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	require.NoError(t, pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}))
}

func generateTestClientCert(t *testing.T, certPath, keyPath string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "Test Client"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certFile, err := os.Create(certPath)
	require.NoError(t, err)
	defer func() { _ = certFile.Close() }()
	require.NoError(t, pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}))

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	keyFile, err := os.Create(keyPath)
	require.NoError(t, err)
	defer func() { _ = keyFile.Close() }()
	require.NoError(t, pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
}
