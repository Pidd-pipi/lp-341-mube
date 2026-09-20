// 与 backend/internal/constants/report.go 保持一致
export const AbnormalLevel = {
  MILD: 'mild',
  MODERATE: 'moderate',
  SEVERE: 'severe',
} as const;

export const AbnormalLevelLabels: Record<string, string> = {
  [AbnormalLevel.MILD]: '轻度异常',
  [AbnormalLevel.MODERATE]: '中度异常',
  [AbnormalLevel.SEVERE]: '重度异常',
};

export const ReportStatus = {
  DRAFT: 'draft',
  GENERATED: 'generated',
  REVIEWED: 'reviewed',
  PUBLISHED: 'published',
} as const;

export const ReportStatusLabels: Record<string, string> = {
  [ReportStatus.DRAFT]: '草稿',
  [ReportStatus.GENERATED]: '已生成',
  [ReportStatus.REVIEWED]: '已审核',
  [ReportStatus.PUBLISHED]: '已发布',
};

export const ResultStatusLabels: Record<string, string> = {
  pending: '待录入',
  entered: '已录入',
  reviewed: '已审核',
};

export const RegistrationStatusLabels: Record<string, string> = {
  registered: '已登记',
  in_progress: '进行中',
  completed: '已完成',
};

export const PackageStatusLabels: Record<string, string> = {
  active: '启用',
  inactive: '停用',
};

export const PackageTypeLabels: Record<string, string> = {
  entry: '入职体检',
  annual: '年度体检',
  premium: '高端体检',
  other: '其他',
};

export const FollowUpStatusLabels: Record<string, string> = {
  pending: '待复查',
  done: '已复查',
};
