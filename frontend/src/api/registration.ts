import request from '../utils/request';
import type { PageData, Registration } from '../types';

export function listRegistrations(params: { status?: string; page?: number; page_size?: number }): Promise<PageData<Registration>> {
  return request.get('/registrations', { params });
}

export function createRegistration(examineeId: number, packageId: number): Promise<Registration> {
  return request.post('/registrations', { examinee_id: examineeId, package_id: packageId });
}

export function updateRegistrationStatus(id: number, status: string): Promise<void> {
  return request.put(`/registrations/${id}/status`, { status });
}
