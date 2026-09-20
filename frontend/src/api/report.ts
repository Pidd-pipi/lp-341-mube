import request from '../utils/request';
import type { PageData, Report, ReportVersionChain, ReportWithdrawRequest } from '../types';

export function listReports(params: { status?: string; page?: number; page_size?: number }): Promise<PageData<Report>> {
  return request.get('/reports', { params });
}

export function draftReport(registrationId: number): Promise<Report> {
  return request.post('/reports/draft', null, { params: { registration_id: registrationId } });
}

export function generateReport(id: number, data: { conclusion?: string; health_advice?: string; follow_up_reminder?: string }): Promise<Report> {
  return request.post(`/reports/${id}/generate`, data);
}

export function reviewReport(id: number): Promise<Report> {
  return request.post(`/reports/${id}/review`);
}

export function publishReport(id: number): Promise<Report> {
  return request.post(`/reports/${id}/publish`);
}

export function getReport(id: number): Promise<Report> {
  return request.get(`/reports/${id}`);
}

export function reportPDFUrl(id: number): string {
  return `/api/v1/reports/${id}/pdf`;
}

// 撤回重签闭环
export function requestWithdraw(id: number, reason: string): Promise<ReportWithdrawRequest> {
  return request.post(`/reports/${id}/withdraw`, { reason });
}

export function approveWithdraw(requestId: number, comment?: string): Promise<ReportWithdrawRequest> {
  return request.post(`/reports/withdrawals/${requestId}/approve`, { comment });
}

export function rejectWithdraw(requestId: number, comment?: string): Promise<ReportWithdrawRequest> {
  return request.post(`/reports/withdrawals/${requestId}/reject`, { comment });
}

export function getVersionChain(id: number): Promise<ReportVersionChain> {
  return request.get(`/reports/${id}/versions`);
}
