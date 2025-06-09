package config

// Config はアプリケーション設定を保持する構造体
type Config struct {
	DB DBConfig
}

// DBConfig はデータベース接続設定を保持する構造体
type DBConfig struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// NewConfig は新しい設定インスタンスを作成する
func NewConfig() *Config {
	return &Config{
		DB: DBConfig{
			Driver:   "mysql",
			Host:     "localhost",
			Port:     "3306",
			User:     "user",
			Password: "password",
			DBName:   "locktest",
		},
	}
}

// GetDSN はデータベース接続文字列を返す
func (c *DBConfig) GetDSN() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + c.Port + ")/" + c.DBName + "?parseTime=true"
}
