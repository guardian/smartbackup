package postgres

import "fmt"

type DatabaseConfig struct {
	Name      string `yaml:"name"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	DBName    string `yaml:"default_db"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	SSLMode   string `yaml:"ssl_mode"`
	FastStart bool   `yaml:"fast_start"`
}

func (d *DatabaseConfig) GetConnectionString() string {
	connstr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s", d.Host, d.Port, d.User, d.Password, d.DBName)
	if d.SSLMode != "" {
		connstr = connstr + fmt.Sprintf(" sslmode=%s", d.SSLMode)
	}

	return connstr
}
