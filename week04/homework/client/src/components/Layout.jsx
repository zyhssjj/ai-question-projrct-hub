import React, { useEffect, useMemo, useState } from 'react';
import { Avatar, Button, Drawer, Dropdown, Grid, Layout, Menu, Space, Spin, Typography } from 'antd';
import {
  BookOutlined, CloudOutlined, FileDoneOutlined, LogoutOutlined, MenuFoldOutlined,
  MenuUnfoldOutlined, RobotOutlined, SettingOutlined, UserOutlined, WarningOutlined,
} from '@ant-design/icons';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import request from '../utils/request';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

const CustomLayout = ({ children }) => {
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [networkBusy, setNetworkBusy] = useState(false);
  const [profile, setProfile] = useState(() => {
    try { return JSON.parse(localStorage.getItem('study-profile') || 'null'); } catch { return null; }
  });
  const screens = Grid.useBreakpoint();
  const mobile = screens.md === false;
  const isCollapsed = collapsed || mobile;
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    const updateNetwork = (event) => setNetworkBusy(event.detail > 0);
    const updateProfile = (event) => setProfile((current) => {
      const next = { ...current, ...event.detail };
      localStorage.setItem('study-profile', JSON.stringify(next));
      return next;
    });
    window.addEventListener('study-network-change', updateNetwork);
    window.addEventListener('study-profile-change', updateProfile);
    request({ url: '/profile', silent: true }).then((response) => {
      setProfile(response.data);
      localStorage.setItem('study-profile', JSON.stringify(response.data));
    }).catch(() => {});
    return () => {
      window.removeEventListener('study-network-change', updateNetwork);
      window.removeEventListener('study-profile-change', updateProfile);
    };
  }, []);

  const selectedKey = useMemo(() => {
    if (location.pathname.startsWith('/attempts') || location.pathname.startsWith('/exams')) return '/exams';
    return location.pathname === '/' ? '/materials' : location.pathname;
  }, [location.pathname]);
  const menuItems = [
    { key: '/materials', icon: <BookOutlined />, label: <Link to="/materials">资料库</Link> },
    { key: '/agent', icon: <RobotOutlined />, label: <Link to="/agent">AI 出卷</Link> },
    { key: '/exams', icon: <FileDoneOutlined />, label: <Link to="/exams">我的试卷</Link> },
    { key: '/wrong-questions', icon: <WarningOutlined />, label: <Link to="/wrong-questions">错题本</Link> },
    { type: 'divider' },
    { key: '/settings/agent', icon: <SettingOutlined />, label: <Link to="/settings/agent">智能体配置</Link> },
  ];
  const accountMenu = { items: [
    { key: 'profile', icon: <UserOutlined />, label: '个人主页', onClick: () => navigate('/profile') },
    { type: 'divider' },
    { key: 'session', icon: <LogoutOutlined />, label: '当前为本地学习账号', disabled: true },
  ] };
  const navigation = <>
    <Link to="/materials" className={`brand-lockup ${isCollapsed && !mobileOpen ? 'collapsed' : ''}`} onClick={() => setMobileOpen(false)}>
      <span className="brand-mark"><CloudOutlined /></span>
      {(!isCollapsed || mobileOpen) && <span><strong>雨云</strong><small>讲义测验 Agent</small></span>}
    </Link>
    <Menu mode="inline" selectedKeys={[selectedKey]} items={menuItems} className="app-menu" onClick={() => setMobileOpen(false)} />
    {(!isCollapsed || mobileOpen) && <div className="sidebar-note"><RobotOutlined /><div><strong>学习智能体在线</strong><span>基于讲义，专注每一次练习</span></div></div>}
  </>;

  return <Layout className="app-shell">
    {!mobile && <Sider className="app-sidebar" width={236} collapsedWidth={76} collapsed={collapsed} trigger={null}>{navigation}</Sider>}
    <Drawer className="mobile-nav-drawer" placement="left" width={236} open={mobile && mobileOpen} onClose={() => setMobileOpen(false)} closable={false} styles={{ body: { padding: 0 } }}>{navigation}</Drawer>
    <Layout className="app-main">
      <Header className="top-header">
        <Button type="text" className="collapse-button" icon={isCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />} onClick={() => mobile ? setMobileOpen(true) : setCollapsed((value) => !value)} aria-label="切换导航" />
        <div className="header-title"><Text>雨云讲义测验 Agent</Text><span>学有所依，练有所获</span></div>
        <Space className="header-actions">
          {networkBusy && <div className="network-status"><Spin size="small" /><span>正在处理</span></div>}
          <Dropdown menu={accountMenu} placement="bottomRight">
            <Button type="text" className="account-button"><Avatar size={34} icon={<UserOutlined />} /><span className="account-copy"><strong>{profile?.name || '学习者'}</strong><small>{profile?.account || '个人空间'}</small></span></Button>
          </Dropdown>
        </Space>
      </Header>
      <Content className="app-content"><div className="content-canvas">{children}</div></Content>
    </Layout>
  </Layout>;
};

export default CustomLayout;
