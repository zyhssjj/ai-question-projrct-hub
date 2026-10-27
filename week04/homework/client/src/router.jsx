import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout.jsx';
import { Materials, Agent, Exams, ExamDetail, Attempt, Result, WrongQuestions } from './pages/Learning.jsx';
import { AgentSettings, Profile } from './pages/Account.jsx';

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
          {/* 默认路由：资料库 */}
          <Route path="/" element={<Materials />} />
          <Route path="/materials" element={<Materials />} />
          <Route path="/agent" element={<Agent />} />
          <Route path="/exams" element={<Exams />} />
          <Route path="/wrong-questions" element={<WrongQuestions />} />
          <Route path="/exams/:id" element={<ExamDetail />} />
          <Route path="/attempts/:id" element={<Attempt />} />
          <Route path="/attempts/:id/result" element={<Result />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/settings/agent" element={<AgentSettings />} />
          <Route path="*" element={<Materials />} />
        </Routes>
      </Layout>
    </Router>
  );
};

export default AppRouter;
