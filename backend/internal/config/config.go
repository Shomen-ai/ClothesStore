package config

import "os"

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

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	uploadsDir := os.Getenv("UPLOADS_DIR")
	if uploadsDir == "" {
		uploadsDir = "./uploads"
	}
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "465"
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
