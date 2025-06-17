-- 在庫履歴テーブル
CREATE TABLE IF NOT EXISTS inventory_history (
  id INT AUTO_INCREMENT PRIMARY KEY,
  product_id INT NOT NULL,
  quantity INT NOT NULL,
  action VARCHAR(50) NOT NULL, -- 'add', 'subtract', 'adjust' など
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (product_id) REFERENCES products(id)
);
