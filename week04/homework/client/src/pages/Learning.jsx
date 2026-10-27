import React, { useEffect, useMemo, useRef, useState } from 'react';
import {
  Alert, Button, Card, Checkbox, Descriptions, Empty, Flex, Input, List,
  Progress, Radio, Result as ResultView, Segmented, Select, Skeleton, Space, Spin, Statistic,
  Tag, Typography, Upload, message,
} from 'antd';
import {
  CheckCircleOutlined, ClockCircleOutlined, CloudUploadOutlined,
  FileTextOutlined, FlagFilled, FlagOutlined, HistoryOutlined, LeftOutlined,
  RedoOutlined, RightOutlined, SendOutlined,
} from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import request from '../utils/request';

const { Dragger } = Upload;
const { Paragraph, Text, Title } = Typography;
const api = (config) => request(config).then((response) => response.data);
const statusMeta = {
  queued: ['等待处理', 'default'], running: ['处理中', 'processing'], processing: ['解析中', 'processing'],
  ready: ['可使用', 'success'], succeeded: ['已完成', 'success'], generating: ['生成中', 'processing'], failed: ['失败', 'error'],
  in_progress: ['进行中', 'warning'], submitted: ['已交卷', 'success'],
};
const StatusTag = ({ status }) => {
  const [label, color] = statusMeta[status] || [status, 'default'];
  return <Tag color={color}>{label}</Tag>;
};
const errorText = (error) => error?.response?.data?.msg || error?.message || '操作失败';
const materialsCacheKey = 'study-agent-materials';
const readMaterialsCache = () => {
  try { return JSON.parse(localStorage.getItem(materialsCacheKey) || '[]'); }
  catch { return []; }
};
const writeMaterialsCache = (items) => localStorage.setItem(materialsCacheKey, JSON.stringify(items));
const wrongBooksCacheKey = 'study-agent-wrong-books';
const wrongResultCacheKey = (attemptId) => `study-agent-attempt-result-${attemptId}`;
const readCache = (key, fallback = null) => {
  try { return JSON.parse(localStorage.getItem(key) || 'null') ?? fallback; }
  catch { return fallback; }
};
const writeCache = (key, value) => localStorage.setItem(key, JSON.stringify(value));
const examsCacheKey = 'study-agent-exams';
const examCacheKey = (id) => `study-agent-exam-${id}`;
const attemptCacheKey = (id) => `study-agent-attempt-${id}`;
const PageLoading = ({ text = '正在加载...' }) => <div className="page-loading" aria-live="polite">
  <Spin size="large" />
  <Text type="secondary">{text}</Text>
  <Skeleton active paragraph={{ rows: 4 }} title={false} />
</div>;

