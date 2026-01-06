import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Select, Button, Checkbox, List, Divider, message } from 'antd';
import useStore from '../store';

const { Option } = Select;
const { TextArea } = Input;

/**
 * QuestionModal组件 - 出题弹窗组件
 * 作用：处理手工出题、编辑题目、AI出题三种模式
 * 特点：一个组件支持多种操作，根据modalType切换显示
 */
const QuestionModal = () => {
  // 从store中获取状态和方法
  const {
    isModalOpen,
    modalType,
    currentQuestion,
    aiParams,
    aiPreviewList,
    selectedAiIds,
    updateAiParams,
    fetchAiQuestions,
    toggleAiSelect,
    saveQuestion,
    confirmAiQuestions,
    closeModal,
  } = useStore();

  // Ant Design表单实例
  const [form] = Form.useForm();
  
  // 当前表单中的题型（用于动态显示/隐藏字段）
  const [currentFormType, setCurrentFormType] = useState('单选题');

  /**
   * 获取弹窗标题
   * 根据modalType返回不同的标题
   */
  const getModalTitle = () => {
    switch (modalType) {
      case 'add':
        return '手工出题';
      case 'edit':
        return '编辑题目';
      case 'ai':
        return 'AI 出题';
      default:
        return '出题';
    }
  };

  /**
   * 监听modalType和currentQuestion变化
   * 用于编辑模式下的数据回填
   */
  useEffect(() => {
    if (modalType === 'edit' && currentQuestion) {
      // 编辑模式：将当前题目的数据回填到表单中
      try {
        const options = currentQuestion.options 
          ? JSON.parse(currentQuestion.options).join('\n')
          : '';
        
        form.setFieldsValue({
          type: currentQuestion.type,
          content: currentQuestion.content,
          difficulty: currentQuestion.difficulty,
          language: currentQuestion.language,
          answer: currentQuestion.answer || '',
          options: options,
        });
        
        setCurrentFormType(currentQuestion.type);
      } catch (error) {
        console.error('数据回填失败：', error);
        message.error('数据加载失败！');
      }
    } else {
      // 新增/AI模式：重置表单
      form.resetFields();
      setCurrentFormType('单选题');
    }
  }, [modalType, currentQuestion, form]);

  /**
   * 表单提交处理
   * 处理手工出题和编辑题目的提交
   */
  const handleSubmit = () => {
    form.validateFields()
      .then((values) => {
        // 处理选项：将文本转换为数组
        const options = values.options 
          ? values.options.split('\n').filter(opt => opt.trim() !== '')
          : [];
        
        // 构建题目对象
        const question = {
          type: values.type,
          content: values.content,
          difficulty: values.difficulty,
          language: values.language,
          answer: values.answer || '',
          options: options.length > 0 ? JSON.stringify(options) : '',
        };
        
        // 调用store中的保存方法
        saveQuestion(question);
      })
      .catch((error) => {
        console.error('表单验证失败：', error);
        message.error('请检查表单填写是否正确！');
      });
  };

  /**
   * 处理AI参数变化
   * @param {string} key - 参数名
   * @param {any} value - 参数值
   */
  const handleAiParamChange = (key, value) => {
    updateAiParams(key, value);
  };

  /**
   * 处理题型变化
   * 用于动态显示/隐藏选项和答案字段
   */
  const handleTypeChange = (value) => {
    setCurrentFormType(value);
    
    // 如果是编程题，清空选项和答案字段
    if (value === '编程题') {
      form.setFieldsValue({ options: '', answer: '' });
    }
  };

  // 弹窗宽度设置
  const modalWidth = modalType === 'ai' ? 1200 : 1800;

  return (
    <Modal
      title={getModalTitle()}
      open={isModalOpen}
      onCancel={closeModal}
      footer={null} // 自定义底部按钮
      width={modalWidth}
      destroyOnClose // 关闭时销毁组件，避免状态残留
      maskClosable={false} // 点击遮罩层不关闭弹窗
    >
      {/* 手工出题/编辑表单 */}
      {modalType === 'add' || modalType === 'edit' ? (
        <Form
          form={form}
          layout="vertical" // 垂直布局
          initialValues={{
            type: '单选题',
            difficulty: '中等',
            language: 'Go',
          }}
          onValuesChange={(changedValues) => {
            // 监听题型变化
            if (changedValues.type) {
              handleTypeChange(changedValues.type);
            }
          }}
        >
          {/* 题型选择 */}
          <Form.Item
            name="type"
            label="题型"
            rules={[{ required: true, message: '请选择题型！' }]}
          >
            <Select>
              <Option value="单选题">单选题</Option>
              <Option value="多选题">多选题</Option>
              <Option value="编程题">编程题</Option>
            </Select>
          </Form.Item>

          {/* 题目内容 */}
          <Form.Item
            name="content"
            label="题目内容"
            rules={[{ required: true, message: '请输入题目内容！' }]}
          >
            <TextArea 
              rows={4} 
              placeholder="请输入题目描述..." 
              maxLength={500}
              showCount
            />
          </Form.Item>

          {/* 难度选择 */}
          <Form.Item
            name="difficulty"
            label="难度"
            rules={[{ required: true, message: '请选择难度！' }]}
          >
            <Select>
              <Option value="简单">简单</Option>
              <Option value="中等">中等</Option>
              <Option value="困难">困难</Option>
            </Select>
          </Form.Item>

          {/* 编程语言选择 */}
          <Form.Item
            name="language"
            label="编程语言"
            rules={[{ required: true, message: '请选择编程语言！' }]}
          >
            <Select>
              <Option value="Go">Go</Option>
              <Option value="JavaScript">JavaScript</Option>
              <Option value="Python">Python</Option>
              <Option value="Java">Java</Option>
              <Option value="C++">C++</Option>
            </Select>
          </Form.Item>

          {/* 选项输入（仅显示/多选题显示） */}
          {(currentFormType === '单选题' || currentFormType === '多选题') && (
            <>
              <Form.Item
                name="options"
                label="选项（每行一个选项，以A. B. C. D.开头）"
                rules={[{ required: true, message: '请输入选项！' }]}
              >
                <TextArea 
                  rows={4} 
                  placeholder="例如：&#10;A. 选项1&#10;B. 选项2&#10;C. 选项3&#10;D. 选项4" 
                />
              </Form.Item>

              <Form.Item
                name="answer"
                label="正确答案"
                rules={[{ required: true, message: '请输入正确答案！' }]}
              >
                <Input 
                  placeholder="例如：A 或 AB（多选题）" 
                  maxLength={10}
                />
              </Form.Item>
            </>
          )}

          {/* 表单按钮 */}
          <Form.Item style={{ textAlign: 'right', marginBottom: 0 }}>
            <Button 
              type="primary" 
              onClick={handleSubmit} 
              style={{ marginRight: 8 }}
            >
              {modalType === 'edit' ? '保存修改' : '添加题目'}
            </Button>
            <Button onClick={closeModal}>取消</Button>
          </Form.Item>
        </Form>
      ) : (
        // AI出题模式
        <div style={{ display: 'flex', flexDirection: 'column', height: '500px' }}>
          {/* AI参数配置区域 */}
          <div style={{ marginBottom: 16 }}>
            <h4>AI出题参数配置</h4>
            <Form layout="inline" style={{ display: 'flex', flexWrap: 'wrap', gap: 16 }}>
              {/* 题型选择 */}
              <Form.Item label="题型" required>
                <Select
                  value={aiParams.type}
                  onChange={(value) => handleAiParamChange('type', value)}
                  style={{ width: 120 }}
                >
                  <Option value="单选题">单选题</Option>
                  <Option value="多选题">多选题</Option>
                  <Option value="编程题">编程题</Option>
                </Select>
              </Form.Item>

              {/* 题目数量 */}
              <Form.Item label="题目数量" required>
                <Input
                  type="number"
                  value={aiParams.count}
                  onChange={(e) => {
                    let count = parseInt(e.target.value) || 1;
                    // 限制在1-10之间
                    count = Math.max(1, Math.min(10, count));
                    handleAiParamChange('count', count);
                  }}
                  style={{ width: 80 }}
                  min={1}
                  max={10}
                />
              </Form.Item>

              {/* 难度选择 */}
              <Form.Item label="难度" required>
                <Select
                  value={aiParams.difficulty}
                  onChange={(value) => handleAiParamChange('difficulty', value)}
                  style={{ width: 100 }}
                >
                  <Option value="简单">简单</Option>
                  <Option value="中等">中等</Option>
                  <Option value="困难">困难</Option>
                </Select>
              </Form.Item>

              {/* 编程语言选择 */}
              <Form.Item label="编程语言" required>
                <Select
                  value={aiParams.language}
                  onChange={(value) => handleAiParamChange('language', value)}
                  style={{ width: 120 }}
                >
                  <Option value="Go">Go</Option>
                  <Option value="JavaScript">JavaScript</Option>
                  <Option value="Python">Python</Option>
                  <Option value="Java">Java</Option>
                  <Option value="C++">C++</Option>
                </Select>
              </Form.Item>

              {/* 生成按钮 */}
              <Button 
                type="primary" 
                onClick={fetchAiQuestions}
                style={{ marginLeft: 'auto' }}
              >
                生成并预览题库
              </Button>
            </Form>
          </div>

          <Divider />

          {/* AI题目预览区域 */}
          <div style={{ flex: 1, overflow: 'auto', marginBottom: 16 }}>
            {aiPreviewList.length === 0 ? (
              <div style={{ 
                textAlign: 'center', 
                padding: 40, 
                color: '#999',
                fontSize: 16 
              }}>
                <p>👆 请在上方配置参数并点击"生成并预览题库"</p>
                <p>AI将根据您的配置生成题目</p>
              </div>
            ) : (
              <>
                <div style={{ 
                  display: 'flex', 
                  justifyContent: 'space-between',
                  marginBottom: 8 
                }}>
                  <span>已生成 {aiPreviewList.length} 道题目</span>
                  <span>已选中 {selectedAiIds.length} 道题目</span>
                </div>
                
                <List
                  dataSource={aiPreviewList}
                  renderItem={(item) => {
                    // 解析选项（如果是JSON字符串）
                    let options = [];
                    try {
                      if (item.options) {
                        options = JSON.parse(item.options);
                      }
                    } catch (e) {
                      options = [];
                    }
                    
                    return (
                      <List.Item
                        key={item.id}
                        actions={[
                          <Checkbox
                            checked={selectedAiIds.includes(item.id)}
                            onChange={() => toggleAiSelect(item.id)}
                          />,
                        ]}
                        style={{
                          border: '1px solid #f0f0f0',
                          marginBottom: 8,
                          borderRadius: 4,
                          padding: '8px 16px',
                          background: selectedAiIds.includes(item.id) ? '#f0f9ff' : '#fff',
                        }}
                      >
                        <List.Item.Meta
                          title={`${item.type}：${item.content}`}
                          description={
                            <div>
                              <p><strong>难度：</strong>{item.difficulty}</p>
                              <p><strong>编程语言：</strong>{item.language}</p>
                              {options.length > 0 && (
                                <div>
                                  <strong>选项：</strong>
                                  {options.map((opt, idx) => (
                                    <div key={idx} style={{ marginLeft: 8 }}>
                                      {opt}
                                    </div>
                                  ))}
                                </div>
                              )}
                              {item.answer && <p><strong>正确答案：</strong>{item.answer}</p>}
                            </div>
                          }
                        />
                      </List.Item>
                    );
                  }}
                />
              </>
            )}
          </div>

          {/* AI题目确认按钮 */}
          
<div style={{ marginBottom: 16 }}>
  <div style={{ 
    display: 'flex', 
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8 
  }}>
    
    
    <h4>AI出题 - 阿里云百炼</h4>
    <span style={{ 
      fontSize: 12, 
      padding: '4px 8px', 
      borderRadius: 4,
      backgroundColor: '#f0f0f0'
    }}>
      {aiPreviewList.length > 0 ? '✅ 已生成' : '🔄 准备生成'}
    </span>
  </div>
  
  <Form layout="inline" style={{ display: 'flex', flexWrap: 'wrap', gap: 16 }}>
    {/* 其他表单项目保持不变 */}
  </Form>
   
   
           {/* 添加选中题目按钮 */}
           
              <Button
                type="primary"
                onClick={async () => {
                  const invalidQuestions = [];
                  let successCount = 0;

                  // 遍历选中的AI题目ID
                  for (const id of selectedAiIds) {
                    const question = aiPreviewList.find(q => q.id === id);
                    if (!question) {
                      invalidQuestions.push(id);
                      continue;
                    }

                    // 基本字段校验
                    if (!question.type || !question.content || !question.answer) {
                      invalidQuestions.push(id);
                      continue;
                    }

                    // 确保 options 是有效的 JSON 字符串
                    let options = '';
                    if (question.options && typeof question.options === 'string') {
                      try {
                        const parsedOptions = JSON.parse(question.options);
                        options = JSON.stringify(parsedOptions);
                      } catch (e) {
                        console.warn('选项解析失败，使用空数组:', e);
                        options = '[]';
                      }
                    } else if (Array.isArray(question.options)) {
                      options = JSON.stringify(question.options);
                    } else {
                      options = '[]';
                    }

                    // 构建完整的题目对象
                    const fullQuestion = {
                      type: question.type,
                      content: question.content,
                      difficulty: question.difficulty,
                      language: question.language,
                      answer: question.answer,
                      options: options,
                    };

                    try {
                      await saveQuestion(fullQuestion);
                      successCount++;
                    } catch (error) {
                      console.error('保存题目失败:', error);
                      invalidQuestions.push(id);
                    }
                  }

                  if (invalidQuestions.length > 0) {
                    message.error(`部分题目保存失败，共 ${invalidQuestions.length} 道无效或无法保存`);
                  }

                  if (successCount > 0) {
                    message.success(`已成功添加 ${successCount} 道题目`);
                  }
                }}
                style={{ marginTop: 16, marginLeft: 'auto' }}
              >
                添加选中题目
              </Button>
            
          

  <div style={{ 
    marginTop: 8, 
    fontSize: 12, 
    color: '#666',
    padding: '8px 12px',
    backgroundColor: '#f9f9f9',
    borderRadius: 4
  }}>
    <div>🔗 正在使用阿里云百炼AI服务</div>
    <div>📝 每次生成消耗API额度，请合理使用</div>
  </div>
</div>
        </div>
      )}
    </Modal>
  );
};

export default QuestionModal;