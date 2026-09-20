import axios from 'axios';
import { message } from 'antd';

const request = axios.create({ baseURL: '/api/v1', timeout: 15000 });

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('gbcheckup_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

request.interceptors.response.use(
  (response) => {
    const body = response.data;
    if (body && typeof body.code === 'number' && body.code !== 0) {
      message.error(body.message || '请求失败');
      return Promise.reject(new Error(body.message));
    }
    return body?.data ?? body;
  },
  (error) => {
    const status = error.response?.status;
    const msg = error.response?.data?.message || error.message || '网络错误';
    if (status === 401) {
      localStorage.removeItem('gbcheckup_token');
      localStorage.removeItem('gbcheckup_user');
      if (!location.pathname.startsWith('/login')) location.href = '/login';
    }
    message.error(msg);
    return Promise.reject(error);
  }
);

export default request;
