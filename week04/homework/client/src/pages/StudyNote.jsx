import React, { useState, useEffect } from 'react';
import { Typography, Spin, Alert, Button } from 'antd';

const { Title } = Typography;

const StudyNote = () => {
  const [noteContent, setNoteContent] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchNote = async () => {
      try {
        setLoading(true);
        console.log('🌐 开始加载学习心得...');
        
        // 使用完整的后端URL，确保能获取数据
        const response = await fetch('http://localhost:8080/api/note/get');
        
        if (!response.ok) {
          throw new Error(`HTTP错误: ${response.status}`);
        }
        
        const data = await response.json();
        console.log('📦 收到数据:', data);
        
        if (data.code === 200) {
          setNoteContent(data.data);
          setError(null);
        } else {
          setError(`后端返回错误: ${data.msg}`);
        }
      } catch (err) {
        console.error('❌ 加载学习心得失败：', err);
        setError(`加载失败: ${err.message}`);
      } finally {
        setLoading(false);
      }
    };

    fetchNote();
  }, []);

  // 修复：正确的Spin使用方式
  if (loading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '400px',
        flexDirection: 'column'
      }}>
        <Spin size="large" />
        <div style={{ marginTop: 16, color: '#666' }}>正在加载学习心得...</div>
      </div>
    );
  }

  // 简单的markdown渲染
  const renderMarkdown = (text) => {
    if (!text) return null;
    
    return text.split('\n').map((line, index) => {
      const trimmed = line.trim();
      if (trimmed.startsWith('# ')) {
        return <h1 key={index} style={{ fontSize: '2em', marginTop: '1em' }}>{trimmed.substring(2)}</h1>;
      }
      if (trimmed.startsWith('## ')) {
        return <h2 key={index} style={{ fontSize: '1.5em', marginTop: '0.8em' }}>{trimmed.substring(3)}</h2>;
      }
      if (trimmed.startsWith('### ')) {
        return <h3 key={index} style={{ fontSize: '1.2em', marginTop: '0.6em' }}>{trimmed.substring(4)}</h3>;
      }
      if (trimmed.startsWith('- ') || trimmed.startsWith('* ')) {
        return <li key={index} style={{ marginLeft: '20px', marginBottom: '0.5em' }}>{trimmed.substring(2)}</li>;
      }
      if (trimmed.startsWith('1. ')) {
        return <li key={index} style={{ marginLeft: '20px', marginBottom: '0.5em' }}>{trimmed.substring(3)}</li>;
      }
      if (trimmed === '') {
        return <br key={index} />;
      }
      return <p key={index} style={{ marginBottom: '0.5em' }}>{line}</p>;
    });
  };

  return (
    <div>
      <Title level={2} style={{ marginBottom: 24 }}>学习心得</Title>
      
      {error && (
        <Alert
          message="加载失败"
          description={
            <div>
              <p><strong>错误:</strong> {error}</p>
              <p><strong>请求URL:</strong> http://localhost:8080/api/note/get</p>
              <div style={{ marginTop: 16 }}>
                <Button 
                  type="primary" 
                  onClick={() => window.location.reload()}
                  style={{ marginRight: 8 }}
                >
                  重试加载
                </Button>
                <Button onClick={() => window.open('http://localhost:8080/api/note/get', '_blank')}>
                  直接访问API
                </Button>
              </div>
            </div>
          }
          type="error"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}
      
      {!error && (
        <div style={{
          background: '#fff',
          padding: 24,
          borderRadius: 8,
          border: '1px solid #f0f0f0',
          lineHeight: 1.6,
        }}>
          {renderMarkdown(noteContent)}
        </div>
      )}
      
      {!loading && (
        <div style={{ textAlign: 'right', marginTop: 16 }}>
          <Button onClick={() => window.location.reload()}>
            刷新内容
          </Button>
        </div>
      )}
    </div>
  );
};

export default StudyNote;