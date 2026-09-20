import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Form, Input, Modal, Space, Table, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { draftReport, generateReport, listReports, publishReport, reportPDFUrl, reviewReport } from '../api/report';
import type { Report } from '../types';
import ReportStatusBadge from '../components/common/ReportStatusBadge';
import EmptyState from '../components/common/EmptyState';
import { formatDateTime } from '../utils/dateFormat';
import { usePagination } from '../hooks/usePagination';

export default function ReportManage() {
  const [items, setItems] = useState<Report[]>([]);
  const [loading, setLoading] = useState(false);
  const { pagination, setTotal, onPageChange } = usePagination(1, 10);
  const total = pagination.total;
  const [genOpen, setGenOpen] = useState(false);
  const [current, setCurrent] = useState<Report | null>(null);
  const [form] = Form.useForm();
  const [draftRegId, setDraftRegId] = useState('');

  const load = useCallback(async (page = pagination.page, size = pagination.pageSize) => {
    setLoading(true);
    try {
      const data = await listReports({ page, page_size: size });
      setItems(data.list);
      setTotal(data.total);
    } finally {
      setLoading(false);
    }
  }, [pagination.page, pagination.pageSize, setTotal]);

  useEffect(() => { load(); }, [load]);

  async function onDraft() {
    const regId = Number(draftRegId);
    if (!regId) { message.warning('请输入登记 ID'); return; }
    const report = await draftReport(regId);
    message.success('草稿已创建');
    setDraftRegId('');
    load();
    return report;
  }

  async function onGenerate(values: any) {
    if (!current) return;
    await generateReport(current.id, values);
    message.success('报告已生成（含 PDF）');
    setGenOpen(false);
    form.resetFields();
    load();
  }

  async function onStatus(report: Report, action: 'review' | 'publish') {
    if (action === 'review') await reviewReport(report.id);
    else await publishReport(report.id);
    message.success(action === 'review' ? '已审核' : '已发布');
    load();
  }

  const columns: ColumnsType<Report> = [
    { title: '报告编号', dataIndex: 'report_no' },
    { title: '体检人', render: (_, r) => r.examinee?.name ?? '-' },
    { title: '状态', dataIndex: 'status', render: (v) => <ReportStatusBadge status={v} /> },
    { title: '生成时间', dataIndex: 'generated_at', render: (v) => formatDateTime(v) },
    { title: '操作', render: (_, r) => (
      <Space>
        {r.status === 'draft' && <a onClick={() => { setCurrent(r); setGenOpen(true); }}>生成报告</a>}
        {r.status === 'generated' && <a onClick={() => onStatus(r, 'review')}>审核</a>}
        {r.status === 'reviewed' && <a onClick={() => onStatus(r, 'publish')}>发布</a>}
        {(r.status === 'generated' || r.status === 'reviewed' || r.status === 'published') && <a href={reportPDFUrl(r.id)} target="_blank" rel="noreferrer">下载 PDF</a>}
      </Space>
    ) },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card size="small">
        <Space>
          <Input placeholder="登记 ID（创建报告草稿）" value={draftRegId} onChange={(e) => setDraftRegId(e.target.value)} style={{ width: 200 }} />
          <Button type="primary" onClick={onDraft}>创建草稿</Button>
        </Space>
      </Card>
      <Card size="small" title="体检报告列表">
        <Table rowKey="id" columns={columns} dataSource={items} loading={loading} pagination={{ current: pagination.page, pageSize: pagination.pageSize, total, onChange: onPageChange }} locale={{ emptyText: <EmptyState /> }} />
      </Card>

      <Modal open={genOpen} title={`生成报告：${current?.report_no ?? ''}`} onOk={() => form.submit()} onCancel={() => setGenOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={onGenerate}>
          <Form.Item name="conclusion" label="体检结论"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="health_advice" label="健康建议"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="follow_up_reminder" label="复查提醒"><Input /></Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
