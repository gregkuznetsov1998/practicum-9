CREATE TABLE IF NOT EXISTS crm_data (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    branch VARCHAR(50),
    employee_type VARCHAR(50),
    registration_date DATE,
    last_activity_date DATE,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Вставляем данные для пользователей из LDAP
INSERT INTO crm_data (user_id, branch, employee_type, registration_date, last_activity_date) VALUES
('user1', 'MSK001', 'regular', '2023-01-15', CURRENT_DATE - INTERVAL '1 day'),
('user2', 'MSK002', 'regular', '2023-02-20', CURRENT_DATE - INTERVAL '2 days'),
('admin1', 'MSK001', 'administrator', '2023-01-10', CURRENT_DATE),
('prothetic1', 'SPB001', 'prothetic', '2023-03-05', CURRENT_DATE - INTERVAL '3 days'),
('prothetic2', 'SPB002', 'prothetic', '2023-03-10', CURRENT_DATE - INTERVAL '4 days'),
('prothetic3', 'EKB001', 'prothetic', '2023-04-01', CURRENT_DATE - INTERVAL '5 days');