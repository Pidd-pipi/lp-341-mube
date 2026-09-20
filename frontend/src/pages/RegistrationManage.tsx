import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Form, Input, Modal, Select, Space, Table, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { batchImportExaminees, createExaminee, listExaminees } from '../api/examinee';
import { listPackages } from '../api/package';
import { createRegistration, listRegistrations, updateRegistrationStatus } from '../api/registration';
import type { Examinee, Package, Registration } from '../types';
import StatusBadge from '../components/common/StatusBadge';
import EmptyState from '../components/common/EmptyState';
import { RegistrationStatusLabels } from '../constants/report';
import { formatDateTime } from '../utils/dateFormat';
import { usePagination } from '../hooks/usePagination';

export default function RegistrationManage() {
  const [items, setItems] = useState<Registration[]>([]);
  const [packages, setPackages] = useState<Package[]>([]);
  const [examinees, setExaminees] = useState<Examinee[]>([]);
  const [loading, setLoading] = useState(false);
  const { pagination, setTotal, onPageChange } = usePagination(1, 10);
  const total = pagination.total;
  const [open, setOpen] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [form] = Form.useForm();
  const [importForm] = Form.useForm();

  const load = useCallback(async (page = pagination.page, size = pagination.pageSize) => {
    setLoading(true);
    try {
      const data = await listRegistrations({ page, page_size: size });
      setItems(data.list);
      setTotal(data.total);
      const p = await listPackages({ page_size: 100 });
      setPackages(p.list);
      const e = await listExaminees({ page_size: 100 });
      setExaminees(e.list);
    } finally {
      setLoading(false);
    }
  }, [pagination.page, pagination.pageSize, setTotal]);

  useEffect(() => { load(); }, [load]);

  async function onRegister(values: { examinee_id: number; package_id: number }) {
    await createRegistration(values.examinee_id, values.package_id);
    message.success('登记成功');
    setOpen(false);
    load();
  }

  async function onStatus(reg: Registration, status: string) {
    await updateRegistrationStatus(reg.id, status);
    message.success('状态已更新');
    load();
  }

  async function onNewExaminee(values: { name: string; id_card_no: string; phone?: string; gender?: string; age?: number }) {
    await createExaminee(values);
    message.success('体检人已登记');
    load();
  }

  async function onImport(values: { csv_text: string }) {
    await batchImportExaminees(values.csv_text);
    message.success('批量导入成功');
    setImportOpen(false);
    importForm.resetFields();
  }

  const columns: ColumnsType<Registration> = [
    { title: '导检单号', dataIndex: 'guide_no' },
    { title: '体检人', render: (_, r) => r.examinee?.name ?? '-' },
    { title: '套餐', render: (_, r) => r.package?.name ?? '-' },
    { title: '状态', dataIndex: 'status', render: (v) => <StatusBadge status={v} /> },
    { title: '登记时间', dataIndex: 'registered_at', render: (v) => formatDateTime(v) },
    { title: '操作', render: (_, r) => (
      <Space>
        {r.status !== 'completed' && (
          <Select size="small" style={{ width: 110 }} value={r.status}
            options={Object.entries(RegistrationStatusLabels).map(([value, label]) => ({ value, label }))}
            onChange={(v) => onStatus(r, v)} />
        )}
      </Space>
    ) },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card size="small">
        <Space wrap>
          <Button type="primary" onClick={() => setOpen(true)}>体检登记</Button>
          <Button onClick={() => setImportOpen(true)}>团体批量导入</Button>
        </Space>
      </Card>
      <Card size="small" title="体检登记与导检">
        <Table rowKey="id" columns={columns} dataSource={items} loading={loading} pagination={{ current: pagination.page, pageSize: pagination.pageSize, total, onChange: onPageChange }} locale={{ emptyText: <EmptyState /> }} />
      </Card>

      <Modal open={open} title="体检登记" width={560} onCancel={() => setOpen(false)} footer={null} destroyOnClose>
        <TabsItems packages={packages} examinees={examinees} onRegister={onRegister} onNewExaminee={onNewExaminee} />
      </Modal>

      <Modal open={importOpen} title="团体批量导入（姓名,身份证号,手机号,性别,年龄）" onOk={() => importForm.submit()} onCancel={() => setImportOpen(false)} destroyOnClose>
        <Form form={importForm} layout="vertical" onFinish={onImport}>
          <Form.Item name="csv_text" rules={[{ required: true }]}><Input.TextArea rows={8} placeholder={'张三,110101199001011111,13900000001,male,30\n李四,310101199002022222,13900000002,female,28'} /></Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}

function TabsItems({ packages, examinees, onRegister, onNewExaminee }: {
  packages: Package[]; examinees: Examinee[];
  onRegister: (v: { examinee_id: number; package_id: number }) => void;
  onNewExaminee: (v: { name: string; id_card_no: string; phone?: string; gender?: string; age?: number }) => void;
}) {
  const [tab, setTab] = useState('register');
  const [form] = Form.useForm();
  const [newForm] = Form.useForm();
  return (
    <>
      <TabsComp tab={tab} setTab={setTab} />
      {tab === 'register' ? (
        <Form form={form} layout="vertical" onFinish={onRegister}>
          <Form.Item name="examinee_id" label="体检人" rules={[{ required: true }]}>
            <Select showSearch optionFilterProp="label" options={examinees.map((e) => ({ value: e.id, label: `${e.name}（${e.id_card_no}）` }))} />
          </Form.Item>
          <Form.Item name="package_id" label="体检套餐" rules={[{ required: true }]}>
            <Select options={packages.map((p) => ({ value: p.id, label: `${p.name} ¥${p.price}` }))} />
          </Form.Item>
          <Button type="primary" htmlType="submit" block>确认登记</Button>
        </Form>
      ) : (
        <Form form={newForm} layout="vertical" onFinish={onNewExaminee}>
          <Form.Item name="name" label="姓名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="id_card_no" label="身份证号" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="phone" label="手机号"><Input /></Form.Item>
          <Form.Item name="gender" label="性别"><Select options={[{ value: 'male', label: '男' }, { value: 'female', label: '女' }]} /></Form.Item>
          <Form.Item name="age" label="年龄"><Input type="number" /></Form.Item>
          <Button type="primary" htmlType="submit" block>登记体检人</Button>
        </Form>
      )}
    </>
  );
}

function TabsComp({ tab, setTab }: { tab: string; setTab: (v: string) => void }) {
  return (
    <div style={{ marginBottom: 12 }}>
      <Button type={tab === 'register' ? 'primary' : 'default'} size="small" onClick={() => setTab('register')}>选择已有体检人</Button>
      <Button type={tab === 'new' ? 'primary' : 'default'} size="small" style={{ marginLeft: 8 }} onClick={() => setTab('new')}>新体检人</Button>
    </div>
  );
}
