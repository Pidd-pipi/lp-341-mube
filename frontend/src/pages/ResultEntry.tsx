import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Input, Modal, Select, Space, Table, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { enterResult, listPendingResults, reviewResult } from '../api/examResult';
import type { ExamResult } from '../types';
import AbnormalTag from '../components/common/AbnormalTag';
import { ResultStatusLabels } from '../constants/report';
import { formatReferenceRange, isAbnormal } from '../utils/formatReferenceRange';
import { usePagination } from '../hooks/usePagination';

export default function ResultEntry() {
  const [items, setItems] = useState<ExamResult[]>([]);
  const [loading, setLoading] = useState(false);
  const { pagination, setTotal, onPageChange } = usePagination(1, 10);
  const total = pagination.total;
  const [editing, setEditing] = useState<ExamResult | null>(null);
  const [value, setValue] = useState('');
  const [text, setText] = useState('');

  const load = useCallback(async (page = pagination.page, size = pagination.pageSize) => {
    setLoading(true);
    try {
      const data = await listPendingResults({ page, page_size: size });
      setItems(data.list);
      setTotal(data.total);
    } finally {
      setLoading(false);
    }
  }, [pagination.page, pagination.pageSize, setTotal]);

  useEffect(() => { load(); }, [load]);

  async function onSave() {
    if (!editing) return;
    const abnormal = isAbnormal(editing.package_item?.ref_value_range ?? '', value);
    await enterResult(editing.id, { result_value: value, result_text: text });
    message.success(abnormal ? '已录入（判定异常）' : '已录入');
    setEditing(null);
    load();
  }

  async function onReview(id: number) {
    await reviewResult(id);
    message.success('已审核');
    load();
  }

  const columns: ColumnsType<ExamResult> = [
    { title: '检查项目', render: (_, r) => r.package_item?.item_name ?? '-' },
    { title: '科室', render: (_, r) => r.package_item?.department ?? '-' },
    { title: '参考值', render: (_, r) => formatReferenceRange(r.package_item?.ref_value_range) },
    { title: '结果值', dataIndex: 'result_value' },
    { title: '异常', dataIndex: 'is_abnormal', render: (v, r) => <AbnormalTag abnormal={v} level={r.is_abnormal ? 'mild' : undefined} /> },
    { title: '状态', dataIndex: 'status', render: (v) => ResultStatusLabels[v] ?? v },
    { title: '操作', render: (_, r) => (
      <Space>
        {r.status !== 'reviewed' && <a onClick={() => { setEditing(r); setValue(r.result_value); setText(r.result_text); }}>录入</a>}
        {r.status === 'entered' && <a onClick={() => onReview(r.id)}>审核</a>}
      </Space>
    ) },
  ];

  return (
    <Card size="small" title="结果录入工作台">
      <Table rowKey="id" columns={columns} dataSource={items} loading={loading} pagination={{ current: pagination.page, pageSize: pagination.pageSize, total, onChange: onPageChange }} />
      <Modal open={!!editing} title={`录入结果：${editing?.package_item?.item_name ?? ''}`} onOk={onSave} onCancel={() => setEditing(null)} destroyOnClose>
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>参考值范围：{formatReferenceRange(editing?.package_item?.ref_value_range)}</div>
          <div>
            <span style={{ marginRight: 8 }}>结果值：</span>
            <Input style={{ width: 200 }} value={value} onChange={(e) => setValue(e.target.value)} placeholder="如 5.2" />
            {isAbnormal(editing?.package_item?.ref_value_range ?? '', value) && <span style={{ color: '#ff4d4f', marginLeft: 8 }}>超出参考值！</span>}
          </div>
          <div>结果描述：<Input.TextArea rows={3} value={text} onChange={(e) => setText(e.target.value)} /></div>
        </Space>
      </Modal>
    </Card>
  );
}
