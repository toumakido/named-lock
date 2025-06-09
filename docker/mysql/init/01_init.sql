CREATE TABLE IF NOT EXISTS users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- テスト用のユーザーデータを挿入
INSERT INTO users (name, email) VALUES
  ('ユーザー1', 'user1@example.com'),
  ('ユーザー2', 'user2@example.com'),
  ('ユーザー3', 'user3@example.com');

-- ロック情報を記録するテーブル（オプション：ロックの履歴を保存する場合）
CREATE TABLE IF NOT EXISTS lock_history (
  id INT AUTO_INCREMENT PRIMARY KEY,
  lock_name VARCHAR(255) NOT NULL,
  session_id VARCHAR(255) NOT NULL,
  acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  released_at TIMESTAMP NULL,
  status ENUM('acquired', 'released') DEFAULT 'acquired'
);
