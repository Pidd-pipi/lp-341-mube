import type { ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../stores/authStore';

export function RequireAuth({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  const location = useLocation();
  if (!token) return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  return <>{children}</>;
}

export function RequireAdmin({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  if (user && user.role !== 'admin') return <Navigate to="/dashboard" replace />;
  return <>{children}</>;
}