export function Materials() {
  const initialMaterials = useMemo(readMaterialsCache, []);
  const [materials, setMaterials] = useState(initialMaterials);
  const [jobs, setJobs] = useState({});
  const [uploading, setUploading] = useState(false);
  const [uploadName, setUploadName] = useState('');
  const [loading, setLoading] = useState(!initialMaterials.length);
  const [syncError, setSyncError] = useState('');
  const materialsRef = useRef(initialMaterials);
  const jobsRef = useRef({});
  const load = async (silent = false) => {
    if (!silent) setLoading(true);
    try {
      const items = await api({ url: '/materials', silent, timeout: 60000 });
      materialsRef.current = items;
      writeMaterialsCache(items);
      setMaterials(items);
      setSyncError('');
      return items;
    } finally {
      if (!silent) setLoading(false);
    }
  };

  useEffect(() => {
    load().catch((error) => setSyncError(errorText(error)));
  }, []);
  useEffect(() => { materialsRef.current = materials; }, [materials]);
  useEffect(() => { jobsRef.current = jobs; }, [jobs]);
  useEffect(() => {
    let cancelled = false;
    let timer;
    const poll = async () => {
      const currentJobs = jobsRef.current;
      const activeJobs = Object.entries(currentJobs).filter(([, job]) => !['succeeded', 'failed'].includes(job.status));
      const hasProcessing = materialsRef.current.some((item) => item.status === 'processing');
      if (activeJobs.length || hasProcessing) {
        try {
          const updates = await Promise.all(activeJobs.map(async ([materialId, job]) => {
            try { return [materialId, await api({ url: `/jobs/${job.id}`, silent: true, timeout: 60000 })]; }
            catch { return [materialId, job]; }
          }));
          if (updates.length && !cancelled) {
            const nextJobs = { ...currentJobs, ...Object.fromEntries(updates) };
            jobsRef.current = nextJobs;
            setJobs(nextJobs);
          }
          if (!cancelled) await load(true);
        } catch (error) {
          if (!cancelled) setSyncError(`云端同步较慢，系统会自动重试：${errorText(error)}`);
        }
      }
      if (!cancelled) timer = setTimeout(poll, 5000);
    };
    timer = setTimeout(poll, 5000);
    return () => { cancelled = true; clearTimeout(timer); };
  }, []);

  const upload = async ({ file, onSuccess, onError }) => {
    setUploading(true);
    setUploadName(file.name);
    try {
      const body = new FormData();
      body.append('file', file);
      const result = await api({ url: '/materials', method: 'post', data: body, headers: { 'Content-Type': 'multipart/form-data' }, timeout: 120000 });
      setMaterials((current) => {
        const next = [result.material, ...current.filter((item) => item.id !== result.material.id)];
        materialsRef.current = next;
        writeMaterialsCache(next);
        return next;
      });
      setJobs((current) => {
        const next = { ...current, [result.material.id]: { id: result.jobId, status: 'queued', progress: 0 } };
        jobsRef.current = next;
        return next;
      });
      message.success('上传成功，正在后台解析');
      onSuccess?.(result);
    } catch (error) {
      message.error(errorText(error));
      onError?.(error);
    } finally { setUploading(false); setUploadName(''); }
  };

  return <div className="learning-page">
    <Title level={2}>资料库</Title>
    <Paragraph type="secondary">上传课程讲义、考研或考证资料，解析完成后即可交给 AI 出卷。</Paragraph>
    {syncError && <Alert className="sync-alert" type="warning" showIcon message={syncError} action={<Button size="small" loading={loading} onClick={() => load().catch(() => {})}>立即重试</Button>} />}
    <Dragger customRequest={upload} showUploadList={false} disabled={uploading} accept=".txt,.md,.pdf,.docx" maxCount={1}>
      <Spin spinning={uploading} tip={`正在上传 ${uploadName || '讲义'}...`}>
        <div className="upload-content">
          <p className="ant-upload-drag-icon"><CloudUploadOutlined /></p>
          <p className="ant-upload-text">{uploading ? '正在上传，请勿关闭页面' : '点击或拖入讲义文件'}</p>
          <p className="ant-upload-hint">支持 TXT、Markdown、DOCX、PDF，单文件最大 50MB</p>
        </div>
      </Spin>
    </Dragger>
    <List loading={loading && !materials.length} className="section-list" locale={{ emptyText: <Empty description="尚未上传资料" /> }} dataSource={materials} renderItem={(item) => {
      const job = jobs[item.id];
      const progress = job?.progress ?? (item.status === 'ready' ? 100 : 0);
      return <List.Item>
        <List.Item.Meta avatar={<FileTextOutlined className="file-icon" />} title={<Space><Text strong>{item.name}</Text><StatusTag status={item.status} /></Space>}
          description={<Space direction="vertical" size={4} style={{ width: '100%' }}>
            <Text type="secondary">{new Date(item.createdAt).toLocaleString()}</Text>
            {(item.status === 'processing' || job?.status === 'running' || job?.status === 'queued') && <Progress percent={progress} size="small" status="active" />}
            {item.chapterSummary && <Text>章节：{item.chapterSummary}</Text>}
            {item.errorMessage && <Alert type="error" showIcon message={item.errorMessage} />}
          </Space>} />
      </List.Item>;
    }} />
  </div>;
}

