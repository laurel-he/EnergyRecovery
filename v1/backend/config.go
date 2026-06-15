package main

const (
    DB_USER   = "root"
    DB_PASS   = "your_mysql_password"   // 修改为你的MySQL密码
    DB_HOST   = "127.0.0.1"
    DB_PORT   = "3306"
    DB_NAME   = "energy_recovery"
    SERVER_PORT = ":8080"

    // DeepSeek 云端 API（填你自己的key）
    DEEPSEEK_API_KEY = "sk-your-api-key"
    DEEPSEEK_API_URL = "https://api.deepseek.com/chat/completions"
    DEEPSEEK_MODEL   = "deepseek-chat"
)