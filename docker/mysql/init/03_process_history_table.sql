-- ProcessItemInOtherTableで使用するテーブル
CREATE TABLE IF NOT EXISTS process_history (
  id INT AUTO_INCREMENT PRIMARY KEY,
  lock_name VARCHAR(255) NOT NULL,
  item_key VARCHAR(255) NOT NULL,
  value TEXT,
  processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX (lock_name, item_key)
);

-- process_lock_itemsテーブルにquantityカラムを追加
ALTER TABLE process_lock_items
ADD COLUMN quantity INT DEFAULT 0 AFTER value;
