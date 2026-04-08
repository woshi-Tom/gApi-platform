package service

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"gapi-platform/internal/config"
	"gapi-platform/internal/logger"
)

type EmailMailer struct {
	cfg *config.EmailConfig
}

type EmailConfig = config.EmailConfig

func NewEmailMailer(cfg *config.EmailConfig) *EmailMailer {
	return &EmailMailer{cfg: cfg}
}

func (m *EmailMailer) SendVerificationEmail(toEmail, code, purpose string) error {
	var subject, content string
	if purpose == "reset" {
		return nil
	}

	subject = "[gAPI] 您的注册验证码"
	content = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>注册验证码</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #409eff 0%%, #337ecc 100%%); padding: 30px; text-align: center; border-radius: 12px 12px 0 0;">
        <h1 style="color: white; margin: 0; font-size: 24px;">gAPI Platform</h1>
    </div>
    <div style="background: #fff; padding: 30px; border: 1px solid #e4e7ed; border-top: none; border-radius: 0 0 12px 12px;">
        <h2 style="color: #303133; margin-top: 0;">验证码</h2>
        <p style="color: #606266; font-size: 16px; line-height: 1.6;">
            您好，<br><br>
            您的注册验证码是：
        </p>
        <div style="background: #f5f7fa; padding: 20px; text-align: center; border-radius: 8px; margin: 20px 0;">
            <span style="font-size: 32px; font-weight: bold; color: #409eff; letter-spacing: 8px;">%s</span>
        </div>
        <p style="color: #909399; font-size: 14px;">
            验证码有效期为 <strong>10分钟</strong>，请尽快完成验证。<br>
            如果您没有注册 gAPI 账号，请忽略此邮件。
        </p>
        <hr style="border: none; border-top: 1px solid #ebeef5; margin: 20px 0;">
        <p style="color: #c0c4cc; font-size: 12px;">
            此邮件由系统自动发送，请勿回复。<br>
            为保护您的账户安全，请勿将验证码告知他人。
        </p>
    </div>
</body>
</html>
`, code)

	return m.sendEmail(toEmail, subject, content)
}

func (m *EmailMailer) SendPasswordResetEmail(toEmail, resetLink string) error {
	subject := "[gAPI] 密码重置请求"
	content := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>密码重置</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #67c23a 0%%, #529b2e 100%%); padding: 30px; text-align: center; border-radius: 12px 12px 0 0;">
        <h1 style="color: white; margin: 0; font-size: 24px;">gAPI Platform</h1>
    </div>
    <div style="background: #fff; padding: 30px; border: 1px solid #e4e7ed; border-top: none; border-radius: 0 0 12px 12px;">
        <h2 style="color: #303133; margin-top: 0;">密码重置</h2>
        <p style="color: #606266; font-size: 16px; line-height: 1.6;">
            您好，<br><br>
            我们收到了您的密码重置请求。请点击下方按钮重置密码：
        </p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="display: inline-block; background: linear-gradient(135deg, #67c23a 0%%, #529b2e 100%%); color: white; padding: 14px 40px; text-decoration: none; border-radius: 6px; font-weight: bold;">
                重置密码
            </a>
        </div>
        <p style="color: #909399; font-size: 14px;">
            此链接有效期为 <strong>1小时</strong>。<br>
            如果您没有请求重置密码，请忽略此邮件，您的账户安全不受影响。
        </p>
        <div style="background: #fdf6ec; border: 1px solid #f5dab1; border-radius: 6px; padding: 15px; margin: 20px 0;">
            <p style="color: #e6a23c; font-size: 14px; margin: 0;">
                <strong>安全提示：</strong>链接有效期结束后将自动失效，请妥善保管。
            </p>
        </div>
        <hr style="border: none; border-top: 1px solid #ebeef5; margin: 20px 0;">
        <p style="color: #c0c4cc; font-size: 12px;">
            此邮件由系统自动发送，请勿回复。<br>
            为保护您的账户安全，请勿将重置链接分享给他人。
        </p>
    </div>
</body>
</html>
`, resetLink)

	return m.sendEmail(toEmail, subject, content)
}

func (m *EmailMailer) sendEmail(to, subject, htmlContent string) error {
	if !m.cfg.SMTP.Enabled {
		logger.Debug("SMTP disabled, email not sent",
			"to", logger.RedactEmail(to),
			"subject", subject)
		return nil
	}

	headers := make(map[string]string)
	headers["From"] = m.cfg.SMTP.From
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(htmlContent)

	addr := fmt.Sprintf("%s:%d", m.cfg.SMTP.Host, m.cfg.SMTP.Port)

	var auth smtp.Auth
	if m.cfg.SMTP.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.SMTP.Username, m.cfg.SMTP.Password, m.cfg.SMTP.Host)
	}

	if m.cfg.SMTP.UseTLS {
		return m.sendWithTLS(addr, auth, m.cfg.SMTP.From, []string{to}, message.String())
	}

	return smtp.SendMail(addr, auth, m.cfg.SMTP.Username, []string{to}, []byte(message.String()))
}

func (m *EmailMailer) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: m.cfg.SMTP.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect with TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.cfg.SMTP.Host)
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
