import { Button, Card, Form, Input, Upload, message } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { useUserStore } from '../stores/userStore';
import { useAuth } from '../stores/authStore';

export default function Profile() {
  const { user, refresh, save } = useUserStore();
  const { updateUser } = useAuth();
  const [form] = Form.useForm();
  const [saving, setSaving] = useState(false);

  useEffect(() => { refresh().catch(() => undefined); }, []);
  useEffect(() => {
    if (user) form.setFieldsValue({ name: user.name, phone: user.phone, department: user.department, avatar: user.avatar });
  }, [user, form]);

  async function onFinish(values: { name: string; department?: string; avatar?: string }) {
    setSaving(true);
    try {
      const updated = await save({ name: values.name, department: values.department, avatar: values.avatar || '' });
      updateUser(updated);
      message.success('资料已更新');
    } finally {
      setSaving(false);
    }
  }

  return (
    <Card title="个人中心" style={{ maxWidth: 560 }}>
      <Form form={form} layout="vertical" onFinish={onFinish}>
        <Form.Item name="phone" label="手机号"><Input disabled /></Form.Item>
        <Form.Item name="name" label="姓名" rules={[{ required: true }]}><Input /></Form.Item>
        <Form.Item name="department" label="科室/部门"><Input /></Form.Item>
        <Form.Item name="avatar" label="头像">
          <Upload listType="picture" maxCount={1} beforeUpload={() => false} onChange={(info) => {
            const file = info.fileList[0]?.originFileObj;
            if (file) {
              const reader = new FileReader();
              reader.onload = () => form.setFieldValue('avatar', String(reader.result));
              reader.readAsDataURL(file as Blob);
            }
          }}>
            <Button icon={<UploadOutlined />}>上传头像</Button>
          </Upload>
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={saving}>保存资料</Button>
      </Form>
    </Card>
  );
}
