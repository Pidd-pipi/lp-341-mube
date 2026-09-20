import { getMe, updateProfile } from '../api/user';
import type { User } from '../types';
import { useAuth } from './authStore';

export function useUserStore() {
  const { user, updateUser } = useAuth();

  async function refresh(): Promise<User> {
    const me = await getMe();
    updateUser(me);
    return me;
  }

  async function save(data: { name?: string; avatar?: string; department?: string }): Promise<User> {
    const updated = await updateProfile(data);
    updateUser(updated);
    return updated;
  }

  return { user, refresh, save };
}