export function Agent() {
  const [materials, setMaterials] = useState(() => readMaterialsCache().filter((item) => item.status === 'ready'));
  const [materialsLoading, setMaterialsLoading] = useState(true);
  const [materialsError, setMaterialsError] = useState('');
  const [selected, setSelected] = useState([]);
  const [conversation, setConversation] = useState(null);
  const [messages, setMessages] = useState([]);
  const [plan, setPlan] = useState(null);
  const [draft, setDraft] = useState('');
  const [sending, setSending] = useState(false);
  const [planning, setPlanning] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const [job, setJob] = useState(null);
  const navigate = useNavigate();
  const bottomRef = useRef(null);

  const loadAgentMaterials = () => {
    setMaterialsLoading(true);
    setMaterialsError('');
    return api({ url: '/materials', silent: true, timeout: 60000 }).then((items) => {
      writeMaterialsCache(items);
      setMaterials(items.filter((item) => item.status === 'ready'));
    }).catch((error) => setMaterialsError(errorText(error))).finally(() => setMaterialsLoading(false));
  };
  useEffect(() => { loadAgentMaterials(); }, []);
  useEffect(() => { bottomRef.current?.scrollIntoView({ behavior: 'smooth' }); }, [messages]);
  useEffect(() => {
    if (!job) return undefined;
    let cancelled = false;
    let timer;
    const poll = async () => {
      try {
        const current = await api({ url: `/jobs/${job.id}`, silent: true, timeout: 60000 });
        if (cancelled) return;
        setJob(current);
        if (current.status === 'succeeded') { navigate(`/exams/${current.examId || job.examId}`); return; }
        if (current.status === 'failed') { message.error(current.errorMessage); return; }
        timer = setTimeout(poll, 3000);
      } catch { if (!cancelled) timer = setTimeout(poll, 5000); }
    };
    timer = setTimeout(poll, 1000);
    return () => { cancelled = true; clearTimeout(timer); };
  }, [job?.id]);

  const createConversation = async () => {
    if (planning) return;
    setPlanning(true);
    try {
      const value = await api({ url: '/agent/conversations', method: 'post', data: { materialIds: selected } });
      setConversation(value);
      const detail = await api({ url: `/agent/conversations/${value.id}` });
      setMessages(detail.messages || []);
    } catch (error) {
      message.error(errorText(error));
    } finally { setPlanning(false); }
  };
  const send = async () => {
    const content = draft.trim();
    if (!content || sending) return;
    setDraft('');
    setMessages((current) => [...current, { id: `local-${Date.now()}`, role: 'user', content }]);
    setSending(true);
    try {
      const value = await api({ url: `/agent/conversations/${conversation.id}/messages`, method: 'post', data: { content } });
      setPlan(value.plan);
      setConversation((current) => ({ ...current, status: value.status }));
      setMessages((current) => [...current, { id: `assistant-${Date.now()}`, role: 'assistant', content: value.reply }]);
    } catch (error) { message.error(errorText(error)); } finally { setSending(false); }
  };
  const confirm = async () => {
    if (confirming) return;
    setConfirming(true);
    try {
      const value = await api({ url: `/agent/conversations/${conversation.id}/confirm`, method: 'post' });
      setJob({ id: value.jobId, examId: value.examId, status: 'queued', progress: 0 });
    } catch (error) {
      message.error(errorText(error));
    } finally { setConfirming(false); }
  };

  if (!conversation) return <div className="learning-page">
    <Title level={2}>AI 出卷</Title>
    <Paragraph type="secondary">选择已经解析完成的资料，然后用自然语言告诉 Agent 你想要怎样的试卷。</Paragraph>
    <Card title="选择资料">
      {materialsError && !materials.length && <Alert type="warning" showIcon message="资料加载较慢" description={materialsError} action={<Button size="small" loading={materialsLoading} onClick={loadAgentMaterials}>重试</Button>} />}
      {materialsLoading && !materials.length ? <PageLoading text="正在加载可用资料..." /> : materials.length ? <Checkbox.Group className="material-options" value={selected} onChange={setSelected}>
        {materials.map((item) => <Checkbox key={item.id} value={item.id}>{item.name}</Checkbox>)}
      </Checkbox.Group> : <Empty description="请先上传并等待资料解析完成" />}
      <Button type="primary" disabled={!selected.length} loading={planning} onClick={createConversation}>{planning ? '正在创建对话' : '开始规划'}</Button>
    </Card>
  </div>;

  return <div className="learning-page agent-layout">
    <div>
      <Title level={2}>AI 出卷 Agent</Title>
      <div className="chat-panel">
        {messages.map((item) => <div key={item.id} className={`chat-row ${item.role}`}><div className="chat-bubble">{item.content}</div></div>)}
        {sending && <div className="chat-row assistant"><div className="chat-bubble"><Spin size="small" /> 正在整理计划...</div></div>}
        <div ref={bottomRef} />
      </div>
      <Flex gap={8} className="chat-input">
        <Input.TextArea value={draft} onChange={(event) => setDraft(event.target.value)} autoSize={{ minRows: 2, maxRows: 5 }}
          placeholder="例如：生成一份 10 题的中等难度测试，6 道单选、4 道多选，重点考第一章，限时 20 分钟"
          onPressEnter={(event) => { if (!event.shiftKey) { event.preventDefault(); send(); } }} />
        <Button type="primary" icon={<SendOutlined />} onClick={send} loading={sending} aria-label="发送" />
      </Flex>
    </div>
    <Card title="出卷计划" className="plan-panel">
      {plan && conversation.status === 'ready' ? <>
        <Descriptions column={1} size="small" items={[
          { key: 'title', label: '名称', children: plan.title },
          { key: 'count', label: '题量', children: `${plan.questionCount} 题（单选 ${plan.singleCount} / 多选 ${plan.multipleCount}）` },
          { key: 'difficulty', label: '难度', children: plan.difficulty },
          { key: 'focus', label: '范围', children: plan.focus },
          { key: 'duration', label: '时限', children: `${plan.durationMinutes} 分钟` },
        ]} />
        {!job ? <Button type="primary" block loading={confirming} onClick={confirm}>{confirming ? '正在创建生成任务' : '确认并生成隐藏试卷'}</Button> : <Space direction="vertical" style={{ width: '100%' }}>
          <Progress percent={job.progress || 0} status={job.status === 'failed' ? 'exception' : 'active'} />
          <Text type="secondary">题目正在后台生成，答案和解析将在交卷后显示。</Text>
        </Space>}
      </> : <Empty description="计划完整后会显示在这里" />}
    </Card>
  </div>;
}

