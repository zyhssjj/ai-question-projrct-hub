import axios from 'axios';
import { message } from 'antd';

/**
 * 创建axios实例
 * 作用：统一配置所有HTTP请求
 */
const service = axios.create({
  baseURL: '/api', // 基础URL，会被vite代理转发到后端
  timeout: 30000, // 请求超时时间（30秒）
  headers: {
    'Content-Type': 'application/json', // 请求头类型
  },
});

/**
 * 请求拦截器
 * 作用：在发送请求之前做一些处理（如添加token等）
 */
service.interceptors.request.use(
  (config) => {
    // 可以在这里添加统一的请求头，如token
    // const token = localStorage.getItem('token');
    // if (token) {
    //   config.headers.Authorization = `Bearer ${token}`;
    // }
    return config;
  },
  (error) => {
    // 请求错误处理
    console.error('请求错误：', error);
    message.error('请求失败，请检查网络连接！');
    return Promise.reject(error);
  }
);

/**
 * 响应拦截器
 * 作用：在收到响应后做一些统一处理
 */
service.interceptors.response.use(
  (response) => {
    // 响应成功处理
    const res = response.data;
    
    // 检查业务状态码（根据后端返回的格式）
    if (res.code !== 200) {
      // 如果业务状态码不是200，显示错误消息
      message.error(res.msg || '请求出错！');
      return Promise.reject(new Error(res.msg || 'Error'));
    }
    
    // 返回响应数据
    return res;
  },
  (error) => {
    // 响应错误处理
    console.error('响应错误：', error);
    
    if (error.response) {
      // 服务器返回了错误状态码
      const status = error.response.status;
      const msg = error.response.data?.msg || '请求失败';
      
      switch (status) {
        case 400:
          message.error(`参数错误：${msg}`);
          break;
        case 401:
          message.error('未授权，请重新登录！');
          // 可以在这里跳转到登录页
          break;
        case 403:
          message.error('禁止访问！');
          break;
        case 404:
          message.error('请求的资源不存在！');
          break;
        case 500:
          message.error(`服务器错误：${msg}`);
          break;
        default:
          message.error(`请求失败：${status}`);
      }
    } else if (error.request) {
      // 请求发送了但没有收到响应
      message.error('网络连接失败，请检查网络设置！');
    } else {
      // 请求配置出错
      message.error(`请求配置错误：${error.message}`);
    }
    
    return Promise.reject(error);
  }
);

export default service;