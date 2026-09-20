import { Tag } from 'antd';
import { RegistrationStatusLabels } from '../../constants/report';

interface Props {
  status?: string;
  type?: 'registration' | 'package' | 'order' | 'delivery';
}

const map: Record<string, Record<string, string>> = {
  registration: { registered: 'blue', in_progress: 'orange', completed: 'green' },
  package: { active: 'green', inactive: 'default' },
  order: { pending: 'orange', confirmed: 'blue', done: 'green' },
  delivery: { pending: 'orange', delivered: 'green' },
};

// 看板/套餐/登记/团检 共用状态徽标
export default function StatusBadge({ status, type = 'registration' }: Props) {
  const m = map[type] ?? {};
  const color = m[status ?? ''] ?? 'default';
  const label = type === 'registration' ? (RegistrationStatusLabels[status ?? ''] ?? status ?? '-') : (status ?? '-');
  return <Tag color={color}>{label}</Tag>;
}
