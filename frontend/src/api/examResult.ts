import request from '../utils/request';
import type { ExamResult, PageData } from '../types';

export function listResultsByRegistration(registrationId: number): Promise<ExamResult[]> {
  return request.get('/exam-results', { params: { registration_id: registrationId } });
}

export function listPendingResults(params: { page?: number; page_size?: number }): Promise<PageData<ExamResult>> {
  return request.get('/exam-results/pending', { params });
}

export function enterResult(id: number, data: { result_value?: string; result_text?: string; image_url?: string }): Promise<ExamResult> {
  return request.post(`/exam-results/${id}/enter`, data);
}

export function reviewResult(id: number): Promise<void> {
  return request.post(`/exam-results/${id}/review`);
}
