import React, { useEffect } from 'react';
import { 
  Table, Button, Input, Select, Space, 
  Popconfirm, Typography, Row, Col, 
  Checkbox, Tag, message, Dropdown, Menu 
} from 'antd';
import { 
  SearchOutlined, EditOutlined, DeleteOutlined, 
  PlusOutlined, DownloadOutlined 
} from '@ant-design/icons';
import useStore from '../store';
import QuestionModal from '../components/QuestionModal';

const { Title } = Typography;
const { Option } = Select;
const { Search } = Input;

/**
 * QuestionManager组件 - 题库管理页面
 * 作用：显示题目列表，支持增删改查、搜索、筛选、分页等功能
 * 特点：核心业务页面，功能最复杂
 */
const QuestionManager = () => {
  // 从store中获取状态和方法
  const {
    questionList,
    total,
    pageNum,
    pageSize,
    selectedType,
    searchKey,
    selectedIds,
    fetchQuestionList,
    updateFilter,
    updatePage,
    openModal,
    deleteQuestion,
    toggleQuestionSelect,
    selectAllQuestions,
    batchDeleteQuestion,
  } = useStore();

  /**
   * 组件挂载时加载题目列表
   * 依赖项为空数组，表示只在组件挂载时执行一次
   */
  useEffect(() => {
    fetchQuestionList();
  }, []);

  /**
   * 处理搜索
   * @param {string} value - 搜索关键词
   */
  const handleSearch = (value) => {
    updateFilter(selectedType, value);
  };

  /**
   * 处理题型筛选变化
   * @param {string} type - 题型
   */
  const handleTypeChange = (type) => {
    updateFilter(type, searchKey);
  };

  /**
   * 处理表格分页变化
   * @param {Object} pagination - 分页参数
   * @param {number} pagination.current - 当前页码
   * @param {number} pagination.pageSize - 每页大小
   */
  const handleTableChange = (pagination) => {
    updatePage(pagination.current, pagination.pageSize);
  };

  /**
   * 处理批量删除确认
   */
  const handleBatchDelete = () => {
    if (selectedIds.length === 0) {
      message.warning('请选择要删除的题目！');
      return;
    }
    
    batchDeleteQuestion();
  };

  // 出题按钮的下拉菜单
  const addMenu = (
    <Menu>
      <Menu.Item key="ai" onClick={() => openModal('ai')}>
        AI 出题
      </Menu.Item>
      <Menu.Item key="manual" onClick={() => openModal('add')}>
        手工出题
      </Menu.Item>
    </Menu>
  );

  /**
   * 表格列配置
   * 定义了表格每一列的显示方式和行为
   */
  const columns = [
    {
        title: () => {
            // 判断当前页是否全选
            const currentPageIds = questionList.map(item => item.id);
            const isAllSelected = currentPageIds.length > 0 && 
                                 currentPageIds.every(id => selectedIds.includes(id));
            
            return (
              <Checkbox
                checked={isAllSelected}
                indeterminate={
                  selectedIds.length > 0 && 
                  selectedIds.length < currentPageIds.length
                }
                onChange={(e) => {
                  if (e.target.checked) {
                    // 全选当前页
                    currentPageIds.forEach(id => {
                      if (!selectedIds.includes(id)) {
                        toggleQuestionSelect(id);
                      }
                    });
                  } else {
                    // 取消全选当前页
                    currentPageIds.forEach(id => {
                      if (selectedIds.includes(id)) {
                        toggleQuestionSelect(id);
                      }
                    });
                  }
                }}
              />
            );
          },
      dataIndex: 'id',
      key: 'select',
      width: 60,
      align: 'center',
      render: (id) => (
        <Checkbox
          checked={selectedIds.includes(id)}
          onChange={() => toggleQuestionSelect(id)}
        />
      ),
    },
    {
      title: '题目',
      dataIndex: 'content',
      key: 'content',
      width: '40%',
      render: (text, record) => (
        <div>
          <div style={{ 
            fontWeight: 'bold', 
            marginBottom: 4,
            color: '#1890ff',
            cursor: 'pointer',
          }}
          onClick={() => openModal('edit', record)}
          >
            {text.length > 100 ? text.substring(0, 100) + '...' : text}
          </div>
          <div style={{ fontSize: 12, color: '#999' }}>
            ID: {record.id} | 创建时间: {new Date(record.createdAt).toLocaleDateString()}
          </div>
        </div>
      ),
    },
    {
      title: '题型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      filters: [
        { text: '单选题', value: '单选题' },
        { text: '多选题', value: '多选题' },
        { text: '编程题', value: '编程题' },
      ],
      onFilter: (value, record) => record.type === value,
      render: (type) => {
        let color = '';
        switch (type) {
          case '单选题':
            color = 'green';
            break;
          case '多选题':
            color = 'blue';
            break;
          case '编程题':
            color = 'orange';
            break;
          default:
            color = 'gray';
        }
        return <Tag color={color}>{type}</Tag>;
      },
    },
    {
      title: '难度',
      dataIndex: 'difficulty',
      key: 'difficulty',
      width: 80,
      render: (difficulty) => {
        let color = '';
        switch (difficulty) {
          case '简单':
            color = 'green';
            break;
          case '中等':
            color = 'orange';
            break;
          case '困难':
            color = 'red';
            break;
          default:
            color = 'gray';
        }
        return <Tag color={color}>{difficulty}</Tag>;
      },
    },
    {
      title: '编程语言',
      dataIndex: 'language',
      key: 'language',
      width: 100,
      render: (language) => <Tag>{language}</Tag>,
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space size="small">
          {/* 编辑按钮 */}
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => openModal('edit', record)}
            style={{ color: '#1890ff' }}
          >
            编辑
          </Button>
          
          {/* 删除按钮（带确认） */}
          <Popconfirm
            title="确定要删除这道题目吗？"
            description="删除后将无法恢复，请谨慎操作！"
            onConfirm={() => deleteQuestion(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      {/* 页面标题和操作按钮区域 */}
      <div style={{ 
        marginBottom: 24, 
        display: 'flex', 
        justifyContent: 'space-between', 
        alignItems: 'center' 
      }}>
        <Title level={2} style={{ margin: 0 }}>
          题库管理
          <span style={{ 
            fontSize: 14, 
            color: '#999', 
            marginLeft: 8, 
            fontWeight: 'normal' 
          }}>
            共 {total} 道题目
          </span>
        </Title>
        
        <Space>
          {/* 批量删除按钮 */}
          {selectedIds.length > 0 && (
            <Button
              type="primary"
              danger
              onClick={handleBatchDelete}
            >
              批量删除 ({selectedIds.length})
            </Button>
          )}
          
          {/* 导出按钮（示例功能） */}
          <Button
            type="default"
            icon={<DownloadOutlined />}
            onClick={() => message.info('导出功能开发中...')}
          >
            导出
          </Button>
          
          {/* 出题按钮（合并了AI出题和手工出题） */}
          <Dropdown overlay={addMenu} placement="bottomRight">
            <Button type="primary" icon={<PlusOutlined />}>
              出题
            </Button>
          </Dropdown>
        </Space>
      </div>

      {/* 筛选和搜索区域 */}
      <div style={{ 
        background: '#fff', 
        padding: 16, 
        borderRadius: 8, 
        marginBottom: 16,
        border: '1px solid #f0f0f0',
      }}>
        <Row gutter={16}>
          <Col span={6}>
            <Select
              value={selectedType}
              onChange={handleTypeChange}
              style={{ width: '100%' }}
              placeholder="请选择题型"
            >
              <Option value="全部">全部题型</Option>
              <Option value="单选题">单选题</Option>
              <Option value="多选题">多选题</Option>
              <Option value="编程题">编程题</Option>
            </Select>
          </Col>
          <Col span={18}>
            <Search
              placeholder="请输入题目关键词进行搜索"
              value={searchKey}
              onChange={(e) => updateFilter(selectedType, e.target.value)}
              onSearch={handleSearch}
              enterButton={
                <Button type="primary" icon={<SearchOutlined />}>
                  搜索
                </Button>
              }
              allowClear
            />
          </Col>
        </Row>
        
        {/* 选中状态提示 */}
        {selectedIds.length > 0 && (
          <div style={{ marginTop: 16, color: '#1890ff' }}>
            已选中 {selectedIds.length} 道题目，可进行批量操作
          </div>
        )}
      </div>

      {/* 题目表格区域 */}
      <div style={{ 
        background: '#fff', 
        padding: 16, 
        borderRadius: 8,
        border: '1px solid #f0f0f0',
      }}>
        <Table
          columns={columns}
          dataSource={questionList.map(item => ({ ...item, key: item.id }))}
          pagination={{
            current: pageNum,
            pageSize: pageSize,
            total: total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            pageSizeOptions: ['10', '20', '50', '100'],
          }}
          onChange={handleTableChange}
          
          bordered
          size="middle"
          scroll={{ x: 1000 }}
        />
      </div>

      {/* 出题弹窗组件 */}
      <QuestionModal />
    </div>
  );
};

export default QuestionManager;