export function Exams() {
  const initialItems = useMemo(() => readCache(examsCacheKey, []), []);
  const [items, setItems] = useState(initialItems);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const navigate = useNavigate();
  const loadExams = () => {
    setLoading(true);
    setLoadError('');
    return api({ url: '/exams', silent: true, timeout: 60000 }).then((value) => {
      writeCache(examsCacheKey, value);
      setItems(value);
    }).catch((error) => setLoadError(errorText(error))).finally(() => setLoading(false));
  };
  useEffect(() => { loadExams(); }, []);
  return <div className="learning-page"><Title level={2}>我的试卷</Title>
    <Paragraph type="secondary">保留每份试卷及全部作答记录，可查看历史解析、续做或重新练习。</Paragraph>
    {loadError && <Alert type="warning" showIcon message="试卷同步失败，当前显示本地记录" action={<Button size="small" loading={loading} onClick={loadExams}>重试</Button>} />}
    {loading && !items.length ? <PageLoading text="正在加载试卷..." /> : <List loading={loading} locale={{ emptyText: <Empty description="还没有生成试卷" /> }} dataSource={items} renderItem={(item) => <List.Item actions={[<Button key="detail" type="primary" ghost onClick={() => navigate(`/exams/${item.id}`)}>试卷详情</Button>]}>
      <List.Item.Meta title={<Space wrap><Text strong>{item.title}</Text><StatusTag status={item.status} />{item.latestAttempt?.status === 'in_progress' && <Tag color="warning">待完成</Tag>}</Space>}
        description={<Space wrap split="·"><Text>{item.questionCount} 题</Text><Text>{item.durationMinutes} 分钟</Text><Text>{item.attemptCount || 0} 次作答</Text>{item.latestAttempt?.status === 'submitted' && <Text>最近得分 {item.latestAttempt.score.toFixed(2)}</Text>}</Space>} />
    </List.Item>} />}
  </div>;
}

