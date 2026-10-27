import React, { useEffect, useState } from 'react';
import {
  Alert, Avatar, Button, Card, Col, Descriptions, Divider, Form, Input,
  Row, Space, Spin, Statistic, Tag, Typography, message,
} from 'antd';
import {
  ApiOutlined, BookOutlined, CheckCircleOutlined, KeyOutlined,
  LockOutlined, SafetyCertificateOutlined, UserOutlined,
} from '@ant-design/icons';
import request from '../utils/request';

const { Paragraph, Text, Title } = Typography;
const api = (config) => request(config).then((response) => response.data);
const errorText = (error) => error?.response?.data?.msg || error?.message || '操作失败';

export function Profile() {
  const [profile, setProfile] = useState(null);
  const [saving, setSaving] = useState(false);
  const [passwordSaving, setPasswordSaving] = useState(false);
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();

  useEffect(() => {
    api({ url: '/profile', silent: true }).then((value) => {
      setProfile(value);
      profileForm.setFieldsValue({ name: value.name });
    }).catch((error) => message.error(errorText(error)));
  }, [profileForm]);

  const saveProfile = async (values) => {
    setSaving(true);
    try {
      const value = await api({ url: '/profile', method: 'put', data: values });
      setProfile((current) => ({ ...current, ...value }));
      window.dispatchEvent(new CustomEvent('study-profile-change', { detail: value }));
      message.success('个人资料已更新');
    } finally { setSaving(false); }
  };

  const changePassword = async (values) => {
    setPasswordSaving(true);
    try {
      await api({ url: '/profile/password', method: 'put', data: values });
      passwordForm.resetFields();
      setProfile((current) => ({ ...current, hasPassword: true }));
      message.success('密码修改成功');
    } finally { setPasswordSaving(false); }
  };

  if (!profile) return <div className="center-loading"><Spin size="large" /><Text type="secondary">正在加载个人资料...</Text></div>;

  return <div className="workspace-page">
    <div className="page-heading"><div><Text className="page-eyebrow">ACCOUNT CENTER</Text><Title level={2}>个人主页</Title><Paragraph>管理你的学习身份与账号安全。</Paragraph></div></div>
    <div className="profile-hero">
      <Avatar size={76} icon={<UserOutlined />} className="profile-avatar" />
      <div><Title level={3}>{profile.name}</Title><Text>{profile.account}</Text><div><Tag color="blue">个人学习空间</Tag><Tag color="cyan">数据独立</Tag></div></div>
    </div>
    <Row gutter={[20, 20]}>
      <Col xs={24} lg={14}><Card className="surface-card" title="基本信息" extra={<SafetyCertificateOutlined />}>
        <Form form={profileForm} layout="vertical" onFinish={saveProfile} requiredMark={false}>
          <Form.Item label="姓名" name="name" rules={[{ required: true, message: '请输入姓名' }, { max: 40 }]}><Input prefix={<UserOutlined />} placeholder="请输入姓名" /></Form.Item>
          <Form.Item label="账号"><Input value={profile.account} disabled /></Form.Item>
          <Button type="primary" htmlType="submit" loading={saving}>保存资料</Button>
        </Form>
        <Divider />
        <Descriptions column={1} size="small" items={[
          { key: 'created', label: '空间创建时间', children: new Date(profile.createdAt).toLocaleString() },
          { key: 'mode', label: '账号模式', children: '当前浏览器匿名学习账号' },
        ]} />
      </Card></Col>
      <Col xs={24} lg={10}><Card className="surface-card" title="账号安全" extra={<LockOutlined />}>
        {!profile.hasPassword && <Alert type="info" showIcon message="首次设置密码无需填写当前密码" />}
        <Form form={passwordForm} layout="vertical" onFinish={changePassword} requiredMark={false} className="password-form">
          {profile.hasPassword && <Form.Item label="当前密码" name="currentPassword" rules={[{ required: true }]}><Input.Password prefix={<KeyOutlined />} /></Form.Item>}
          <Form.Item label="新密码" name="newPassword" rules={[{ required: true }, { min: 8, message: '至少 8 位' }, { max: 72 }]}><Input.Password prefix={<LockOutlined />} /></Form.Item>
          <Form.Item label="确认新密码" name="confirmPassword" dependencies={['newPassword']} rules={[{ required: true }, ({ getFieldValue }) => ({ validator(_, value) { return !value || getFieldValue('newPassword') === value ? Promise.resolve() : Promise.reject(new Error('两次密码不一致')); } })]}><Input.Password prefix={<LockOutlined />} /></Form.Item>
          <Button type="primary" htmlType="submit" loading={passwordSaving} block>{profile.hasPassword ? '修改密码' : '设置密码'}</Button>
        </Form>
      </Card></Col>
    </Row>
  </div>;
}

