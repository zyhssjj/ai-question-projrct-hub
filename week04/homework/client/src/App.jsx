import React from 'react';
import AppRouter from './router.jsx'; // 导入路由配置

/**
 * App组件 - 应用根组件
 * 作用：渲染路由组件，作为整个应用的入口
 */
function App() {
  return (
    <div className="App">
      <AppRouter />
    </div>
  );
}

export default App;