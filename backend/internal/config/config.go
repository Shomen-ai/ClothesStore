// Package config loads all runtime configuration from environment variables,
// applying sane defaults so the server can boot in development without a fully
// populated environment.
package config

import "os"

// Config holds every runtime setting the server needs, populated from the
// environment by Load. Some zero values are meaningful: an empty SMTPHost, for
// example, switches the mailer to its stdout/dev fallback (see cmd/server).
type Config struct {
	DBConnStr  string
	JWTSecret  string
	Port       string
	UploadsDir string

	// SMTP — when SMTPHost is empty the mailer falls back to logging
	// the code to stdout, which keeps the registration flow working in dev.
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string
}

// Load reads configuration from the process environment and returns a populated
// Config. Fields that have a sensible fallback (HTTP port, uploads dir, SMTP
// port) default when their env var is unset; the rest are taken verbatim and
// left empty if absent.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default HTTP listen port when PORT is unset
	}
	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads" // local directory served for uploaded product images
	}
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "465" // implicit-TLS submission port (Yandex/Timeweb default)
	}
	return &Config{
		DBConnStr:  os.Getenv("DB_CONN_STR"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		Port:       port,
		UploadsDir: uploadsDir,
		SMTPHost:   os.Getenv("SMTP_HOST"),
		SMTPPort:   smtpPort,
		SMTPUser:   os.Getenv("SMTP_USER"),
		SMTPPass:   os.Getenv("SMTP_PASS"),
		SMTPFrom:   os.Getenv("SMTP_FROM"),
	}
}