export function AgentSettings() {
  const [settings, setSettings] = useState(null);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState(null);
  const [form] = Form.useForm();

  const load = () => api({ url: '/settings/ai', silent: true }).then((value) => {
    setSettings(value);
    form.setFieldsValue({ baseUrl: value.baseUrl, model: value.model, apiKey: '' });
  }).catch((error) => message.error(errorText(error)));
  useEffect(() => { load(); }, []);

  const save = async (values) => {
    setSaving(true);
    try {
      const value = await api({ url: '/settings/ai', method: 'put', data: values });
      setSettings(value);
      form.setFieldValue('apiKey', '');
      setTestResult(null);
      message.success('配置已写入服务端并立即生效');
    } finally { setSaving(false); }
  };

  const test = async () => {
    setTesting(true);
    setTestResult(null);
    try {
      const value = await api({ url: '/settings/ai/test', method: 'post', timeout: 100000 });
      setTestResult({ ok: true, text: `连接成功，模型返回：${value.reply}` });
    } catch (error) { setTestResult({ ok: false, text: errorText(error) }); }
    finally { setTesting(false); }
  };

  if (!settings) return <div className="center-loading"><Spin size="large" /><Text type="secondary">正在读取智能体配置...</Text></div>;
  return <div className="workspace-page settings-page">
    <div className="page-heading"><div><Text className="page-eyebrow">AGENT CONSOLE</Text><Title level={2}>智能体配置</Title><Paragraph>管理出卷 Agent 使用的模型服务。密钥仅保存在后端，不会返回浏览器。</Paragraph></div><Tag icon={<CheckCircleOutlined />} color={settings.configured ? 'success' : 'warning'}>{settings.configured ? '已配置' : '待配置'}</Tag></div>
    <Row gutter={[20, 20]}>
      <Col xs={24} lg={16}><Card className="surface-card agent-config-card" title={<Space><ApiOutlined />模型连接</Space>}>
        <Form form={form} layout="vertical" onFinish={save} requiredMark={false}>
          <Form.Item label="API Key" name="apiKey" extra={settings.apiKeyMasked ? `当前密钥：${settings.apiKeyMasked}，留空则保持不变` : '密钥会写入 server/.env，并只在服务端使用'}><Input.Password prefix={<KeyOutlined />} placeholder={settings.configured ? '留空保持当前密钥' : '请输入 API Key'} autoComplete="new-password" /></Form.Item>
          <Form.Item label="Base URL" name="baseUrl" rules={[{ required: true, message: '请输入 Base URL' }, { type: 'url', message: '请输入合法 URL' }]}><Input prefix={<ApiOutlined />} placeholder="https://api.deepseek.com" /></Form.Item>
          <Form.Item label="模型名称" name="model" rules={[{ required: true, message: '请输入模型名称' }]}><Input placeholder="deepseek-v4-flash" /></Form.Item>
          <Space wrap><Button type="primary" htmlType="submit" loading={saving}>保存并应用</Button><Button onClick={test} loading={testing}>测试连接</Button></Space>
        </Form>
        {testResult && <Alert className="settings-result" type={testResult.ok ? 'success' : 'error'} showIcon message={testResult.text} />}
      </Card></Col>
      <Col xs={24} lg={8}><Card className="surface-card status-card" title="当前运行状态">
        <Statistic title="模型" value={settings.model || '未设置'} prefix={<BookOutlined />} />
        <Divider />
        <Text type="secondary">服务地址</Text><Paragraph copyable={{ text: settings.baseUrl }} className="setting-value">{settings.baseUrl || '未设置'}</Paragraph>
        <Text type="secondary">密钥状态</Text><Paragraph className="setting-value">{settings.apiKeyMasked || '未设置'}</Paragraph>
      </Card></Col>
    </Row>
  </div>;
}
