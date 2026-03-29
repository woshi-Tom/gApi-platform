package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

type testMailer struct {
	host      string
	port      int
	username  string
	password  string
	useTLS    bool
	fromName  string
	fromEmail string
}

func (m *testMailer) sendTestEmail(to string) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	subject := "[gAPI] 邮箱配置测试"
	content := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>邮箱配置测试</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #67c23a 0%%, #529b2e 100%%); padding: 30px; text-align: center; border-radius: 12px 12px 0 0;">
        <h1 style="color: white; margin: 0; font-size: 24px;">gAPI Platform</h1>
    </div>
    <div style="background: #fff; padding: 30px; border: 1px solid #e4e7ed; border-top: none; border-radius: 0 0 12px 12px;">
        <h2 style="color: #303133; margin-top: 0;">邮箱配置测试</h2>
        <p style="color: #606266; font-size: 16px; line-height: 1.6;">
            您好，<br><br>
            如果您收到这封邮件，说明您的 gAPI Platform 邮箱配置正确无误！
        </p>
        <div style="background: #f0f9eb; border: 1px solid #e1f3d8; border-radius: 6px; padding: 15px; margin: 20px 0;">
            <p style="color: #67c23a; font-size: 14px; margin: 0;">
                <strong>✓ 配置成功</strong><br>
                您的 SMTP 邮箱设置已通过测试验证。
            </p>
        </div>
        <hr style="border: none; border-top: 1px solid #ebeef5; margin: 20px 0;">
        <p style="color: #c0c4cc; font-size: 12px;">
            此邮件由系统自动发送，请勿回复。
        </p>
    </div>
</body>
</html>
`)

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", m.fromName, m.fromEmail)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(content)

	if m.useTLS {
		return m.sendWithTLS(addr, auth, m.fromEmail, []string{to}, message.String())
	}

	return smtp.SendMail(addr, auth, m.username, []string{to}, []byte(message.String()))
}

func (m *testMailer) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: m.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect with TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth failed: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL command failed: %w", err)
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("SMTP RCPT command failed: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA command failed: %w", err)
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close email writer: %w", err)
	}

	return client.Quit()
}
