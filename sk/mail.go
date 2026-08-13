package sk

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

type MailSender struct {
	SMTPHost string
	SMTPPort string
	From     string
	Password string
}

func NewMailSender(host string, port string, from string, password string) *MailSender {
	return &MailSender{
		SMTPHost: host,
		SMTPPort: port,
		From:     from,
		Password: password,
	}
}

func (m *MailSender) SendAll(receivers []string, subject string, content string) error {
	serverAddr := fmt.Sprintf("%s:%s", m.SMTPHost, m.SMTPPort)
	tlsConfig := &tls.Config{ServerName: m.SMTPHost}
	conn, err := tls.Dial("tcp", serverAddr, tlsConfig)
	if err != nil {
		fmt.Println("无法与邮件服务器建立TLS连接:", err)
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.SMTPHost)
	if err != nil {
		fmt.Println("创建邮件客户端失败:", err)
		return err
	}
	defer client.Close()

	auth := smtp.PlainAuth("", m.From, m.Password, m.SMTPHost)
	err = client.Auth(auth)
	if err != nil {
		fmt.Println("邮件服务器认证失败:", err)
		return err
	}

	_ = client.Mail(m.From)
	for _, receiver := range receivers {
		if receiver != "" {
			_ = client.Rcpt(receiver)
		}
	}

	w, err := client.Data()
	if err != nil {
		fmt.Println("发送数据失败:", err)
		return err
	}
	defer w.Close()

	msg := fmt.Sprintf("From: %s\r\nSubject: %s\r\n\r\n%s", m.From, subject, content)
	_, _ = w.Write([]byte(msg))
	_ = w.Close()
	_ = client.Quit()

	return nil
}

func (m *MailSender) Send(receiver string, subject string, content string) error {
	return m.SendAll([]string{receiver}, subject, content)
}
