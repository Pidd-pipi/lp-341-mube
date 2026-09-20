import request from '../utils/request';
import type { AbnormalMetric, PageData } from '../types';

export function listAbnormalMetrics(params: { examinee_id?: number; page?: number; page_size?: number }): Promise<PageData<AbnormalMetric>> {
  return request.get('/abnormal-metrics', { params });
}

export function updateFollowUp(id: number, data: { status: string; specialist_advice?: string }): Promise<AbnormalMetric> {
  return request.put(`/abnormal-metrics/${id}/follow-up`, data);
}
