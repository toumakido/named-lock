-- AcquireProcessReleaseLockで使用するテーブル
CREATE TABLE IF NOT EXISTS process_lock_items (
  id INT AUTO_INCREMENT PRIMARY KEY,
  lock_name VARCHAR(255) NOT NULL,
  item_key VARCHAR(255) NOT NULL,
  value TEXT,
  status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  processed_at TIMESTAMP NULL,
  session_id VARCHAR(255),
  UNIQUE KEY (lock_name, item_key)
);

-- テスト用のデータを挿入
INSERT INTO process_lock_items (lock_name, item_key, value, status) VALUES
  ('test_process_lock', 'item1', '{"data": "value1"}', 'pending'),
  ('test_process_lock', 'item2', '{"data": "value2"}', 'pending'),
  ('test_process_lock', 'item3', '{"data": "value3"}', 'pending');
