import { createContext, createElement, useContext, useMemo, useState, type ReactNode } from 'react';
import type { User } from '../types';

interface AuthState {
  user: User | null;
  token: string;
  setAuth: (token: string, user: User) => void;
  updateUser: (user: User) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string>(() => localStorage.getItem('gbcheckup_token') || '');
  const [user, setUser] = useState<User | null>(() => {
    const raw = localStorage.getItem('gbcheckup_user');
    return raw ? (JSON.parse(raw) as User) : null;
  });

  const value = useMemo<AuthState>(
    () => ({
      user,
      token,
      setAuth: (t, u) => {
        localStorage.setItem('gbcheckup_token', t);
        localStorage.setItem('gbcheckup_user', JSON.stringify(u));
        setToken(t);
        setUser(u);
      },
      updateUser: (u) => {
        localStorage.setItem('gbcheckup_user', JSON.stringify(u));
        setUser(u);
      },
      logout: () => {
        localStorage.removeItem('gbcheckup_token');
        localStorage.removeItem('gbcheckup_user');
        setToken('');
        setUser(null);
      },
    }),
    [user, token]
  );

  return createElement(AuthContext.Provider, { value }, children);
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