export function ExamDetail() {
  const { id } = useParams();
  const [detail, setDetail] = useState(() => readCache(examCacheKey(id)));
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [starting, setStarting] = useState(false);
  const navigate = useNavigate();
  useEffect(() => {
    setLoading(true);
    api({ url: `/exams/${id}`, silent: true, timeout: 60000 }).then((value) => {
      writeCache(examCacheKey(id), value);
      setDetail(value);
      setLoadError('');
    }).catch((error) => setLoadError(errorText(error))).finally(() => setLoading(false));
  }, [id]);
  if (!detail) return loadError ? <Alert type="error" showIcon message="试卷详情加载失败" description={loadError} /> : <PageLoading text="正在加载试卷详情..." />;
  const { exam, attempts = [] } = detail;
  const active = attempts.find((attempt) => attempt.status === 'in_progress' && new Date(attempt.expiresAt) > new Date());
  const begin = async () => {
    setStarting(true);
    try { const attempt = await api({ url: `/exams/${id}/start`, method: 'post' }); navigate(`/attempts/${attempt.id}`); } finally { setStarting(false); }
  };
  return <div className="learning-page">{loadError && <Alert type="warning" showIcon message="云端同步失败，当前显示本地记录" />}{loading && <Progress percent={100} showInfo={false} status="active" size="small" />}<Card>
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <div><Title level={2}>{exam.title}</Title><StatusTag status={exam.status} /></div>
      <Descriptions column={{ xs: 1, sm: 2 }} items={[
        { key: 'count', label: '题目数量', children: `${exam.questionCount} 题` },
        { key: 'duration', label: '考试时长', children: `${exam.durationMinutes} 分钟` },
      ]} />
      {exam.status === 'failed' && <Alert type="error" message="试卷生成失败，请重新发起出卷" showIcon />}
      <Button type="primary" size="large" loading={starting} disabled={exam.status !== 'ready'} onClick={begin}>{active ? '继续答题' : attempts.length ? '重新做题' : '开始做题'}</Button>
    </Space>
  </Card>
    <Title level={3} className="history-title"><HistoryOutlined /> 作答记录</Title>
    <List locale={{ emptyText: <Empty description="暂无作答记录" /> }} dataSource={attempts} renderItem={(attempt, attemptIndex) => <List.Item actions={[
      attempt.status === 'submitted' ? <Button key="result" onClick={() => navigate(`/attempts/${attempt.id}/result`)}>查看全部题目</Button> : <Button key="resume" onClick={() => navigate(`/attempts/${attempt.id}`)}>继续答题</Button>,
    ]}><List.Item.Meta title={`第 ${attempts.length - attemptIndex} 次作答`} description={<Space wrap split="·"><Text>{new Date(attempt.startedAt).toLocaleString()}</Text><StatusTag status={attempt.status} />{attempt.status === 'submitted' && <Text strong>得分 {attempt.score.toFixed(2)}</Text>}</Space>} /></List.Item>} />
  </div>;
}

