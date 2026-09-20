import request from '../utils/request';
import type { User } from '../types';

export interface LoginResult {
  token: string;
  user: User;
}

export function login(phone: string, password: string): Promise<LoginResult> {
  return request.post('/auth/login', { phone, password });
}

export function register(phone: string, password: string, name: string): Promise<LoginResult> {
  return request.post('/auth/register', { phone, password, name });
}

export function getMe(): Promise<User> {
  return request.get('/users/me');
}

export function updateProfile(data: { name?: string; avatar?: string; department?: string }): Promise<User> {
  return request.put('/users/me', data);
}
