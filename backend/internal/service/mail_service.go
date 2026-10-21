// Package service 实现业务逻辑。
package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"

	"roadside-league/backend/internal/config"
)

// ErrMailNotConfigured 表示 SMTP 未配置。
var ErrMailNotConfigured = errors.New("邮件服务未配置")

// MailService 通过 SMTP 发送邮件。
type MailService struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

// NewMailService 从配置构造邮件服务。
func NewMailService(cfg config.Config) *MailService {
	from := cfg.SMTPFrom
	if from == "" {
		from = cfg.SMTPUser
	}
	return &MailService{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		user:     cfg.SMTPUser,
		password: cfg.SMTPPassword,
		from:     from,
	}
}

// Configured 判断 SMTP 是否已配置。
func (m *MailService) Configured() bool {
	return m.host != "" && m.user != "" && m.password != ""
}

// SendVerificationCode 向指定邮箱发送 6 位注册验证码。
func (m *MailService) SendVerificationCode(to, code string) error {
	if !m.Configured() {
		return ErrMailNotConfigured
	}

	subject := "【路边小站】注册验证码"
	body := fmt.Sprintf(
		"您的注册验证码是：<b>%s</b><br><br>验证码 10 分钟内有效，请勿泄露给他人。<br><br>—— 路边小站",
		code,
	)
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		m.from, to, subject, body,
	)

	addr := net.JoinHostPort(m.host, fmt.Sprintf("%d", m.port))
	auth := smtp.PlainAuth("", m.user, m.password, m.host)

	// 465 端口通常为隐式 TLS（SMTPS）；587 为 STARTTLS。
	if m.port == 465 {
		return m.sendImplicitTLS(addr, to, []byte(msg))
	}
	return smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg))
}

// sendImplicitTLS 用于 465 端口的隐式 TLS 发送。
func (m *MailService) sendImplicitTLS(addr, to string, msg []byte) error {
	tlsConfig := &tls.Config{ServerName: m.host}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return err
	}
	defer client.Close()

	auth := smtp.PlainAuth("", m.user, m.password, m.host)
	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(m.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
