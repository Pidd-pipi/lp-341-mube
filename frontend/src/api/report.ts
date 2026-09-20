import request from '../utils/request';
import type { PageData, Report, ReportWithdraw, VersionChain } from '../types';

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

// ---- 撤回重签闭环 ----

export function applyWithdraw(id: number, reason: string): Promise<ReportWithdraw> {
  return request.post(`/reports/${id}/withdraw`, { reason });
}

export function listWithdraws(params: { status?: string; report_id?: number; page?: number; page_size?: number }): Promise<PageData<ReportWithdraw>> {
  return request.get('/reports/withdraws', { params });
}

export function getWithdraw(id: number): Promise<ReportWithdraw> {
  return request.get(`/reports/withdraws/${id}`);
}

export function approveWithdraw(id: number, comment = ''): Promise<{ withdraw: ReportWithdraw; original: Report; resign_report: Report }> {
  return request.post(`/reports/withdraws/${id}/approve`, { approve: true, comment });
}

export function rejectWithdraw(id: number, comment = ''): Promise<ReportWithdraw> {
  return request.post(`/reports/withdraws/${id}/reject`, { approve: false, comment });
}

export function getVersionChain(id: number): Promise<VersionChain> {
  return request.get(`/reports/${id}/versions`);
}
