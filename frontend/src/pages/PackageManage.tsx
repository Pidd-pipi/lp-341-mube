import { useCallback, useEffect, useState } from 'react';
import { Button, Card, Drawer, Form, Input, InputNumber, Modal, Select, Space, Table, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { addPackageItem, createPackage, deletePackageItem, getPackage, listPackages, updatePackage, updatePackageItem } from '../api/package';
import type { Package, PackageItem } from '../types';
import StatusBadge from '../components/common/StatusBadge';
import EmptyState from '../components/common/EmptyState';
import { PackageStatusLabels, PackageTypeLabels } from '../constants/report';
import { usePagination } from '../hooks/usePagination';

export default function PackageManage() {
  const [items, setItems] = useState<Package[]>([]);
  const [loading, setLoading] = useState(false);
  const { pagination, setTotal, onPageChange } = usePagination(1, 10);
  const total = pagination.total;
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Package | null>(null);
  const [form] = Form.useForm();
  const [detail, setDetail] = useState<{ package: Package; items: PackageItem[] } | null>(null);
  const [itemOpen, setItemOpen] = useState(false);
  const [itemForm] = Form.useForm();

  const load = useCallback(async (page = pagination.page, size = pagination.pageSize) => {
    setLoading(true);
    try {
      const data = await listPackages({ page, page_size: size });
      setItems(data.list);
      setTotal(data.total);
    } finally {
      setLoading(false);
    }
  }, [pagination.page, pagination.pageSize, setTotal]);

  useEffect(() => { load(); }, [load]);

  async function onSave(values: any) {
    if (editing) await updatePackage(editing.id, values);
    else await createPackage(values);
    message.success('保存成功');
    setOpen(false);
    load();
  }

  async function onAddItem(values: any) {
    if (!detail) return;
    await addPackageItem(detail.package.id, values);
    message.success('项目添加成功');
    setItemOpen(false);
    itemForm.resetFields();
    setDetail(await getPackage(detail.package.id));
  }

  async function onDeleteItem(item: PackageItem) {
    if (!detail) return;
    await deletePackageItem(detail.package.id, item.id);
    message.success('项目已删除');
    setDetail(await getPackage(detail.package.id));
  }

  const columns: ColumnsType<Package> = [
    { title: '套餐名称', dataIndex: 'name' },
    { title: '类型', dataIndex: 'package_type', render: (v) => PackageTypeLabels[v] ?? v },
    { title: '价格（元）', dataIndex: 'price' },
    { title: '状态', dataIndex: 'status', render: (v) => <StatusBadge status={v} type="package" /> },
    { title: '操作', render: (_, r) => (
      <Space>
        <a onClick={() => getPackage(r.id).then(setDetail)}>详情</a>
        <a onClick={() => { setEditing(r); form.setFieldsValue(r); setOpen(true); }}>编辑</a>
      </Space>
    ) },
  ];

  return (
    <Space direction="vertical" size={16} style={{ width: '100%' }}>
      <Card size="small">
        <Button type="primary" icon={<PlusOutlined />} onClick={() => { setEditing(null); form.resetFields(); setOpen(true); }}>新增套餐</Button>
      </Card>
      <Card size="small" title="体检套餐列表">
        <Table rowKey="id" columns={columns} dataSource={items} loading={loading} pagination={{ current: pagination.page, pageSize: pagination.pageSize, total, onChange: onPageChange }} locale={{ emptyText: <EmptyState /> }} />
      </Card>

      <Drawer open={open} title={editing ? '编辑套餐' : '新增套餐'} width={420} onClose={() => setOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical" onFinish={onSave} initialValues={{ package_type: 'entry', status: 'active', price: 0 }}>
          <Form.Item name="name" label="套餐名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="package_type" label="套餐类型" rules={[{ required: true }]}><Select options={Object.entries(PackageTypeLabels).map(([value, label]) => ({ value, label }))} /></Form.Item>
          <Form.Item name="price" label="价格（元）"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="status" label="状态"><Select options={Object.entries(PackageStatusLabels).map(([value, label]) => ({ value, label }))} /></Form.Item>
          <Form.Item name="description" label="描述"><Input.TextArea rows={3} /></Form.Item>
          <Button type="primary" htmlType="submit" block>保存</Button>
        </Form>
      </Drawer>

      <Modal open={!!detail} title={detail?.package.name} footer={null} onCancel={() => setDetail(null)} width={700}>
        {detail && (
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <Button type="primary" size="small" onClick={() => setItemOpen(true)}>添加检查项目</Button>
            <Table size="small" rowKey="id" pagination={false} dataSource={detail.items} locale={{ emptyText: <EmptyState description="暂无检查项目" /> }}
              columns={[
                { title: '项目', dataIndex: 'item_name' },
                { title: '分组', dataIndex: 'item_group' },
                { title: '参考值', dataIndex: 'ref_value_range' },
                { title: '科室', dataIndex: 'department' },
                { title: '操作', render: (_, it: PackageItem) => <a style={{ color: '#ff4d4f' }} onClick={() => Modal.confirm({ title: '确认删除？', onOk: () => onDeleteItem(it) })}>删除</a> },
              ]} />
          </Space>
        )}
      </Modal>

      <Modal open={itemOpen} title="添加检查项目" onOk={() => itemForm.submit()} onCancel={() => setItemOpen(false)} destroyOnClose>
        <Form form={itemForm} layout="vertical" onFinish={onAddItem}>
          <Form.Item name="item_name" label="项目名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="item_group" label="分组"><Input /></Form.Item>
          <Form.Item name="ref_value_range" label="参考值范围"><Input placeholder="如 3.5-9.5" /></Form.Item>
          <Form.Item name="department" label="检查科室"><Input /></Form.Item>
          <Form.Item name="sort_order" label="排序"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
