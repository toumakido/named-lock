-- 商品テーブル
CREATE TABLE IF NOT EXISTS products (
  id INT AUTO_INCREMENT PRIMARY KEY,
  product_code VARCHAR(50) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  price DECIMAL(10, 2) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 商品在庫テーブル
CREATE TABLE IF NOT EXISTS product_inventory (
  id INT AUTO_INCREMENT PRIMARY KEY,
  product_id INT NOT NULL,
  quantity INT NOT NULL DEFAULT 0,
  last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (product_id) REFERENCES products(id)
);

-- テスト用のデータを挿入
INSERT INTO products (product_code, name, price) VALUES
  ('P001', '商品1', 1000),
  ('P002', '商品2', 2000),
  ('P003', '商品3', 3000);

-- 初期在庫を設定
INSERT INTO product_inventory (product_id, quantity) VALUES
  (1, 10),
  (2, 20),
  (3, 30);