export function Attempt() {
  const { id } = useParams();
  const [data, setData] = useState(() => readCache(attemptCacheKey(id)));
  const [index, setIndex] = useState(0);
  const [remaining, setRemaining] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const [saving, setSaving] = useState(false);
  const [marking, setMarking] = useState(false);
  const [loadError, setLoadError] = useState('');
  const navigate = useNavigate();
  const submittingRef = useRef(false);
  const pendingSaves = useRef(0);

  const submit = async () => {
    if (submittingRef.current) return;
    submittingRef.current = true;
    setSubmitting(true);
    try { await api({ url: `/attempts/${id}/submit`, method: 'post' }); navigate(`/attempts/${id}/result`, { replace: true }); }
    catch (error) { message.error(errorText(error)); submittingRef.current = false; setSubmitting(false); }
  };
  useEffect(() => {
    api({ url: `/attempts/${id}/questions`, silent: true, timeout: 60000 }).then((value) => {
      setData(value);
      const local = JSON.parse(localStorage.getItem(`attempt-${id}`) || '{}');
      value.questions.forEach((q) => { if ((!q.savedAnswer || !q.savedAnswer.length) && local[q.id]) q.savedAnswer = local[q.id]; });
      writeCache(attemptCacheKey(id), value);
      setLoadError('');
    }).catch((error) => {
      if (error?.response?.status === 409) navigate(`/attempts/${id}/result`, { replace: true });
      else setLoadError(errorText(error));
    });
  }, [id]);
  useEffect(() => {
    if (!data) return undefined;
    const tick = () => {
      const seconds = Math.max(0, Math.ceil((new Date(data.expiresAt).getTime() - Date.now()) / 1000));
      setRemaining(seconds);
      if (seconds === 0) submit();
    };
    tick(); const timer = setInterval(tick, 1000); return () => clearInterval(timer);
  }, [data?.expiresAt]);
  if (!data) return loadError ? <Alert type="error" showIcon message="答题数据加载失败" description={loadError} /> : <PageLoading text="正在加载题目和答题记录..." />;
  const question = data.questions[index];
  const answered = data.questions.filter((item) => item.savedAnswer?.length).length;
  const save = async (answer) => {
    question.savedAnswer = answer;
    setData({ ...data });
    const local = JSON.parse(localStorage.getItem(`attempt-${id}`) || '{}');
    local[question.id] = answer; localStorage.setItem(`attempt-${id}`, JSON.stringify(local));
    writeCache(attemptCacheKey(id), data);
    pendingSaves.current += 1;
    setSaving(true);
    try { await api({ url: `/attempts/${id}/answers/${question.id}`, method: 'put', data: { answer }, silent: true }); }
    catch (error) { message.error(`答案保存失败：${errorText(error)}`); }
    finally {
      pendingSaves.current -= 1;
      if (pendingSaves.current === 0) setSaving(false);
    }
  };
  const toggleMark = async () => {
    if (marking) return;
    const marked = !question.marked;
    question.marked = marked;
    setData({ ...data });
    writeCache(attemptCacheKey(id), data);
    setMarking(true);
    try { await api({ url: `/attempts/${id}/marks/${question.id}`, method: 'put', data: { marked }, silent: true }); }
    catch (error) { question.marked = !marked; setData({ ...data }); message.error(`标记保存失败：${errorText(error)}`); }
    finally { setMarking(false); }
  };
  const time = `${String(Math.floor(remaining / 60)).padStart(2, '0')}:${String(remaining % 60).padStart(2, '0')}`;
  const choices = question.type === 'single'
    ? <Radio.Group value={question.savedAnswer?.[0]} onChange={(event) => save([event.target.value])}><Space direction="vertical">{question.options.map((o) => <Radio key={o.id} value={o.id}>{o.id}. {o.content}</Radio>)}</Space></Radio.Group>
    : <Checkbox.Group value={question.savedAnswer || []} onChange={save}><Space direction="vertical">{question.options.map((o) => <Checkbox key={o.id} value={o.id}>{o.id}. {o.content}</Checkbox>)}</Space></Checkbox.Group>;
  return <div className="attempt-page">
    <header className="attempt-topbar"><Space><ClockCircleOutlined /><Text strong>考试倒计时</Text></Space><span className={remaining < 60 ? 'countdown danger' : 'countdown'}>{time}</span><Space><Text type="secondary">{saving ? '正在保存...' : '已自动保存'}</Text><Text>已答 {answered}/{data.questions.length}</Text><Button danger loading={submitting} onClick={submit}>交卷</Button></Space></header>
    <div className="attempt-layout">
    <aside className="question-nav"><Flex justify="space-between"><Text strong>答题卡</Text><Text type="secondary">共 {data.questions.length} 题</Text></Flex><div className="number-grid">{data.questions.map((item, i) => <Button key={item.id} type={i === index ? 'primary' : 'default'} className={`${item.savedAnswer?.length ? 'answered' : ''} ${item.marked ? 'marked' : ''}`} onClick={() => setIndex(i)}>{item.ordinal}</Button>)}</div><div className="nav-legend"><span><i className="dot answered-dot" />已答</span><span><i className="dot marked-dot" />标记</span></div></aside>
    <main>
      <Card className="question-card"><Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Flex justify="space-between" align="center"><Space><Tag>{question.type === 'single' ? '单选题' : '多选题'}</Tag><Text strong>第 {question.ordinal} 题</Text></Space><Button loading={marking} type={question.marked ? 'primary' : 'text'} danger={question.marked} icon={question.marked ? <FlagFilled /> : <FlagOutlined />} onClick={toggleMark}>{question.marked ? '已标记' : '标记'}</Button></Flex>
        <Title level={4}>{question.content}</Title>{choices}
      </Space></Card>
      <Flex justify="space-between" className="question-actions">
        <Button icon={<LeftOutlined />} disabled={index === 0} onClick={() => setIndex(index - 1)}>上一题</Button>
        <Button type="primary" disabled={index === data.questions.length - 1} onClick={() => setIndex(index + 1)}>下一题<RightOutlined /></Button>
      </Flex>
    </main>
    </div>
  </div>;
}

