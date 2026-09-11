package clients

import (
	"strings"
	"testing"
)

func TestBuildDSN_PasswordAuth(t *testing.T) {
	dsn, err := buildDSN(FactoryConfig{
		Host:     "db.internal",
		Username: "cruisekube",
		Database: "cruisekube",
		Password: "hunter2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"host=db.internal", "user=cruisekube", "dbname=cruisekube", "port=5432", "sslmode=disable", "password=hunter2"} {
		if !strings.Contains(dsn, want) {
			t.Errorf("expected DSN to contain %q, got %q", want, dsn)
		}
	}
	if strings.Contains(dsn, "sslcert=") || strings.Contains(dsn, "sslkey=") {
		t.Errorf("password-auth DSN must not include client-certificate params, got %q", dsn)
	}
}

func TestBuildDSN_ClientCertificateAuth(t *testing.T) {
	dsn, err := buildDSN(FactoryConfig{
		Host:        "db.internal",
		Username:    "cruisekube",
		Database:    "cruisekube",
		SSLCert:     "/certs/tls.crt",
		SSLKey:      "/certs/tls.key",
		SSLRootCert: "/certs/ca.crt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"sslcert=/certs/tls.crt", "sslkey=/certs/tls.key", "sslrootcert=/certs/ca.crt"} {
		if !strings.Contains(dsn, want) {
			t.Errorf("expected DSN to contain %q, got %q", want, dsn)
		}
	}
	if strings.Contains(dsn, "password=") {
		t.Errorf("certificate-auth DSN must not include a password when none is set, got %q", dsn)
	}
}

func TestBuildDSN_ClientCertificateAuthWithoutRootCert(t *testing.T) {
	dsn, err := buildDSN(FactoryConfig{
		Host:     "db.internal",
		Username: "cruisekube",
		Database: "cruisekube",
		SSLCert:  "/certs/tls.crt",
		SSLKey:   "/certs/tls.key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(dsn, "sslrootcert=") {
		t.Errorf("expected no sslrootcert param when SSLRootCert is unset, got %q", dsn)
	}
}

func TestBuildDSN_RejectsMissingCredentials(t *testing.T) {
	if _, err := buildDSN(FactoryConfig{Host: "db.internal", Username: "cruisekube", Database: "cruisekube"}); err == nil {
		t.Fatal("expected an error when neither Password nor SSLCert/SSLKey is set")
	}
}

func TestBuildDSN_RejectsPartialClientCertificate(t *testing.T) {
	if _, err := buildDSN(FactoryConfig{Host: "db.internal", Username: "cruisekube", Database: "cruisekube", SSLCert: "/certs/tls.crt"}); err == nil {
		t.Fatal("expected an error when SSLCert is set without SSLKey")
	}
	if _, err := buildDSN(FactoryConfig{Host: "db.internal", Username: "cruisekube", Database: "cruisekube", SSLKey: "/certs/tls.key"}); err == nil {
		t.Fatal("expected an error when SSLKey is set without SSLCert")
	}
}

func TestBuildDSN_Defaults(t *testing.T) {
	dsn, err := buildDSN(FactoryConfig{Username: "cruisekube", Database: "cruisekube", Password: "hunter2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"host=localhost", "port=5432", "sslmode=disable"} {
		if !strings.Contains(dsn, want) {
			t.Errorf("expected default %q in DSN, got %q", want, dsn)
		}
	}
}
