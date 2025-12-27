package email_helper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// EmailLog 邮件日志结构
type EmailLog struct {
	Timestamp   string      `json:"timestamp"`
	RequestIP   string      `json:"request_ip"`
	To          []string    `json:"to"`
	Cc          []string    `json:"cc,omitempty"`
	Subject     string      `json:"subject"`
	Body        string      `json:"body"`
	IsHTML      bool        `json:"is_html"`
	Success     bool        `json:"success"`
	Error       string      `json:"error,omitempty"`
	SmtpHost    string      `json:"smtp_host"`
	SmtpPort    int         `json:"smtp_port"`
	RequestData interface{} `json:"request_data,omitempty"`
}

var logMutex sync.Mutex

// LogEmail 记录邮件发送日志到 runtime 目录，按天分文件
func LogEmail(log EmailLog) error {
	logMutex.Lock()
	defer logMutex.Unlock()

	// 确保 runtime/email_logs 目录存在
	logDir := "./runtime/email_logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 生成按天分的日志文件名
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("email_%s.log", today))

	// 设置时间戳
	if log.Timestamp == "" {
		log.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	}

	// 将日志转为 JSON
	logData, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("序列化日志失败: %v", err)
	}

	// 追加写入文件
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(string(logData) + "\n"); err != nil {
		return fmt.Errorf("写入日志失败: %v", err)
	}

	return nil
}

// LogEmailRequest 记录邮件请求和结果
func LogEmailRequest(requestIP string, message EmailMessage, config EmailConfig, result EmailResult, requestData interface{}) {
	log := EmailLog{
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
		RequestIP:   requestIP,
		To:          message.To,
		Cc:          message.Cc,
		Subject:     message.Subject,
		Body:        message.Body,
		IsHTML:      message.IsHTML,
		Success:     result.Success,
		Error:       result.Error,
		SmtpHost:    config.Host,
		SmtpPort:    config.Port,
		RequestData: requestData,
	}

	if err := LogEmail(log); err != nil {
		fmt.Printf("记录邮件日志失败: %v\n", err)
	}
}
