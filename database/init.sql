-- lp-341 GbCheckup 体检报告数字化平台 初始化脚本（容器首次启动自动执行）
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    phone VARCHAR(20) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(50),
    avatar VARCHAR(255) DEFAULT '',
    role VARCHAR(20) DEFAULT 'examinee',
    department VARCHAR(50) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS packages (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    package_type VARCHAR(20) NOT NULL,
    price DOUBLE PRECISION DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    description VARCHAR(500) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS package_items (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL,
    item_name VARCHAR(100) NOT NULL,
    item_group VARCHAR(50) DEFAULT '',
    ref_value_range VARCHAR(100) DEFAULT '',
    department VARCHAR(50) DEFAULT '',
    sort_order INT DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_item_package ON package_items(package_id);

CREATE TABLE IF NOT EXISTS examinees (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    id_card_no VARCHAR(18) UNIQUE NOT NULL,
    phone VARCHAR(20) DEFAULT '',
    gender VARCHAR(10) DEFAULT '',
    age INT DEFAULT 0,
    source_type VARCHAR(20) DEFAULT 'personal',
    enterprise_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS registrations (
    id BIGSERIAL PRIMARY KEY,
    examinee_id BIGINT NOT NULL,
    package_id BIGINT NOT NULL,
    guide_no VARCHAR(32) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'registered',
    registered_at TIMESTAMPTZ DEFAULT now(),
    register_user_id BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_reg_examinee ON registrations(examinee_id);

CREATE TABLE IF NOT EXISTS exam_results (
    id BIGSERIAL PRIMARY KEY,
    registration_id BIGINT NOT NULL,
    examinee_id BIGINT NOT NULL,
    package_item_id BIGINT NOT NULL,
    result_value VARCHAR(100) DEFAULT '',
    result_text VARCHAR(500) DEFAULT '',
    is_abnormal BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'pending',
    image_url VARCHAR(255) DEFAULT '',
    doctor_id BIGINT DEFAULT 0,
    entered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_result_reg ON exam_results(registration_id);

CREATE TABLE IF NOT EXISTS reports (
    id BIGSERIAL PRIMARY KEY,
    registration_id BIGINT NOT NULL,
    examinee_id BIGINT NOT NULL,
    report_no VARCHAR(32) UNIQUE NOT NULL,
    status VARCHAR(20) DEFAULT 'draft',
    conclusion TEXT DEFAULT '',
    health_advice TEXT DEFAULT '',
    follow_up_reminder VARCHAR(500) DEFAULT '',
    pdf_url VARCHAR(255) DEFAULT '',
    doctor_id BIGINT DEFAULT 0,
    generated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS abnormal_metrics (
    id BIGSERIAL PRIMARY KEY,
    examinee_id BIGINT NOT NULL,
    package_item_id BIGINT NOT NULL,
    abnormal_level VARCHAR(20) NOT NULL,
    value VARCHAR(100) DEFAULT '',
    ref_value_range VARCHAR(100) DEFAULT '',
    trend_json TEXT DEFAULT '[]',
    follow_up_status VARCHAR(20) DEFAULT 'pending',
    specialist_advice VARCHAR(500) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_metric_examinee ON abnormal_metrics(examinee_id);

CREATE TABLE IF NOT EXISTS enterprises (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    contact VARCHAR(50) DEFAULT '',
    phone VARCHAR(20) DEFAULT '',
    address VARCHAR(200) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS group_orders (
    id BIGSERIAL PRIMARY KEY,
    enterprise_id BIGINT NOT NULL,
    package_id BIGINT NOT NULL,
    examinee_count INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending',
    report_delivery_status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT now()
);

-- 种子数据（密码 bcrypt：admin123/doctor123/front123/examinee123）
INSERT INTO users (phone, password_hash, name, role, department) VALUES
('13800000001', '$2a$10$lZmZzCvRaYR.pAKTsPaXnuqcYHGKshT.v9FnKf2a7YCSMd6PmOPq.', '系统管理员', 'admin', '管理部'),
('13800000002', '$2a$10$EzzUDgEDI166KFDxKY1U1eqBjRn6RREWud8QEMySCqYkCe6kunOFO', '张医生', 'doctor', '内科'),
('13800000003', '$2a$10$ZA8dJGujBbhOhv5WgI5c9eICxtMBlo3xohQ62k5RJqK.w5tGwW5P2', '前台小李', 'front_desk', '导检台'),
('13800000004', '$2a$10$7EBfjZG3KPgCFBupHx1/hOowtPJrx.lzFVY0ps1i77ckROP9PCE7C', '体检用户', 'examinee', '')
ON CONFLICT (phone) DO NOTHING;

INSERT INTO packages (name, package_type, price, status, description) VALUES
('入职基础体检', 'entry', 199, 'active', '血常规、尿常规、肝功能、胸片'),
('年度标准体检', 'annual', 599, 'active', '内外科、血生化、腹部彩超、心电图'),
('高端深度体检', 'premium', 1999, 'active', '含 CT、肿瘤标志物、心脑血管深度筛查');

INSERT INTO package_items (package_id, item_name, item_group, ref_value_range, department, sort_order) VALUES
(1, '血常规', '检验', '3.5-9.5', '检验科', 1),
(1, '肝功能ALT', '检验', '9-50', '检验科', 2),
(1, '胸片', '影像', '正常', '放射科', 3),
(2, '空腹血糖', '生化', '3.9-6.1', '检验科', 1),
(2, '心电图', '心电', '正常', '功能科', 2),
(3, '胸部CT', '影像', '正常', '放射科', 1),
(3, '肿瘤标志物CEA', '检验', '0-5', '检验科', 2);

INSERT INTO examinees (name, id_card_no, phone, gender, age, source_type) VALUES
('王小明', '110101199501011111', '13911110001', 'male', 30, 'personal'),
('李小红', '310101198812121222', '13911110002', 'female', 36, 'personal'),
('赵大强', '440101199003033333', '13911110003', 'male', 34, 'personal')
ON CONFLICT (id_card_no) DO NOTHING;

INSERT INTO registrations (examinee_id, package_id, guide_no, status, register_user_id) VALUES
(1, 1, 'GUIDE202608160001', 'in_progress', 3),
(2, 2, 'GUIDE202608160002', 'registered', 3)
ON CONFLICT (guide_no) DO NOTHING;

INSERT INTO exam_results (registration_id, examinee_id, package_item_id, result_value, is_abnormal, status, doctor_id, entered_at) VALUES
(1, 1, 1, '5.2', FALSE, 'entered', 2, now()),
(1, 1, 2, '72', TRUE, 'entered', 2, now()),
(1, 1, 3, '正常', FALSE, 'pending', 0, NULL),
(2, 2, 4, '5.5', FALSE, 'pending', 0, NULL),
(2, 2, 5, '正常', FALSE, 'pending', 0, NULL);

INSERT INTO abnormal_metrics (examinee_id, package_item_id, abnormal_level, value, ref_value_range, follow_up_status) VALUES
(1, 2, 'moderate', '72', '9-50', 'pending');

INSERT INTO reports (registration_id, examinee_id, report_no, status, doctor_id) VALUES
(1, 1, 'GB202608160001', 'draft', 2)
ON CONFLICT (report_no) DO NOTHING;

INSERT INTO enterprises (name, contact, phone, address) VALUES
('华信科技', '陈经理', '021-88886666', '上海市浦东新区');

INSERT INTO group_orders (enterprise_id, package_id, examinee_count, status) VALUES
(1, 2, 50, 'confirmed');
