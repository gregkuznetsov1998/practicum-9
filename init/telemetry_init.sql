CREATE TABLE IF NOT EXISTS telemetry_data (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    device_id VARCHAR(50) NOT NULL,
    steps_count INTEGER,
    battery_level INTEGER,
    activity_level FLOAT,
    temperature FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Генерация тестовых данных для пользователей из LDAP
-- Для каждого пользователя создаем отдельные записи с разными временными метками
INSERT INTO telemetry_data (user_id, timestamp, device_id, steps_count, battery_level, activity_level, temperature)
VALUES
-- user1 - 10 записей с разными временными метками
('user1', NOW() - INTERVAL '1 day', 'device_1', 8456, 85, 7.2, 36.8),
('user1', NOW() - INTERVAL '2 days', 'device_1', 7234, 82, 6.8, 36.7),
('user1', NOW() - INTERVAL '3 days', 'device_1', 9123, 78, 7.5, 36.9),
('user1', NOW() - INTERVAL '4 days', 'device_1', 6543, 75, 6.2, 36.6),
('user1', NOW() - INTERVAL '5 days', 'device_1', 7890, 80, 7.0, 36.8),
('user1', NOW() - INTERVAL '6 days', 'device_1', 8345, 83, 6.9, 36.7),
('user1', NOW() - INTERVAL '7 days', 'device_1', 7012, 79, 6.5, 36.6),
('user1', NOW() - INTERVAL '8 days', 'device_1', 8765, 81, 7.3, 36.9),
('user1', NOW() - INTERVAL '9 days', 'device_1', 8123, 84, 7.1, 36.8),
('user1', NOW() - INTERVAL '10 days', 'device_1', 7654, 77, 6.7, 36.7),

-- user2 - 10 записей
('user2', NOW() - INTERVAL '1 day', 'device_2', 5678, 88, 5.8, 36.6),
('user2', NOW() - INTERVAL '2 days', 'device_2', 6345, 85, 6.1, 36.7),
('user2', NOW() - INTERVAL '3 days', 'device_2', 5890, 82, 5.9, 36.5),
('user2', NOW() - INTERVAL '4 days', 'device_2', 7123, 79, 6.3, 36.8),
('user2', NOW() - INTERVAL '5 days', 'device_2', 6456, 83, 6.0, 36.6),
('user2', NOW() - INTERVAL '6 days', 'device_2', 5987, 80, 5.7, 36.5),
('user2', NOW() - INTERVAL '7 days', 'device_2', 6734, 86, 6.2, 36.7),
('user2', NOW() - INTERVAL '8 days', 'device_2', 6210, 81, 5.8, 36.6),
('user2', NOW() - INTERVAL '9 days', 'device_2', 6890, 84, 6.1, 36.7),
('user2', NOW() - INTERVAL '10 days', 'device_2', 6543, 82, 5.9, 36.6),

-- admin1 - 10 записей
('admin1', NOW() - INTERVAL '1 day', 'device_3', 9234, 92, 8.1, 37.0),
('admin1', NOW() - INTERVAL '2 days', 'device_3', 8765, 89, 7.8, 36.9),
('admin1', NOW() - INTERVAL '3 days', 'device_3', 9456, 91, 8.2, 37.1),
('admin1', NOW() - INTERVAL '4 days', 'device_3', 8987, 87, 7.6, 36.8),
('admin1', NOW() - INTERVAL '5 days', 'device_3', 9123, 90, 8.0, 37.0),
('admin1', NOW() - INTERVAL '6 days', 'device_3', 8678, 88, 7.7, 36.9),
('admin1', NOW() - INTERVAL '7 days', 'device_3', 9345, 93, 8.3, 37.1),
('admin1', NOW() - INTERVAL '8 days', 'device_3', 8890, 86, 7.5, 36.8),
('admin1', NOW() - INTERVAL '9 days', 'device_3', 9567, 94, 8.4, 37.2),
('admin1', NOW() - INTERVAL '10 days', 'device_3', 9012, 89, 7.9, 36.9),

-- prothetic1 - 10 записей
('prothetic1', NOW() - INTERVAL '1 day', 'device_4', 4567, 95, 4.5, 36.4),
('prothetic1', NOW() - INTERVAL '2 days', 'device_4', 4234, 92, 4.2, 36.3),
('prothetic1', NOW() - INTERVAL '3 days', 'device_4', 4789, 96, 4.7, 36.5),
('prothetic1', NOW() - INTERVAL '4 days', 'device_4', 4123, 91, 4.1, 36.2),
('prothetic1', NOW() - INTERVAL '5 days', 'device_4', 4654, 94, 4.6, 36.4),
('prothetic1', NOW() - INTERVAL '6 days', 'device_4', 4321, 90, 4.3, 36.3),
('prothetic1', NOW() - INTERVAL '7 days', 'device_4', 4876, 97, 4.8, 36.5),
('prothetic1', NOW() - INTERVAL '8 days', 'device_4', 4210, 93, 4.2, 36.3),
('prothetic1', NOW() - INTERVAL '9 days', 'device_4', 4732, 95, 4.6, 36.4),
('prothetic1', NOW() - INTERVAL '10 days', 'device_4', 4456, 92, 4.4, 36.3),

-- prothetic2 - 10 записей
('prothetic2', NOW() - INTERVAL '1 day', 'device_5', 5123, 87, 5.2, 36.5),
('prothetic2', NOW() - INTERVAL '2 days', 'device_5', 4876, 85, 4.9, 36.4),
('prothetic2', NOW() - INTERVAL '3 days', 'device_5', 5234, 88, 5.3, 36.6),
('prothetic2', NOW() - INTERVAL '4 days', 'device_5', 4765, 84, 4.8, 36.4),
('prothetic2', NOW() - INTERVAL '5 days', 'device_5', 5345, 89, 5.4, 36.6),
('prothetic2', NOW() - INTERVAL '6 days', 'device_5', 4921, 86, 4.9, 36.5),
('prothetic2', NOW() - INTERVAL '7 days', 'device_5', 5456, 90, 5.5, 36.7),
('prothetic2', NOW() - INTERVAL '8 days', 'device_5', 4832, 83, 4.8, 36.4),
('prothetic2', NOW() - INTERVAL '9 days', 'device_5', 5267, 91, 5.3, 36.6),
('prothetic2', NOW() - INTERVAL '10 days', 'device_5', 4987, 87, 5.0, 36.5),

-- prothetic3 - 10 записей
('prothetic3', NOW() - INTERVAL '1 day', 'device_1', 3890, 82, 3.9, 36.3),
('prothetic3', NOW() - INTERVAL '2 days', 'device_1', 3654, 80, 3.7, 36.2),
('prothetic3', NOW() - INTERVAL '3 days', 'device_1', 4012, 83, 4.0, 36.4),
('prothetic3', NOW() - INTERVAL '4 days', 'device_1', 3543, 79, 3.5, 36.1),
('prothetic3', NOW() - INTERVAL '5 days', 'device_1', 4123, 84, 4.1, 36.4),
('prothetic3', NOW() - INTERVAL '6 days', 'device_1', 3789, 81, 3.8, 36.3),
('prothetic3', NOW() - INTERVAL '7 days', 'device_1', 4234, 85, 4.2, 36.5),
('prothetic3', NOW() - INTERVAL '8 days', 'device_1', 3678, 78, 3.7, 36.2),
('prothetic3', NOW() - INTERVAL '9 days', 'device_1', 3987, 86, 4.0, 36.4),
('prothetic3', NOW() - INTERVAL '10 days', 'device_1', 3821, 82, 3.8, 36.3);