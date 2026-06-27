package email

import "crypto/tls"

type tlsConfig struct {
	ServerName string
}

func (t *tlsConfig) standard() *tls.Config {
	return &tls.Config{
		ServerName: t.ServerName,
		MinVersion: tls.VersionTLS12,
	}
}
