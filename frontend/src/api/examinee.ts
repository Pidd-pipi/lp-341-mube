import request from '../utils/request';
import type { Examinee, PageData } from '../types';

export function listExaminees(params: { keyword?: string; page?: number; page_size?: number }): Promise<PageData<Examinee>> {
  return request.get('/examinees', { params });
}

export function createExaminee(data: Partial<Examinee>): Promise<Examinee> {
  return request.post('/examinees', data);
}

export function batchImportExaminees(csvText: string, enterpriseId?: number): Promise<{ imported: number }> {
  return request.post('/examinees/batch-import', { csv_text: csvText, enterprise_id: enterpriseId });
}
