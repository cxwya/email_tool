package email_helper

import (
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
		// 使用 TrimSpace 彻底清除环境变量中可能存在的 \r 等隐藏空白符
		Host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		Port:     port,
		Username: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		Password: strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		From:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
		FromName: strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
	}
}

// SendEmail 发送邮件
func SendEmail(config EmailConfig, message EmailMessage) EmailResult {
	if config.Host == "" {
		return EmailResult{Success: false, Error: "SMTP Host 未配置"}
	}
	if len(message.To) == 0 {
		return EmailResult{Success: false, Error: "收件人不能为空"}
	}

	var msgBuilder strings.Builder

	// 1. 生成 Message-ID，防止部分严格服务器降级报文
	hostname := "localhost"
	if parts := strings.Split(config.Host, ":"); len(parts) > 0 && parts[0] != "" {
		hostname = parts[0]
	}
	msgBuilder.WriteString(fmt.Sprintf("Message-ID: <%d@%s>\r\n", time.Now().UnixNano(), hostname))
	msgBuilder.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))

	// 2. 发件人
	fromName := strings.TrimSpace(config.FromName)
	from := strings.TrimSpace(config.From)
	if fromName != "" {
		encodedFromName := mime.BEncoding.Encode("UTF-8", fromName)
		msgBuilder.WriteString(fmt.Sprintf("From: %s <%s>\r\n", encodedFromName, from))
	} else {
		msgBuilder.WriteString(fmt.Sprintf("From: %s\r\n", from))
	}

	// 3. 收件人
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(message.To, ",")))

	// 4. 抄送人
	if len(message.Cc) > 0 {
		msgBuilder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(message.Cc, ",")))
	}

	// 5. 邮件主题
	subject := strings.TrimSpace(message.Subject)
	encodedSubject := mime.BEncoding.Encode("UTF-8", subject)
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))

	// 6. MIME 版本与类型
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")
	if message.IsHTML {
		msgBuilder.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	} else {
		msgBuilder.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	}

	// 7. 声明 Base64 传输编码
	msgBuilder.WriteString("Content-Transfer-Encoding: base64\r\n")

	// 8. 严格的且唯一的空行分隔符
	msgBuilder.WriteString("\r\n")

	// 9. 正文 Base64 编码分块，每 76 个字符进行换行
	encodedBody := base64.StdEncoding.EncodeToString([]byte(message.Body))
	for i := 0; i < len(encodedBody); i += 76 {
		end := i + 76
		if end > len(encodedBody) {
			end = len(encodedBody)
		}
		msgBuilder.WriteString(encodedBody[i:end] + "\r\n")
	}

	allRecipients := append(message.To, message.Cc...)
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	var err error
	if config.Port == 465 {
		err = sendMailWithSSL(addr, auth, from, allRecipients, []byte(msgBuilder.String()))
	} else {
		err = smtp.SendMail(addr, auth, from, allRecipients, []byte(msgBuilder.String()))
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
