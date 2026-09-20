import { Layout, Menu, Dropdown, Avatar, Spin } from 'antd';
import { DashboardOutlined, GiftOutlined, FileAddOutlined, EditOutlined, FileTextOutlined, WarningOutlined, TeamOutlined, UserOutlined, LogoutOutlined } from '@ant-design/icons';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../stores/authStore';
import { UserRole } from '../constants/user';

const { Header, Sider, Content } = Layout;

export default function Shell() {
  const { user, logout } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  const items = [
    { key: '/dashboard', icon: <DashboardOutlined />, label: '运营统计' },
    { key: '/packages', icon: <GiftOutlined />, label: '体检套餐' },
    { key: '/registrations', icon: <FileAddOutlined />, label: '体检登记' },
    { key: '/results', icon: <EditOutlined />, label: '结果录入' },
    { key: '/reports', icon: <FileTextOutlined />, label: '报告管理' },
    { key: '/abnormal-metrics', icon: <WarningOutlined />, label: '异常追踪' },
    { key: '/group-orders', icon: <TeamOutlined />, label: '团检管理' },
    { key: '/profile', icon: <UserOutlined />, label: '个人中心' },
  ];

  if (!user) return <Spin style={{ margin: '40vh 50%' }} />;

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider breakpoint="lg" collapsedWidth={64}>
        <div style={{ color: '#fff', padding: 16, fontWeight: 600, fontSize: 15 }}>GbCheckup</div>
        <Menu theme="dark" mode="inline" selectedKeys={[location.pathname]} items={items} onClick={({ key }) => navigate(key)} />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div style={{ fontWeight: 600 }}>体检报告数字化平台</div>
          <div>
            <Dropdown
              menu={{
                items: [{ key: 'logout', icon: <LogoutOutlined />, label: '退出登录' }],
                onClick: ({ key }) => { if (key === 'logout') { logout(); navigate('/login'); } },
              }}
            >
              <Avatar size="small" src={user.avatar || undefined} icon={<UserOutlined />} style={{ cursor: 'pointer' }}>
                {!user.avatar && (user.name || 'U').slice(0, 1)}
              </Avatar>
            </Dropdown>
            <span style={{ marginLeft: 8 }}>{user.name}（{user.role === UserRole.ADMIN ? '管理员' : user.role === UserRole.DOCTOR ? '医生' : user.role === UserRole.FRONT_DESK ? '前台' : '体检人'}）</span>
          </div>
        </Header>
        <Content style={{ margin: 24 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
