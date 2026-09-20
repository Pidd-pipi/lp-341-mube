import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Form, Input, InputNumber, Modal, Select, Space, Table, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { createEnterprise, createGroupOrder, deliverReports, listEnterprises, listGroupOrders } from '../api/enterprise';
import { listPackages } from '../api/package';
import type { Enterprise, GroupOrder, Package } from '../types';
import StatusBadge from '../components/common/StatusBadge';
import EmptyState from '../components/common/EmptyState';
import { usePagination } from '../hooks/usePagination';

export default function GroupOrderManage() {
  const [items, setItems] = useState<GroupOrder[]>([]);
  const [enterprises, setEnterprises] = useState<Enterprise[]>([]);
  const [packages, setPackages] = useState<Package[]>([]);
  const [loading, setLoading] = useState(false);
  const { pagination, setTotal, onPageChange } = usePagination(1, 10);
  const total = pagination.total;
  const [orderOpen, setOrderOpen] = useState(false);
  const [entOpen, setEntOpen] = useState(false);
  const [orderForm] = Form.useForm();
  const [entForm] = Form.useForm();

  const load = useCallback(async (page = pagination.page, size = pagination.pageSize) => {
    setLoading(true);
    try {
      const data = await listGroupOrders({ page, page_size: size });
      setItems(data.list);
      setTotal(data.total);
      const ent = await listEnterprises({ page_size: 100 });
      setEnterprises(ent.list);
      const p = await listPackages({ page_size: 100 });
      setPackages(p.list);
    } finally {
      setLoading(false);
    }
  }, [pagination.page, pagination.pageSize, setTotal]);

  useEffect(() => { load(); }, [load]);

  async function onCreateOrder(values: { enterprise_id: number; package_id: number; examinee_count: number }) {
    await createGroupOrder(values);
    message.success('团检订单已创建');
    setOrderOpen(false);
    orderForm.resetFields();
    load();
  }

  async function onCreateEnterprise(values: Partial<Enterprise>) {
    await createEnterprise(values);
    message.success('企业已创建');
    setEntOpen(false);
    entForm.resetFields();
    load();
  }

  async function onDeliver(order: GroupOrder) {
    await deliverReports(order.id);
    message.success('报告已批量交付');
    load();
  }

  const columns: ColumnsType<GroupOrder> = [
    { title: '企业', render: (_, r) => r.enterprise?.name ?? '-' },
    { title: '套餐', render: (_, r) => r.package?.name ?? '-' },
    { title: '人数', dataIndex: 'examinee_count' },
    { title: '订单状态', dataIndex: 'status', render: (v) => <StatusBadge status={v} type="order" /> },
    { title: '交付状态', dataIndex: 'report_delivery_status', render: (v) => <StatusBadge status={v} type="delivery" /> },
    { title: '操作', render: (_, r) => (
      <Space>
        {r.report_delivery_status !== 'delivered' && <a onClick={() => onDeliver(r)}>批量交付报告</a>}
      </Space>
    ) },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card size="small">
        <Space>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setOrderOpen(true)}>创建团检订单</Button>
          <Button onClick={() => setEntOpen(true)}>新增企业</Button>
        </Space>
      </Card>
      <Card size="small" title="团检订单列表">
        <Table rowKey="id" columns={columns} dataSource={items} loading={loading} pagination={{ current: pagination.page, pageSize: pagination.pageSize, total, onChange: onPageChange }} locale={{ emptyText: <EmptyState /> }} />
      </Card>

      <Modal open={orderOpen} title="创建团检订单" onOk={() => orderForm.submit()} onCancel={() => setOrderOpen(false)} destroyOnClose>
        <Form form={orderForm} layout="vertical" onFinish={onCreateOrder}>
          <Form.Item name="enterprise_id" label="企业" rules={[{ required: true }]}>
            <Select options={enterprises.map((e) => ({ value: e.id, label: e.name }))} />
          </Form.Item>
          <Form.Item name="package_id" label="套餐" rules={[{ required: true }]}>
            <Select options={packages.map((p) => ({ value: p.id, label: p.name }))} />
          </Form.Item>
          <Form.Item name="examinee_count" label="人数" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>

      <Modal open={entOpen} title="新增团检企业" onOk={() => entForm.submit()} onCancel={() => setEntOpen(false)} destroyOnClose>
        <Form form={entForm} layout="vertical" onFinish={onCreateEnterprise}>
          <Form.Item name="name" label="企业名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="contact" label="联系人"><Input /></Form.Item>
          <Form.Item name="phone" label="联系电话"><Input /></Form.Item>
          <Form.Item name="address" label="地址"><Input /></Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
