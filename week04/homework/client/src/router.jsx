import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout.jsx';
import StudyNote from './pages/StudyNote.jsx';
import QuestionManager from './pages/QuestionManager.jsx';

/**
 * AppRouter组件 - 路由配置组件
 * 作用：配置前端路由，管理页面跳转
 */
const AppRouter = () => {
  return (
    <Router>
      {/* Layout组件作为全局布局，包含侧边栏 */}
      <Layout>
        <Routes>
          {/* 默认路由：学习心得页面 */}
          <Route path="/" element={<StudyNote />} />
          {/* 题库管理页面 */}
          <Route path="/question" element={<QuestionManager />} />
          {/* 404页面重定向到首页 */}
          <Route path="*" element={<StudyNote />} />
        </Routes>
      </Layout>
    </Router>
  );
};

export default AppRouter;