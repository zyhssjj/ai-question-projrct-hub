import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.jsx';
import 'antd/dist/reset.css'; // 引入Ant Design样式重置
import './styles.css';

if (window.location.hostname === '127.0.0.1') {
  const canonicalURL = new URL(window.location.href);
  canonicalURL.hostname = 'localhost';
  window.location.replace(canonicalURL.toString());
} else {
  ReactDOM.createRoot(document.getElementById('root')).render(<App />);
}
