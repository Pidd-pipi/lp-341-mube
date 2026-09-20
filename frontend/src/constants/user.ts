export const UserRole = {
  ADMIN: 'admin',
  DOCTOR: 'doctor',
  FRONT_DESK: 'front_desk',
  EXAMINEE: 'examinee',
} as const;

export const UserRoleLabels: Record<string, string> = {
  [UserRole.ADMIN]: '管理员',
  [UserRole.DOCTOR]: '医生',
  [UserRole.FRONT_DESK]: '前台',
  [UserRole.EXAMINEE]: '体检人',
};
