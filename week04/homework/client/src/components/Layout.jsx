import React, { useState } from 'react';
import { Layout, Menu, Typography } from 'antd';
import { BookOutlined, DatabaseOutlined } from '@ant-design/icons';
import { Link, useLocation } from 'react-router-dom';

const { Sider, Content } = Layout;
const { Title } = Typography;

/**
 * Layout组件 - 全局布局组件
 * 作用：提供左侧导航栏和右侧内容区域的布局
 * 包含：可折叠的侧边栏、导航菜单、内容区域
 */
const CustomLayout = ({ children }) => {
  // 状态管理：侧边栏是否折叠（默认不折叠）
  const [collapsed, setCollapsed] = useState(false);
  
  // 获取当前路由路径，用于高亮对应的菜单项
  const location = useLocation();

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* 左侧导航栏 */}
      <Sider
        collapsible // 允许折叠
        collapsed={collapsed} // 当前折叠状态
        onCollapse={(value) => setCollapsed(value)} // 折叠/展开时的回调
        style={{
          background: '#fff', // 白色背景
          borderRight: '1px solid #e8e8e8', // 右侧边框
        }}
      >
        {/* 导航栏标题区域 */}
        <div style={{ 
          display: 'flex', 
          alignItems: 'center', 
          padding: '16px',
          borderBottom: '1px solid #f0f0f0',
        }}>
          {!collapsed && (
            <Title level={5} style={{ margin: 0 }}>
              题库管理系统
            </Title>
          )}
        </div>
        
        {/* 导航菜单 */}
        <Menu
          mode="inline" // 垂直菜单模式
          selectedKeys={[location.pathname]} // 高亮当前路由
          items={[
            {
              key: '/', // 菜单项的唯一标识（对应路由路径）
              icon: <BookOutlined />, // 菜单图标
              label: <Link to="/">学习心得</Link>, // 菜单文本（带路由链接）
            },
            {
              key: '/question',
              icon: <DatabaseOutlined />,
              label: <Link to="/question">题库管理</Link>,
            },
          ]}
          style={{ borderRight: 0 }} // 去掉菜单右侧边框
        />
      </Sider>

      {/* 右侧内容区域 */}
      <Layout>
        <Content style={{ 
          padding: '24px', 
          background: '#f5f5f5', // 浅灰色背景
          minHeight: '100vh',
        }}>
          {/* 内容卡片区域 */}
          <div style={{ 
            background: '#fff', // 白色背景
            padding: '24px', 
            borderRadius: '8px', // 圆角边框
            minHeight: 'calc(100vh - 48px)', // 最小高度
          }}>
            {/* 子组件：学习心得或题库管理页面 */}
            {children}
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default CustomLayout;