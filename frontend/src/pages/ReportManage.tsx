import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Alert, Button, Card, Form, Input, Modal, Space, Table, Tabs, Tag, Timeline, message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  approveWithdraw, applyWithdraw, draftReport, generateReport, getVersionChain,
  listReports, listWithdraws, publishReport, rejectWithdraw, reportPDFUrl, reviewReport,
} from '../api/report';
import type { Report, ReportWithdraw, VersionChain } from '../types';
import ReportStatusBadge from '../components/common/ReportStatusBadge';
import EmptyState from '../components/common/EmptyState';
import RoleGuard from '../components/common/RoleGuard';
import { formatDateTime } from '../utils/dateFormat';
import { usePagination } from '../hooks/usePagination';
import { useAuth } from '../stores/authStore';
import {
  ReportStatus, ReportWithdrawStatus, ReportWithdrawStatusLabels, WITHDRAW_WINDOW_HOURS,
} from '../constants/report';

// 已发布报告是否仍在撤回窗口内（与后端 24h 规则一致，仅用于按钮禁用提示）。
function withinWithdrawWindow(report: Report): boolean {
  if (report.status !== ReportStatus.PUBLISHED) return false;
  const base = report.published_at ? new Date(report.published_at).getTime() : 0;
  return Date.now() - base <= WITHDRAW_WINDOW_HOURS * 3600 * 1000;
}

