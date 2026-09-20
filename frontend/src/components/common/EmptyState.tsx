import { Empty } from 'antd';

interface Props {
  description?: string;
}

export default function EmptyState({ description = '暂无数据' }: Props) {
  return <Empty description={description} style={{ padding: 24 }} />;
}
