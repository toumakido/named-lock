-- 商品テーブル
CREATE TABLE IF NOT EXISTS products (
  code VARCHAR(50) PRIMARY KEY,
  quantity INT NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
