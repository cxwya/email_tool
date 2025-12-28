package email_helper

import (
	"encoding/json"
	"gin_base/app/helper/db_helper"
	"gin_base/app/helper/log_helper"
	"gin_base/app/model"
	"strings"
)

// LogEmailRequest 记录邮件请求和结果到数据库（异步）
func LogEmailRequest(requestIP string, message EmailMessage, config EmailConfig, result EmailResult, requestData interface{}) {
	// 异步记录，不阻塞主流程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log_helper.Error("记录邮件日志panic: %v", r)
			}
		}()

		// 将请求参数转为JSON字符串
		requestDataJSON := ""
		if requestData != nil {
			if jsonData, err := json.Marshal(requestData); err == nil {
				requestDataJSON = string(jsonData)
			}
		}

		// 构建邮件记录
		var isHTML int8 = 0
		if message.IsHTML {
			isHTML = 1
		}
		var success int8 = 0
		if result.Success {
			success = 1
		}

		emailLog := model.EmailLog{
			RequestIP:   requestIP,
			ToEmail:     strings.Join(message.To, ","),
			CcEmail:     strings.Join(message.Cc, ","),
			Subject:     message.Subject,
			Body:        message.Body,
			IsHTML:      isHTML,
			Success:     success,
			Error:       result.Error,
			SmtpHost:    config.Host,
			SmtpPort:    config.Port,
			RequestData: requestDataJSON,
		}

		// 保存到数据库
		if err := db_helper.Db().Create(&emailLog).Error; err != nil {
			log_helper.Error("记录邮件日志失败: %v", err)
		}
	}()
}
