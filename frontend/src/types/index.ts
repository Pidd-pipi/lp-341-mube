export interface User {
  id: number;
  phone: string;
  name: string;
  avatar: string;
  role: string;
  department: string;
  created_at: string;
}

export interface Package {
  id: number;
  name: string;
  package_type: string;
  price: number;
  status: string;
  description: string;
}

export interface PackageItem {
  id: number;
  package_id: number;
  item_name: string;
  item_group: string;
  ref_value_range: string;
  department: string;
  sort_order: number;
}

export interface Examinee {
  id: number;
  name: string;
  id_card_no: string;
  phone: string;
  gender: string;
  age: number;
  source_type: string;
  enterprise_id?: number | null;
}

export interface Registration {
  id: number;
  examinee_id: number;
  package_id: number;
  guide_no: string;
  status: string;
  registered_at: string;
  register_user_id: number;
  examinee?: Examinee;
  package?: Package;
}

export interface ExamResult {
  id: number;
  registration_id: number;
  examinee_id: number;
  package_item_id: number;
  result_value: string;
  result_text: string;
  is_abnormal: boolean;
  status: string;
  image_url: string;
  doctor_id: number;
  package_item?: PackageItem;
}

export interface Report {
  id: number;
  registration_id: number;
  examinee_id: number;
  report_no: string;
  status: string;
  conclusion: string;
  health_advice: string;
  follow_up_reminder: string;
  pdf_url: string;
  doctor_id: number;
  examinee?: Examinee;
}

export interface AbnormalMetric {
  id: number;
  examinee_id: number;
  package_item_id: number;
  abnormal_level: string;
  value: string;
  ref_value_range: string;
  trend_json: string;
  follow_up_status: string;
  specialist_advice: string;
  package_item?: PackageItem;
}

export interface Enterprise {
  id: number;
  name: string;
  contact: string;
  phone: string;
  address: string;
}

export interface GroupOrder {
  id: number;
  enterprise_id: number;
  package_id: number;
  examinee_count: number;
  status: string;
  report_delivery_status: string;
  enterprise?: Enterprise;
  package?: Package;
}

export interface PageData<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}
