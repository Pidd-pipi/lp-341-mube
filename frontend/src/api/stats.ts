import request from '../utils/request';

export interface DashboardStats {
  package_count: number;
  registration_count: number;
  report_count: number;
  abnormal_count: number;
  revenue: number;
  package_sold: { name: string; count: number }[];
  dept_workload: { name: string; count: number }[];
  abnormal_top: { name: string; count: number }[];
  monthly_revenue: { month: string; amount: number }[];
}

export function getDashboard(): Promise<DashboardStats> {
  return request.get('/stats/dashboard');
}