export default function ReportManage() {
  const { user } = useAuth();
  const isAdmin = user?.role === 'admin';
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
    await draftReport(regId);
    message.success('草稿已创建');
    setDraftRegId('');
    load();
  }

  async function onGenerate(values: { conclusion?: string; health_advice?: string; follow_up_reminder?: string }) {
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

  // 撤回相关弹窗
  const [applyTarget, setApplyTarget] = useState<Report | null>(null);
  const [chainTarget, setChainTarget] = useState<Report | null>(null);

  async function onApply(values: { reason: string }) {
    if (!applyTarget) return;
    await applyWithdraw(applyTarget.id, values.reason);
    message.success('撤回申请已提交，等待管理员审批');
    setApplyTarget(null);
    load();
  }

  const columns: ColumnsType<Report> = [
    { title: '报告编号', render: (_, r) => <Space size={4}>{r.report_no}<Tag>{`V${r.version_no ?? 1}`}</Tag></Space> },
    { title: '体检人', render: (_, r) => r.examinee?.name ?? '-' },
    { title: '状态', dataIndex: 'status', render: (v) => <ReportStatusBadge status={v} /> },
    { title: '生成时间', dataIndex: 'generated_at', render: (v) => formatDateTime(v) },
    {
      title: '操作', render: (_, r) => (
        <Space wrap>
          {r.status === 'draft' && <a onClick={() => { setCurrent(r); setGenOpen(true); }}>生成报告</a>}
          {r.status === 'generated' && <a onClick={() => onStatus(r, 'review')}>审核</a>}
          {r.status === 'reviewed' && <a onClick={() => onStatus(r, 'publish')}>发布</a>}
          {(r.status === 'generated' || r.status === 'reviewed' || r.status === 'published'
            || r.status === 'withdrawn')
            && <a href={reportPDFUrl(r.id)} target="_blank" rel="noreferrer">{r.status === 'withdrawn' ? '原版 PDF' : '下载 PDF'}</a>}
          {r.status === 'published' && (
            withinWithdrawWindow(r)
              ? <a onClick={() => setApplyTarget(r)}>申请撤回重签</a>
              : <span style={{ color: '#999' }}>撤回窗口已过（24h）</span>
          )}
          {r.status === 'withdrawing' && <Tag color="orange">审批中：下载与生成已冻结</Tag>}
          {(r.version_no > 1 || r.status === 'withdrawn' || r.status === 'withdrawing')
            && <a onClick={() => setChainTarget(r)}>版本链</a>}
        </Space>
      ),
    },
  ];

  const listTab = (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card size="small">
        <Space>
          <Input placeholder="登记 ID（创建报告草稿）" value={draftRegId} onChange={(e) => setDraftRegId(e.target.value)} style={{ width: 200 }} />
          <Button type="primary" onClick={onDraft}>创建草稿</Button>
        </Space>
      </Card>
      <Card size="small" title="体检报告列表">
        <Table rowKey="id" columns={columns} dataSource={items} loading={loading}
          pagination={{ current: pagination.page, pageSize: pagination.pageSize, total, onChange: onPageChange }}
          locale={{ emptyText: <EmptyState /> }} />
      </Card>

      <Modal open={genOpen} title={`生成报告：${current?.report_no ?? ''}`} onOk={() => form.submit()} onCancel={() => setGenOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={onGenerate}>
          <Form.Item name="conclusion" label="体检结论"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="health_advice" label="健康建议"><Input.TextArea rows={3} /></Form.Item>
          <Form.Item name="follow_up_reminder" label="复查提醒"><Input /></Form.Item>
        </Form>
      </Modal>

      <Modal open={!!applyTarget} title={`申请撤回重签：${applyTarget?.report_no ?? ''}`}
        onCancel={() => setApplyTarget(null)} footer={null} destroyOnClose>
        <Alert type="info" showIcon style={{ marginBottom: 12 }}
          message={`已发布报告可在发布后 ${WITHDRAW_WINDOW_HOURS} 小时内申请撤回；审批期间将冻结下载与新版本生成。`} />
        <Form layout="vertical" onFinish={onApply}>
          <Form.Item name="reason" label="撤回原因" rules={[{ required: true, min: 2, max: 500, message: '请填写 2-500 字撤回原因' }]}>
            <Input.TextArea rows={4} placeholder="请说明撤回与重签原因" />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">提交申请</Button>
            <Button onClick={() => setApplyTarget(null)}>取消</Button>
          </Space>
        </Form>
      </Modal>

      <VersionChainModal report={chainTarget} onClose={() => setChainTarget(null)} />
    </Space>
  );

  const items2 = [
    { key: 'reports', label: '报告列表', children: listTab },
    ...(isAdmin ? [{ key: 'withdraws', label: '撤回审批', children: <WithdrawReviewTab onReviewed={load} /> }] : []),
  ];

  return <Tabs items={items2} destroyInactiveTabPane />;
}

// VersionChainModal 报告版本链弹窗：版本时间线 + 撤回记录。
function VersionChainModal({ report, onClose }: { report: Report | null; onClose: () => void }) {
  const [chain, setChain] = useState<VersionChain | null>(null);
  useEffect(() => {
    let alive = true;
    if (report) getVersionChain(report.id).then((d) => { if (alive) setChain(d); });
    else setChain(null);
    return () => { alive = false; };
  }, [report]);

  const withdrawMap = useMemo(() => {
    const m = new Map<number, ReportWithdraw>();
    chain?.withdraws.forEach((w) => m.set(w.report_id, w));
    return m;
  }, [chain]);

  return (
    <Modal open={!!report} title={`版本链：${report?.report_no ?? ''}`} footer={null} onCancel={onClose} width={640}>
      {chain && (
        <Timeline
          items={chain.versions.map((v, idx) => ({
            children: (
              <Space direction="vertical" size={2}>
                <Space>
                  <strong>{`V${v.version_no ?? idx + 1} ${v.report_no}`}</strong>
                  <ReportStatusBadge status={v.status} />
                  {idx === 0 && <Tag>原版</Tag>}
                </Space>
                <span style={{ color: '#888' }}>生成：{formatDateTime(v.generated_at)}　发布：{formatDateTime(v.published_at)}</span>
                {withdrawMap.get(v.id) && (
                  <Tag color="orange">{`撤回申请：${ReportWithdrawStatusLabels[withdrawMap.get(v.id)!.status] ?? withdrawMap.get(v.id)!.status}`}</Tag>
                )}
              </Space>
            ),
          }))}
        />
      )}
      {chain && chain.versions.length <= 1 && <EmptyState description="暂无重签版本" />}
    </Modal>
  );
}

// WithdrawReviewTab 管理员撤回审批台。
function WithdrawReviewTab({ onReviewed }: { onReviewed: () => void }) {
  const [items, setItems] = useState<ReportWithdraw[]>([]);
  const [loading, setLoading] = useState(false);
  const { pagination, setTotal, onPageChange } = usePagination(1, 10);
  const total = pagination.total;
  const [reviewTarget, setReviewTarget] = useState<ReportWithdraw | null>(null);
  const [action, setAction] = useState<'approve' | 'reject'>('approve');

  const load = useCallback(async (page = pagination.page, size = pagination.pageSize) => {
    setLoading(true);
    try {
      const data = await listWithdraws({ page, page_size: size });
      setItems(data.list);
      setTotal(data.total);
    } finally {
      setLoading(false);
    }
  }, [pagination.page, pagination.pageSize, setTotal]);

  useEffect(() => { load(); }, [load]);

  async function onReview(values: { comment?: string }) {
    if (!reviewTarget) return;
    if (action === 'approve') {
      await approveWithdraw(reviewTarget.id, values.comment ?? '');
      message.success('已批准：原版已归档保留，待重签版本已生成（草稿）');
    } else {
      await rejectWithdraw(reviewTarget.id, values.comment ?? '');
      message.success('已驳回：报告恢复为已发布');
    }
    setReviewTarget(null);
    load();
    onReviewed();
  }

  const columns: ColumnsType<ReportWithdraw> = [
    { title: '申请 ID', dataIndex: 'id', width: 80 },
    { title: '报告编号', render: (_, w) => w.report?.report_no ?? `#${w.report_id}` },
    { title: '体检人', render: (_, w) => w.report?.examinee?.name ?? '-' },
    { title: '撤回原因', dataIndex: 'reason', ellipsis: true },
    {
      title: '状态', dataIndex: 'status', width: 100,
      render: (v) => <Tag color={v === ReportWithdrawStatus.PENDING ? 'orange' : v === ReportWithdrawStatus.APPROVED ? 'green' : 'default'}>
        {ReportWithdrawStatusLabels[v] ?? v}
      </Tag>,
    },
    { title: '截止时间', dataIndex: 'expires_at', render: (v) => formatDateTime(v) },
    {
      title: '操作', width: 140, render: (_, w) => w.status === ReportWithdrawStatus.PENDING ? (
        <Space>
          <a onClick={() => { setAction('approve'); setReviewTarget(w); }}>批准</a>
          <a style={{ color: '#cf1322' }} onClick={() => { setAction('reject'); setReviewTarget(w); }}>驳回</a>
        </Space>
      ) : <span style={{ color: '#999' }}>已处理</span>,
    },
  ];

  return (
    <RoleGuard roles={['admin']}>
      <Card size="small" title="撤回重签审批">
        <Alert type="info" showIcon style={{ marginBottom: 12 }}
          message="同一报告仅允许一份待审批申请；批准将保留原版并生成待重签草稿版本，驳回则恢复已发布。重复或并发审批返回冲突且状态不变。" />
        <Table rowKey="id" columns={columns} dataSource={items} loading={loading}
          pagination={{ current: pagination.page, pageSize: pagination.pageSize, total, onChange: onPageChange }}
          locale={{ emptyText: <EmptyState /> }} />
      </Card>
      <Modal open={!!reviewTarget}
        title={`${action === 'approve' ? '批准' : '驳回'}撤回申请 #${reviewTarget?.id ?? ''}`}
        onCancel={() => setReviewTarget(null)} footer={null} destroyOnClose>
        {reviewTarget && (
          <Space direction="vertical" size={8} style={{ width: '100%' }}>
            <div><strong>报告：</strong>{reviewTarget.report?.report_no}</div>
            <div><strong>撤回原因：</strong>{reviewTarget.reason}</div>
            <Form layout="vertical" onFinish={onReview}>
              <Form.Item name="comment" label="审批意见">
                <Input.TextArea rows={3} maxLength={500} placeholder={action === 'reject' ? '可填写驳回原因（选填）' : '可填写批准意见（选填）'} />
              </Form.Item>
              <Button type={action === 'reject' ? 'default' : 'primary'} danger={action === 'reject'} htmlType="submit">
                {action === 'approve' ? '确认批准' : '确认驳回'}
              </Button>
            </Form>
          </Space>
        )}
      </Modal>
    </RoleGuard>
  );
}
