import { useState } from 'react';
import { Button, Card, Form, Input, Tabs, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import { login, register } from '../api/user';
import { useAuth } from '../stores/authStore';

export default function Login() {
  const [tab, setTab] = useState('login');
  const [loading, setLoading] = useState(false);
  const { setAuth } = useAuth();
  const navigate = useNavigate();

  async function onFinish(values: { phone: string; password: string; name?: string }) {
    setLoading(true);
    try {
      const result = tab === 'login'
        ? await login(values.phone, values.password)
        : await register(values.phone, values.password, values.name || '新用户');
      setAuth(result.token, result.user);
      message.success('登录成功');
      navigate('/dashboard');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#f0f2f5' }}>
      <Card title="GbCheckup 体检报告数字化平台" style={{ width: 400 }}>
        <Tabs activeKey={tab} onChange={setTab} items={[
          { key: 'login', label: '登录', children: (
            <Form layout="vertical" onFinish={onFinish}>
              <Form.Item name="phone" label="手机号" rules={[{ required: true }]}><Input placeholder="13800000001" /></Form.Item>
              <Form.Item name="password" label="密码" rules={[{ required: true }]}><Input.Password placeholder="admin123" /></Form.Item>
              <Button type="primary" htmlType="submit" block loading={loading}>登录</Button>
            </Form>
          ) },
          { key: 'register', label: '注册', children: (
            <Form layout="vertical" onFinish={onFinish}>
              <Form.Item name="phone" label="手机号" rules={[{ required: true, len: 11 }]}><Input /></Form.Item>
              <Form.Item name="name" label="姓名" rules={[{ required: true }]}><Input /></Form.Item>
              <Form.Item name="password" label="密码" rules={[{ required: true, min: 6 }]}><Input.Password /></Form.Item>
              <Button type="primary" htmlType="submit" block loading={loading}>注册并登录</Button>
            </Form>
          ) },
        ]} />
      </Card>
    </div>
  );
}
