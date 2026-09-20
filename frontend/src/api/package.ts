import request from '../utils/request';
import type { Package, PackageItem, PageData } from '../types';

export function listPackages(params: { status?: string; page?: number; page_size?: number }): Promise<PageData<Package>> {
  return request.get('/packages', { params });
}

export function getPackage(id: number): Promise<{ package: Package; items: PackageItem[] }> {
  return request.get(`/packages/${id}`);
}

export function createPackage(data: Partial<Package>): Promise<Package> {
  return request.post('/packages', data);
}

export function updatePackage(id: number, data: Partial<Package>): Promise<Package> {
  return request.put(`/packages/${id}`, data);
}

export function addPackageItem(packageId: number, data: Partial<PackageItem>): Promise<PackageItem> {
  return request.post(`/packages/${packageId}/items`, data);
}

export function updatePackageItem(packageId: number, itemId: number, data: Partial<PackageItem>): Promise<PackageItem> {
  return request.put(`/packages/${packageId}/items/${itemId}`, data);
}

export function deletePackageItem(packageId: number, itemId: number): Promise<void> {
  return request.delete(`/packages/${packageId}/items/${itemId}`);
}