export function Result() {
  const { id } = useParams();
  const [data, setData] = useState(() => readCache(wrongResultCacheKey(id)));
  const [loadError, setLoadError] = useState('');
  const [filter, setFilter] = useState('全部题目');
  const [retrying, setRetrying] = useState(false);
  const navigate = useNavigate();
  useEffect(() => {
    api({ url: `/attempts/${id}/result`, silent: true, timeout: 60000 }).then((value) => {
      writeCache(wrongResultCacheKey(id), value);
      setData(value);
      setLoadError('');
    }).catch((error) => setLoadError(errorText(error)));
    localStorage.removeItem(`attempt-${id}`);
    localStorage.removeItem(attemptCacheKey(id));
  }, [id]);
  const questions = useMemo(() => !data ? [] : filter === '错题' ? data.questions.filter((q) => !q.isCorrect) : data.questions, [data, filter]);
  if (!data) return loadError ? <Alert type="error" showIcon message="成绩加载失败" description={loadError} /> : <PageLoading text="正在计算并加载成绩..." />;
  const correct = data.questions.filter((q) => q.isCorrect).length;
  const duration = Math.max(0, Math.round((new Date(data.attempt.submittedAt) - new Date(data.attempt.startedAt)) / 60000));
  return <div className="learning-page result-page">{loadError && <Alert type="warning" showIcon message="云端同步失败，当前显示本地成绩" />}
    <ResultView status={data.attempt.score >= 60 ? 'success' : 'warning'} title={`${data.attempt.score.toFixed(2)} 分`}
      subTitle={`答对 ${correct}/${data.questions.length} 题 · 用时约 ${duration} 分钟 · ${data.attempt.submitReason === 'timeout' ? '超时自动交卷' : '主动交卷'}`} />
    <Flex justify="space-between" align="center" wrap="wrap" gap={12}>
      <Segmented value={filter} onChange={setFilter} options={['全部题目', { label: `错题 (${data.questions.length - correct})`, value: '错题' }]} />
      <Button disabled={correct === data.questions.length} loading={retrying} onClick={async () => {
        setRetrying(true); try { const exam = await api({ url: `/attempts/${id}/wrong-questions/retry`, method: 'post' }); navigate(`/exams/${exam.id}`); } finally { setRetrying(false); }
      }}>错题重练</Button>
    </Flex>
    <List className="result-list" dataSource={questions} renderItem={(q) => <Card>
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        <Space><Tag color={q.isCorrect ? 'success' : 'error'} icon={q.isCorrect ? <CheckCircleOutlined /> : null}>{q.isCorrect ? '正确' : '错误'}</Tag><Text strong>第 {q.ordinal} 题 · {q.type === 'single' ? '单选' : '多选'}</Text></Space>
        <Title level={5}>{q.content}</Title>
        <div className="option-review">{q.options.map((o) => <div key={o.id} className={q.correctAnswer.includes(o.id) ? 'correct-option' : q.userAnswer?.includes(o.id) ? 'wrong-option' : ''}>{o.id}. {o.content}</div>)}</div>
        <Descriptions size="small" column={1} items={[
          { key: 'user', label: '你的答案', children: q.userAnswer?.length ? q.userAnswer.join('、') : '未作答' },
          { key: 'answer', label: '正确答案', children: q.correctAnswer.join('、') },
          { key: 'explanation', label: '题目解析', children: q.explanation },
          { key: 'citation', label: '讲义依据', children: <Text mark>{q.citation}</Text> },
        ]} />
      </Space>
    </Card>} />
  </div>;
}

