import { Tag } from 'antd';
import { AbnormalLevel, AbnormalLevelLabels } from '../../constants/report';

interface Props {
  level?: string;
  abnormal?: boolean;
}

const colorMap: Record<string, string> = {
  [AbnormalLevel.MILD]: 'orange',
  [AbnormalLevel.MODERATE]: 'volcano',
  [AbnormalLevel.SEVERE]: 'red',
};

// 异常等级标签（看板/结果/异常指标共用）
export default function AbnormalTag({ level, abnormal }: Props) {
  if (abnormal === false) return <Tag color="green">正常</Tag>;
  const key = level ?? AbnormalLevel.MILD;
  return <Tag color={colorMap[key] ?? 'orange'}>{AbnormalLevelLabels[key] ?? key}</Tag>;
}
