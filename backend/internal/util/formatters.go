package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// formatters.go 同时包含日期格式化、异常等级文本、报告状态文本、结果状态文本、角色文本、参考值范围格式化。

// FormatTime 时间格式化。
func FormatTime(t time.Time) string { return t.Format("2006-01-02 15:04:05") }

// FormatDate 日期格式化。
func FormatDate(t time.Time) string { return t.Format("2006-01-02") }

// FormatMoney 金额保留两位小数。
func FormatMoney(v float64) string { return fmt.Sprintf("%.2f", v) }

// AbnormalLevelText 异常等级中文文案。
func AbnormalLevelText(level string) string {
	switch level {
	case "mild":
		return "轻度异常"
	case "moderate":
		return "中度异常"
	case "severe":
		return "重度异常"
	default:
		return "无异常"
	}
}

// ReportStatusText 报告状态中文文案。
func ReportStatusText(status string) string {
	switch status {
	case "draft":
		return "草稿"
	case "generated":
		return "已生成"
	case "reviewed":
		return "已审核"
	case "published":
		return "已发布"
	default:
		return "未知"
	}
}

// ResultStatusText 结果状态中文文案。
func ResultStatusText(status string) string {
	switch status {
	case "pending":
		return "待录入"
	case "entered":
		return "已录入"
	case "reviewed":
		return "已审核"
	default:
		return "未知"
	}
}

// RoleText 角色中文文案。
func RoleText(role string) string {
	switch role {
	case "admin":
		return "管理员"
	case "doctor":
		return "医生"
	case "front_desk":
		return "前台"
	case "examinee":
		return "体检人"
	default:
		return "未知"
	}
}

// FormatReferenceRange 参考值范围格式化。
func FormatReferenceRange(rng string) string {
	if rng == "" {
		return "-"
	}
	return strings.ReplaceAll(rng, "~", " ~ ")
}

// ParseFloat 安全解析浮点数。
func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}