export function WrongQuestions() {
  const initialBooks = useMemo(() => readCache(wrongBooksCacheKey), []);
  const initialBook = initialBooks?.[0] || null;
  const [books, setBooks] = useState(initialBooks);
  const [selectedBook, setSelectedBook] = useState(initialBook);
  const [attemptId, setAttemptId] = useState(initialBook?.latestAttemptId || null);
  const [result, setResult] = useState(() => initialBook ? readCache(wrongResultCacheKey(initialBook.latestAttemptId)) : null);
  const [loading, setLoading] = useState(!initialBooks);
  const [loadError, setLoadError] = useState('');
  const navigate = useNavigate();
  const loadBooks = () => {
    setLoadError('');
    return api({ url: '/wrong-books', silent: true, timeout: 60000 }).then((value) => {
      writeCache(wrongBooksCacheKey, value);
      setBooks(value);
      const current = value.find((book) => book.examId === selectedBook?.examId) || value[0] || null;
      setSelectedBook(current);
      if (current && !current.attempts.some((attempt) => attempt.id === attemptId)) setAttemptId(current.latestAttemptId);
    }).catch((error) => setLoadError(errorText(error))).finally(() => setLoading(false));
  };
  useEffect(() => { loadBooks(); }, []);
  useEffect(() => {
    if (!attemptId) return;
    const cached = readCache(wrongResultCacheKey(attemptId));
    setResult(cached);
    api({ url: `/attempts/${attemptId}/result`, silent: true, timeout: 60000 }).then((value) => {
      writeCache(wrongResultCacheKey(attemptId), value);
      setResult(value);
    }).catch((error) => setLoadError(errorText(error)));
  }, [attemptId]);
  if (!books || (loading && !books.length)) return <PageLoading text="正在加载错题本..." />;
  if (!books.length) return <div className="learning-page"><Title level={2}>错题本</Title><Empty description="目前没有错题" /></div>;
  const wrong = result?.questions?.filter((q) => !q.isCorrect) || [];
  return <div className="learning-page"><Title level={2}>错题本</Title>
    <Paragraph type="secondary">按原始试卷归档错题，可切换查看最近一次练习或任意历史作答。</Paragraph>
    {loadError && <Alert className="sync-alert" type="warning" showIcon message="云端同步较慢，当前优先展示本地记录" action={<Button size="small" loading={loading} onClick={loadBooks}>重试</Button>} />}
    <div className="wrong-book-layout">
      <aside className="wrong-book-list">{books.map((book) => <button type="button" key={book.examId} className={selectedBook?.examId === book.examId ? 'book-item active' : 'book-item'} onClick={() => { setSelectedBook(book); setAttemptId(book.latestAttemptId); }}>
        <strong>{book.title}</strong><span>累计错题 {book.totalWrong} · 练习 {book.attempts.length} 次</span>
      </button>)}</aside>
      <section className="wrong-book-content">
        <Flex justify="space-between" align="center" wrap="wrap" gap={12}>
          <div><Title level={4} style={{ marginBottom: 4 }}>{selectedBook?.title}</Title><Text type="secondary">默认展示最近一次作答</Text></div>
          <Space><Select value={attemptId} style={{ minWidth: 210 }} onChange={setAttemptId} options={selectedBook?.attempts.map((attempt, index) => ({ value: attempt.id, label: `第 ${selectedBook.attempts.length - index} 次 · ${attempt.score.toFixed(2)} 分 · 错 ${attempt.wrongCount} 题` }))} /><Button icon={<RedoOutlined />} onClick={() => navigate(`/exams/${selectedBook.examId}`)}>重新做整卷</Button></Space>
        </Flex>
        {!result ? <PageLoading text="正在加载本次错题..." /> : wrong.length ? <List loading={loading && !result} className="result-list" dataSource={wrong} renderItem={(q) => <Card>
          <Space direction="vertical" size="middle" style={{ width: '100%' }}><Space><Tag color="error">错题</Tag><Text strong>第 {q.ordinal} 题</Text></Space><Title level={5}>{q.content}</Title><Text>你的答案：{q.userAnswer?.length ? q.userAnswer.join('、') : '未作答'}</Text><Text>正确答案：{q.correctAnswer.join('、')}</Text><Text>解析：{q.explanation}</Text><Text mark>讲义依据：{q.citation}</Text></Space>
        </Card>} /> : <ResultView status="success" title="本次已全部答对" subTitle="可以切换上方历史记录查看以前的错题" />}
      </section>
    </div>
  </div>;
}
