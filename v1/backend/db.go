package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func initDB() {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
        DB_USER, DB_PASS, DB_HOST, DB_PORT, DB_NAME)
    var err error
    db, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal("数据库连接失败:", err)
    }
    if err = db.Ping(); err != nil {
        log.Fatal("数据库 Ping 失败:", err)
    }
    log.Println("数据库已连接")
}