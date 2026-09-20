import type { ReactNode } from 'react';
import { Result } from 'antd';
import { useAuth } from '../../stores/authStore';

interface Props {
  roles: string[];
  children: ReactNode;
}

export default function RoleGuard({ roles, children }: Props) {
  const { user } = useAuth();
  if (!user || !roles.includes(user.role)) {
    return <Result status="403" title="403" subTitle="当前角色无权访问该功能" />;
  }
  return <>{children}</>;
}
