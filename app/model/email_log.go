package model

import (
	"gin_base/app/helper/type_helper"
)

// EmailLog 邮件发送记录
type EmailLog struct {
	Id          uint             `gorm:"primarykey;autoIncrement;comment:邮件发送记录表" json:"id"`
	RequestIP   string           `gorm:"type:varchar(50);not null;default:'';comment:请求IP" json:"request_ip"`
	ToEmail     string           `gorm:"type:text;comment:收件人(逗号分隔)" json:"to_email"`
	CcEmail     string           `gorm:"type:text;comment:抄送(逗号分隔)" json:"cc_email"`
	Subject     string           `gorm:"type:varchar(500);not null;default:'';comment:邮件主题" json:"subject"`
	Body        string           `gorm:"type:longtext;comment:邮件正文" json:"body"`
	IsHTML      int8             `gorm:"not null;default:0;comment:是否HTML格式,0-否,1-是" json:"is_html"`
	Success     int8             `gorm:"not null;default:0;comment:是否成功,0-失败,1-成功" json:"success"`
	Error       string           `gorm:"type:text;comment:错误信息" json:"error"`
	SmtpHost    string           `gorm:"type:varchar(200);not null;default:'';comment:SMTP服务器" json:"smtp_host"`
	SmtpPort    int              `gorm:"not null;default:0;comment:SMTP端口" json:"smtp_port"`
	RequestData string           `gorm:"type:longtext;comment:请求参数JSON" json:"request_data"`
	CreatedAt   type_helper.Time `gorm:"comment:创建时间" json:"created_at"`
}
