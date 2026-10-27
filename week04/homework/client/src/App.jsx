import React, { useEffect, useState } from 'react';
import { Alert, Button, Flex, Spin } from 'antd';
import AppRouter from './router.jsx'; // 导入路由配置
import request from './utils/request';

/**
 * App组件 - 应用根组件
 * 作用：渲染路由组件，作为整个应用的入口
 */
function App() {
  const hasKnownSession = localStorage.getItem('study-session-ready') === 'true';
  const [session, setSession] = useState(hasKnownSession ? 'ready' : 'loading');
  const bootstrap = async () => {
    if (!hasKnownSession) setSession('loading');
    for (let attempt = 0; attempt < 15; attempt += 1) {
      try {
        await request.post('/session/bootstrap', null, { silent: true, timeout: 10000 });
        localStorage.setItem('study-session-ready', 'true');
        setSession('ready');
        return;
      } catch {
        if (hasKnownSession) return;
        await new Promise((resolve) => setTimeout(resolve, 2000));
      }
    }
    setSession('error');
  };
  useEffect(() => {
    bootstrap();
  }, []);
  if (session === 'loading') return <Flex align="center" justify="center" style={{ minHeight: '100vh' }}><Spin size="large" /></Flex>;
  if (session === 'error') return <Flex align="center" justify="center" style={{ minHeight: '100vh' }}><Alert type="error" showIcon message="无法连接后端服务" action={<Button onClick={bootstrap}>重试</Button>} /></Flex>;
  return (
    <div className="App">
      <AppRouter />
    </div>
  );
}

export default App;
