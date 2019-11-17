package mail

type MailConfig struct {
	SMTPServer  string   `yaml:"smtpserver"`
	SMTPUser    string   `yaml:"smtpuser"`
	SMTPPasswd  string   `yaml:"smtppassword"`
	FromAddress string   `yaml:"mailfrom"`
	AlwaysSend  bool     `yaml:"sendonsuccess"`
	SendTo      []string `yaml:"mailto"`
}
