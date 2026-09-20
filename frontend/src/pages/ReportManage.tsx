import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Descriptions, Drawer, Form, Input, Modal, Space, Table, Tag, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  approveWithdraw,
  draftReport,
  generateReport,
  getVersionChain,
  listReports,
  publishReport,
  rejectWithdraw,
  reportPDFUrl,
  requestWithdraw,
  reviewReport,
} from '../api/report';
import type { Report, ReportVersionChain, ReportWithdrawRequest } from '../types';
import ReportStatusBadge from '../components/common/ReportStatusBadge';
import EmptyState from '../components/common/EmptyState';
import { formatDateTime } from '../utils/dateFormat';
import { usePagination } from '../hooks/usePagination';
import { useAuth } from '../stores/authStore';
import { UserRole } from '../constants/user';
import { ReportStatus, WithdrawStatus, WithdrawStatusLabels } from '../constants/report';

const WITHDRAW_WINDOW_MS = 24 * 60 * 60 * 1000;

// 撤回窗口内（发布后 24 小时）才展示申请入口；最终以后端校验为准。
function withinWithdrawWindow(r: Report): boolean {
  if (!r.published_at) return false;
  return Date.now() - new Date(r.published_at).getTime() <= WITHDRAW_WINDOW_MS;
}

export default function ReportManage() {
  const { user } = useAuth();
  const isAdmin = user?.role === UserRole.ADMIN;
  const [items, setItems] = useState<Report[]>([]);
  const [loading, setLoading] = useState(false);
  const { pagination, setTotal, onPageChange } = usePagination(1, 10);
  const total = pagination.total;
  const [genOpen, setGenOpen] = useState(false);
  const [current, setCurrent] = useState<Report | null>(null);
  const [form] = Form.useForm();
  const [draftRegId, setDraftRegId] = useState('');

  // 撤回申请与版本链
  const [withdrawOpen, setWithdrawOpen] = useState(false);
  const [chainOpen, setChainOpen] = useState(false);
  const [chain, setChain] = useState<ReportVersionChain | null>(null);
  const [chainLoading, setChainLoading] = useState(false);

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
    await draftReport(regId);
    message.success('草稿已创建');
    setDraftRegId('');
    load();
  }

  function openGenerate(r: Report) {
    setCurrent(r);
    form.setFieldsValue({
      conclusion: r.conclusion || '',
      health_advice: r.health_advice || '',
      follow_up_reminder: r.follow_up_reminder || '',
    });
    setGenOpen(true);
  }

  async function onGenerate(values: any) {
    if (!current) return;
    await generateReport(current.id, values);
    message.success(current.status === ReportStatus.RESIGN_PENDING ? '已重新生成报告（含 PDF）' : '报告已生成（含 PDF）');
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

  function openWithdraw(r: Report) {
    setCurrent(r);
    form.setFieldsValue({ reason: '' });
    setWithdrawOpen(true);
  }

  async function onWithdraw(values: any) {
    if (!current) return;
    await requestWithdraw(current.id, values.reason || '');
    message.success('撤回申请已提交，等待管理员审批');
    setWithdrawOpen(false);
    load();
  }

  const openChain = useCallback(async (r: Report) => {
    setCurrent(r);
    setChainOpen(true);
    setChainLoading(true);
    try {
      setChain(await getVersionChain(r.id));
    } finally {
      setChainLoading(false);
    }
  }, []);

  function onApprove(req: ReportWithdrawRequest) {
    Modal.confirm({
      title: `批准撤回申请 #${req.id}？`,
      content: '批准后原版归档保留，并生成一份待重签新版本，审批期间冻结解除。',
      okText: '批准',
      cancelText: '取消',
      onOk: async () => {
        await approveWithdraw(req.id);
        message.success('已批准，待重签版本已生成');
        if (current) await openChain(current);
        load();
      },
    });
  }

  function onReject(req: ReportWithdrawRequest) {
    let comment = '';
    Modal.confirm({
      title: `驳回撤回申请 #${req.id}？`,
      content: <Input.TextArea rows={3} placeholder="驳回说明（可选）" onChange={(e) => { comment = e.target.value; }} />,
      okText: '确认驳回',
      okButtonProps: { danger: true },
      cancelText: '取消',
      onOk: async () => {
        await rejectWithdraw(req.id, comment);
        message.success('已驳回，报告恢复已发布');
        if (current) await openChain(current);
        load();
      },
    });
  }

  const downloadable = (s: string) =>
    ([ReportStatus.GENERATED, ReportStatus.REVIEWED, ReportStatus.PUBLISHED, ReportStatus.WITHDRAWN] as string[]).includes(s);

  const versionColumns: ColumnsType<Report> = [
    { title: '版本', dataIndex: 'version', width: 70 },
    { title: '报告编号', dataIndex: 'report_no' },
    { title: '体检人', render: (_, r) => r.examinee?.name ?? '-' },
    { title: '状态', dataIndex: 'status', render: (v) => <ReportStatusBadge status={v} /> },
    { title: '发布时间', dataIndex: 'published_at', render: (v) => formatDateTime(v) },
    {
      title: '操作',
      render: (_, r) => downloadable(r.status)
        ? <a href={reportPDFUrl(r.id)} target="_blank" rel="noreferrer">下载 PDF</a>
        : <span style={{ color: '#999' }}>无可用 PDF</span>,
    },
  ];

  const requestColumns: ColumnsType<ReportWithdrawRequest> = [
    { title: '申请#', dataIndex: 'id', width: 70 },
    {
      title: '状态', dataIndex: 'status', width: 90,
      render: (v: string) => (
        <Tag color={v === WithdrawStatus.PENDING ? 'orange' : v === WithdrawStatus.APPROVED ? 'green' : 'red'}>
          {WithdrawStatusLabels[v] ?? v}
        </Tag>
      ),
    },
    { title: '撤回原因', dataIndex: 'reason', render: (v) => v || '-' },
    { title: '申请人', dataIndex: 'applicant_id', width: 80 },
    { title: '审批人', dataIndex: 'reviewer_id', width: 80, render: (v) => v || '-' },
    { title: '审批意见', dataIndex: 'review_comment', render: (v) => v || '-' },
    { title: '申请时间', dataIndex: 'created_at', render: (v) => formatDateTime(v) },
    { title: '审批时限', dataIndex: 'expires_at', render: (v) => formatDateTime(v) },
    {
      title: '审批',
      width: 130,
      render: (_, req) => isAdmin && req.status === WithdrawStatus.PENDING
        ? (
          <Space>
            <a onClick={() => onApprove(req)}>批准</a>
            <a style={{ color: '#ff4d4f' }} onClick={() => onReject(req)}>驳回</a>
          </Space>
        )
        : '-',
    },
  ];

  const columns: ColumnsType<Report> = [
    { title: '报告编号', dataIndex: 'report_no' },
    { title: '版本', dataIndex: 'version', width: 70 },
    { title: '体检人', render: (_, r) => r.examinee?.name ?? '-' },
    { title: '状态', dataIndex: 'status', render: (v) => <ReportStatusBadge status={v} /> },
    { title: '生成时间', dataIndex: 'generated_at', render: (v) => formatDateTime(v) },
    {
      title: '操作',
      render: (_, r) => (
        <Space>
          {r.status === ReportStatus.DRAFT && <a onClick={() => openGenerate(r)}>生成报告</a>}
          {r.status === ReportStatus.RESIGN_PENDING && <a onClick={() => openGenerate(r)}>重新生成</a>}
          {r.status === ReportStatus.GENERATED && <a onClick={() => onStatus(r, 'review')}>审核</a>}
          {r.status === ReportStatus.REVIEWED && <a onClick={() => onStatus(r, 'publish')}>发布</a>}
          {r.status === ReportStatus.PUBLISHED && withinWithdrawWindow(r) && <a onClick={() => openWithdraw(r)}>申请撤回</a>}
          {r.status === ReportStatus.WITHDRAWING && <span style={{ color: '#fa8c16' }}>撤回审批中（下载冻结）</span>}
          {downloadable(r.status) && <a href={reportPDFUrl(r.id)} target="_blank" rel="noreferrer">下载 PDF</a>}
          <a onClick={() => openChain(r)}>版本链</a>
        </Space>
      ),
    },
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
        <Form form={form} layout="vertical" onFinish={onGenerate} preserve={false}>
          <Form.Item name="conclusion" label="体检结论"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="health_advice" label="健康建议"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="follow_up_reminder" label="复查提醒"><Input /></Form.Item>
        </Form>
      </Modal>

      <Modal open={withdrawOpen} title={`申请撤回重签：${current?.report_no ?? ''}`} onOk={() => form.submit()} onCancel={() => setWithdrawOpen(false)} destroyOnClose okText="提交申请">
        <Descriptions size="small" column={1} style={{ marginBottom: 12 }}>
          <Descriptions.Item label="撤回窗口">发布后 24 小时内可申请；审批期间冻结下载与新版本生成</Descriptions.Item>
        </Descriptions>
        <Form form={form} layout="vertical" onFinish={onWithdraw} preserve={false}>
          <Form.Item name="reason" label="撤回原因"><Input.TextArea rows={3} maxLength={500} showCount placeholder="请填写撤回原因（可选）" /></Form.Item>
        </Form>
      </Modal>

      <Drawer
        open={chainOpen}
        title={`报告版本链：${current?.report_no ?? ''}`}
        width={1080}
        onClose={() => setChainOpen(false)}
        destroyOnClose
      >
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Card size="small" title="版本链（原版保留，批准后生成待重签新版本）">
            <Table
              rowKey="id"
              size="small"
              loading={chainLoading}
              columns={versionColumns}
              dataSource={chain?.versions ?? []}
              pagination={false}
              locale={{ emptyText: <EmptyState /> }}
            />
          </Card>
          <Card size="small" title={isAdmin ? '撤回申请记录（可审批待处理申请）' : '撤回申请记录'}>
            <Table
              rowKey="id"
              size="small"
              columns={requestColumns}
              dataSource={chain?.requests ?? []}
              pagination={false}
              locale={{ emptyText: <EmptyState /> }}
            />
          </Card>
        </Space>
      </Drawer>
    </Space>
  );
}
