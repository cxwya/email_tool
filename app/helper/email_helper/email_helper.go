package email_helper

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

// EmailConfig SMTP 配置
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// EmailMessage 邮件内容
type EmailMessage struct {
	To      []string // 收件人列表
	Cc      []string // 抄送列表
	Subject string   // 邮件主题
	Body    string   // 邮件正文
	IsHTML  bool     // 是否为HTML格式
}

// EmailResult 发送结果
type EmailResult struct {
	Success bool
	Error   string
}

// GetDefaultConfig 从环境变量获取默认配置
func GetDefaultConfig() EmailConfig {
	portStr := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 587
	}
	return EmailConfig{
		Host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		Port:     port,
		Username: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		Password: strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		From:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
		FromName: strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
	}
}

// cleanHeader 彻底清除导致 Header 截断的非法换行符（修复核心）
func cleanHeader(in string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(in, "\r", ""), "\n", ""))
}

// SendEmail 发送邮件
func SendEmail(config EmailConfig, message EmailMessage) EmailResult {
	if config.Host == "" {
		return EmailResult{Success: false, Error: "SMTP Host 未配置"}
	}
	if len(message.To) == 0 {
		return EmailResult{Success: false, Error: "收件人不能为空"}
	}

	var buf bytes.Buffer

	// Go 底层 smtp.Data() 会自动把 \n 转换为规范的 \r\n，手动写 \r\n 遇到特殊环境会变成 \r\r\n 导致信头破裂
	hostname := "localhost"
	if parts := strings.Split(config.Host, ":"); len(parts) > 0 && parts[0] != "" {
		hostname = parts[0]
	}
	buf.WriteString(fmt.Sprintf("Message-ID: <%d@%s>\n", time.Now().UnixNano(), hostname))
	buf.WriteString(fmt.Sprintf("Date: %s\n", time.Now().Format(time.RFC1123Z)))

	fromName := cleanHeader(config.FromName)
	fromEmail := cleanHeader(config.From)
	if fromName != "" {
		buf.WriteString(fmt.Sprintf("From: %s <%s>\n", mime.BEncoding.Encode("UTF-8", fromName), fromEmail))
	} else {
		buf.WriteString(fmt.Sprintf("From: %s\n", fromEmail))
	}

	var cleanTo []string
	for _, to := range message.To {
		cleanTo = append(cleanTo, cleanHeader(to))
	}
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(cleanTo, ",")))

	if len(message.Cc) > 0 {
		var cleanCc []string
		for _, cc := range message.Cc {
			cleanCc = append(cleanCc, cleanHeader(cc))
		}
		buf.WriteString(fmt.Sprintf("Cc: %s\n", strings.Join(cleanCc, ",")))
	}

	buf.WriteString(fmt.Sprintf("Subject: %s\n", mime.BEncoding.Encode("UTF-8", cleanHeader(message.Subject))))
	buf.WriteString("MIME-Version: 1.0\n")

	if message.IsHTML {
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\n")
	}

	buf.WriteString("Content-Transfer-Encoding: base64\n")

	// 唯一的空行，严格分隔 Header 和 Body
	buf.WriteString("\n")

	// Body Base64 每 76 字符换行
	encodedBody := base64.StdEncoding.EncodeToString([]byte(message.Body))
	for i := 0; i < len(encodedBody); i += 76 {
		end := i + 76
		if end > len(encodedBody) {
			end = len(encodedBody)
		}
		buf.WriteString(encodedBody[i:end] + "\n")
	}

	allRecipients := append(cleanTo, message.Cc...)
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	var err error
	if config.Port == 465 {
		err = sendMailWithSSL(addr, auth, fromEmail, allRecipients, buf.Bytes())
	} else {
		err = smtp.SendMail(addr, auth, fromEmail, allRecipients, buf.Bytes())
	}

	if err != nil {
		return EmailResult{Success: false, Error: err.Error()}
	}

	return EmailResult{Success: true, Error: ""}
}

// sendMailWithSSL 使用 SSL 发送邮件 (端口465)
func sendMailWithSSL(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host := strings.Split(addr, ":")[0]

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// SendEmailWithDefaultConfig 使用默认配置发送邮件
func SendEmailWithDefaultConfig(message EmailMessage) EmailResult {
	config := GetDefaultConfig()
	return SendEmail(config, message)
}
