package service

import "github.com/blueship581/gbcheckup/internal/model"

// VersionChain 报告版本链视图：同一根报告的全部版本 + 撤回重签申请记录。
type VersionChain struct {
	RootReportID   uint                          `json:"root_report_id"`
	Versions       []model.Report                `json:"versions"`
	Requests       []model.ReportWithdrawRequest `json:"requests"`
	PendingRequest *model.ReportWithdrawRequest  `json:"pending_request,omitempty"`
}
