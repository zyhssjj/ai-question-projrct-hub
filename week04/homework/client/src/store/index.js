import { create } from 'zustand';
import { message } from 'antd';
import request from '../utils/request';

/**
 * 创建全局状态存储
 * 使用Zustand状态管理库，类似Redux但更简洁
 * 作用：管理整个应用的状态，包括题目数据、弹窗状态、筛选条件等
 */
const useStore = create((set, get) => ({
  // ==================== 状态定义区域 ====================
  
  // 1. 题库相关状态
  questionList: [], // 题目列表数据
  total: 0, // 题目总数（用于分页）
  pageNum: 1, // 当前页码（默认第1页）
  pageSize: 10, // 每页显示数量（默认10条）
  selectedType: '全部', // 当前选中的题型
  searchKey: '', // 搜索关键词
  selectedIds: [], // 选中的题目ID（用于批量删除）
  
  // 2. 弹窗相关状态
  isModalOpen: false, // 弹窗是否显示
  modalType: 'add', // 弹窗类型：add(添加)/edit(编辑)/ai(AI出题)
  currentQuestion: null, // 当前正在编辑的题目
  
  // 3. AI出题相关状态
  aiParams: {
    type: '单选题', // AI出题题型
    count: 3, // 生成题目数量（1-10）
    difficulty: '中等', // 题目难度
    language: 'Go', // 编程语言
  },
  aiPreviewList: [], // AI生成的题目预览列表
  selectedAiIds: [], // 选中的AI题目ID

  // ==================== 状态操作方法区域 ====================

  // 1. 更新筛选条件并重新查询
  updateFilter: (type, key) => {
    set({
      selectedType: type,
      searchKey: key,
      pageNum: 1, // 筛选后重置到第一页
    });
    get().fetchQuestionList();
  },

  // 2. 更新分页参数并重新查询
  updatePage: (pageNum, pageSize) => {
    set({ pageNum, pageSize });
    get().fetchQuestionList();
  },

  // 3. 获取题目列表（核心方法）
  fetchQuestionList: async () => {
    const { pageNum, pageSize, selectedType, searchKey } = get();
    
    try {
      // 发送GET请求获取题目列表
      const res = await request({
        url: '/question/list',
        method: 'GET',
        params: {
          pageNum,
          pageSize,
          type: selectedType === '全部' ? '' : selectedType,
          keyword: searchKey,
        },
      });
      
      // 更新状态：题目列表和总数
      set({
        questionList: res.data.list || [],
        total: res.data.total || 0,
      });
    } catch (error) {
      console.error('获取题目列表失败：', error);
      // 如果请求失败，显示空列表
      set({ questionList: [], total: 0 });
    }
  },

  // 4. 打开弹窗
  openModal: (type, question = null) => {
    set({
      isModalOpen: true,
      modalType: type,
      currentQuestion: question,
      selectedAiIds: [], // 重置AI选中状态
    });
  },

  // 5. 关闭弹窗
  closeModal: () => {
    set({
      isModalOpen: false,
      currentQuestion: null,
      aiPreviewList: [], // 清空AI预览列表
    });
  },

  // 6. 更新AI出题参数
  updateAiParams: (key, value) => {
    set((state) => ({
      aiParams: { ...state.aiParams, [key]: value },
    }));
  },

  // 7. 调用AI生成题目
  fetchAiQuestions: async () => {
    const { aiParams } = get();
    
    // 参数验证
    if (aiParams.count < 1 || aiParams.count > 10) {
      message.warning('题目数量必须在1-10之间！');
      return;
    }
    
    try {
      // 显示加载提示
      message.loading('AI正在生成题目中...', 0);
      
      // 发送POST请求到AI生成接口
      const res = await request({
        url: '/question/ai-generate',
        method: 'POST',
        data: aiParams,
      });
      
      // 给生成的题目添加临时ID（用于选择）
      const previewList = (res.data || []).map((item, index) => ({
        ...item,
        id: `ai-${Date.now()}-${index}`, // 临时ID，添加到数据库后会生成真实ID
      }));
      
      // 更新AI预览列表
      set({ aiPreviewList: previewList });
      
      // 关闭加载提示
      message.destroy();
      message.success('AI题目生成成功！');
    } catch (error) {
      console.error('AI生成题目失败：', error);
      message.destroy();
      message.error('AI生成题目失败，请重试！');
      
      // 如果AI服务失败，使用模拟数据（用于演示）
      const mockQuestions = Array.from({ length: aiParams.count }, (_, index) => ({
        id: `mock-${index}`,
        type: aiParams.type,
        content: `这是第${index + 1}道${aiParams.type}示例题目（AI服务暂时不可用）`,
        difficulty: aiParams.difficulty,
        language: aiParams.language,
        options: aiParams.type !== '编程题' 
          ? JSON.stringify(['A. 选项A', 'B. 选项B', 'C. 选项C', 'D. 选项D'])
          : '',
        answer: aiParams.type !== '编程题' ? 'A' : '',
      }));
      
      set({ aiPreviewList: mockQuestions });
    }
  },

  // 8. 切换AI题目选择状态
  toggleAiSelect: (id) => {
    set((state) => {
      const isSelected = state.selectedAiIds.includes(id);
      return {
        selectedAiIds: isSelected
          ? state.selectedAiIds.filter((itemId) => itemId !== id)
          : [...state.selectedAiIds, id],
      };
    });
  },

  // 9. 批量删除题目
  batchDeleteQuestion: async () => {
    const { selectedIds } = get();
    
    if (selectedIds.length === 0) {
      message.warning('请选择要删除的题目！');
      return;
    }
    
    try {
      // 发送删除请求
      await request({
        url: '/question/delete',
        method: 'POST',
        data: { ids: selectedIds },
      });
      
      // 重新获取题目列表
      get().fetchQuestionList();
      // 清空选中状态
      set({ selectedIds: [] });
      
      message.success('删除成功！');
    } catch (error) {
      console.error('批量删除失败：', error);
      message.error('删除失败，请重试！');
    }
  },

  // 10. 删除单个题目
  deleteQuestion: async (id) => {
    try {
      await request({
        url: '/question/delete',
        method: 'POST',
        data: { ids: [id] },
      });
      
      get().fetchQuestionList();
      message.success('删除成功！');
    } catch (error) {
      console.error('删除题目失败：', error);
      message.error('删除失败，请重试！');
    }
  },

  // 11. 保存题目（添加/编辑共用）
  saveQuestion: async (question) => {
    const { modalType, currentQuestion } = get();
    
    try {
      if (modalType === 'edit') {
        // 编辑模式：需要包含题目ID
        await request({
          url: '/question/update',
          method: 'POST',
          data: { ...question, id: currentQuestion.id },
        });
      } else {
        // 添加模式
        await request({
          url: '/question/add',
          method: 'POST',
          data: question,
        });
      }
      
      // 关闭弹窗并刷新列表
      get().closeModal();
      get().fetchQuestionList();
      
      message.success(modalType === 'edit' ? '编辑成功！' : '添加成功！');
    } catch (error) {
      console.error('保存题目失败：', error);
      message.error(modalType === 'edit' ? '编辑失败！' : '添加失败！');
    }
  },

  // 12. 确认添加AI题目
  confirmAiQuestions: async () => {
    const { aiPreviewList, selectedAiIds } = get();
    
    // 筛选选中的题目
    const selectedQuestions = aiPreviewList.filter((item) =>
      selectedAiIds.includes(item.id)
    );
    
    if (selectedQuestions.length === 0) {
      message.warning('请选择要添加的题目！');
      return;
    }
    
    try {
      // 移除临时ID，发送到后端
      const questionsToSave = selectedQuestions.map(({ id, ...rest }) => rest);
      
      await request({
        url: '/question/add-batch',
        method: 'POST',
        data: questionsToSave,
      });
      
      // 关闭弹窗并刷新列表
      get().closeModal();
      get().fetchQuestionList();
      
      message.success(`成功添加${selectedQuestions.length}道题目！`);
    } catch (error) {
      console.error('添加AI题目失败：', error);
      message.error('添加失败，请重试！');
    }
  },

  // 13. 切换题目选择状态（用于批量删除）
  toggleQuestionSelect: (id) => {
    set((state) => {
      const isSelected = state.selectedIds.includes(id);
      return {
        selectedIds: isSelected
          ? state.selectedIds.filter((itemId) => itemId !== id)
          : [...state.selectedIds, id],
      };
    });
  },

  // 14. 全选/取消全选
  selectAllQuestions: (selected) => {
    const { questionList } = get();
    
    if (selected) {
      // 全选：获取所有题目的ID
      set({ selectedIds: questionList.map((item) => item.id) });
    } else {
      // 取消全选：清空数组
      set({ selectedIds: [] });
    }
  },
}));

export default useStore;