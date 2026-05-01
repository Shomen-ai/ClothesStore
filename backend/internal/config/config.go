package config

import "os"

type Config struct {
	DBConnStr  string
	JWTSecret  string
	Port       string
	UploadsDir string
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
	return &Config{
		DBConnStr:  os.Getenv("DB_CONN_STR"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
		Port:       port,
		UploadsDir: uploadsDir,
	}
}
