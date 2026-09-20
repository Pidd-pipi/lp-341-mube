import { useCallback, useEffect, useState } from 'react';
import { getDashboard, type DashboardStats } from '../api/stats';

// 运营统计 hook
export function useReportStats() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      setStats(await getDashboard());
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { refresh().catch(() => undefined); }, [refresh]);

  return { stats, loading, refresh };
}
