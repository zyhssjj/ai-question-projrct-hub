import axios from 'axios';

async function testProxy() {
  console.log('🔍 测试Vite代理...');
  
  try {
    // 测试通过前端访问（代理）
    console.log('\n1. 测试前端代理: http://localhost:5173/api/note/get');
    const response = await axios.get('http://localhost:5173/api/note/get', {
      timeout: 5000
    });
    console.log('✅ 代理成功:', response.status, response.data.code);

    // 测试直接后端
    console.log('\n2. 测试直接后端: http://localhost:8080/api/note/get');
    const directResponse = await axios.get('http://localhost:8080/api/note/get', {
      timeout: 5000
    });
    console.log('✅ 直接访问成功:', directResponse.status, directResponse.data.code);

  } catch (error) {
    console.error('❌ 测试失败:');
    if (error.code === 'ECONNREFUSED') {
      console.error('连接被拒绝，请检查服务是否启动');
    } else if (error.response) {
      console.error('HTTP错误:', error.response.status, error.response.data);
    } else {
      console.error('其他错误:', error.message);
    }
  }
}

testProxy();