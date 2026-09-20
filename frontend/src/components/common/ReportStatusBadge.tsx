import { Tag } from 'antd';
import { ReportStatus, ReportStatusLabels } from '../../constants/report';

interface Props {
  status?: string;
}

const colorMap: Record<string, string> = {
  [ReportStatus.DRAFT]: 'default',
  [ReportStatus.GENERATED]: 'blue',
  [ReportStatus.PUBLISHED]: 'green',
  [ReportStatus.REVIEWED]: 'purple',
  [ReportStatus.WITHDRAWING]: 'orange',
  [ReportStatus.WITHDRAWN]: 'red',
};

// 报告状态徽标
export default function ReportStatusBadge({ status }: Props) {
  return <Tag color={colorMap[status ?? ''] ?? 'default'}>{ReportStatusLabels[status ?? ''] ?? status ?? '-'}</Tag>;
}
