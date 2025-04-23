package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// 定义表结构体
type Table struct {
	Name string `ch:"name"`
}

func main() {
	// 配置 ClickHouse 连接
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"ip:8907"},
		Auth: clickhouse.Auth{
			Database: "pc_insight",
			Username: "pc_insight",
			Password: "pc_insight",
		},
		Debug:           true,
		DialTimeout:     time.Second * 10,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		TLS: nil, // 尝试不使用TLS
	})
	if err != nil {
		log.Fatalf("无法连接到 ClickHouse: %v", err)
	}
	defer conn.Close()

	// 检查连接是否有效
	if err := conn.Ping(context.Background()); err != nil {
		log.Fatalf("连接检查失败: %v", err)
	}

	log.Println("成功连接到 ClickHouse 服务器")

	// 创建新数据库
	if err := conn.Exec(context.Background(), "CREATE DATABASE IF NOT EXISTS pc_insight_important_level"); err != nil {
		log.Fatalf("创建新数据库失败: %v", err)
	}

	log.Println("成功创建数据库 pc_insight_important_level")

	// 获取 pc_insight 数据库中的所有表名
	var tableResults []Table
	query := `
		SELECT name 
		FROM system.tables 
		WHERE database = 'pc_insight'
	`
	if err := conn.Select(context.Background(), &tableResults, query); err != nil {
		log.Fatalf("获取表名失败: %v", err)
	}

	// 提取表名
	tables := make([]string, 0, len(tableResults))
	for _, t := range tableResults {
		tables = append(tables, t.Name)
	}

	log.Printf("找到 %d 个表需要复制", len(tables))

	// 复制表结构和数据
	for _, table := range tables {
		// 获取表结构
		var createTableQuery string
		if err := conn.QueryRow(context.Background(), fmt.Sprintf("SHOW CREATE TABLE pc_insight.%s", table)).Scan(&createTableQuery); err != nil {
			log.Printf("获取表 %s 结构失败: %v", table, err)
			continue
		}

		// 修改表名以适应新数据库
		newCreateTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS pc_insight_important_level.%s %s", table, createTableQuery[len(fmt.Sprintf("CREATE TABLE pc_insight.%s", table)):])

		// 在新数据库中创建表
		if err := conn.Exec(context.Background(), newCreateTableQuery); err != nil {
			log.Printf("创建表 pc_insight_important_level.%s 失败: %v", table, err)
			continue
		}

		log.Printf("成功创建表 pc_insight_important_level.%s", table)

		// 复制数据
		insertQuery := fmt.Sprintf("INSERT INTO pc_insight_important_level.%s SELECT * FROM pc_insight.%s", table, table)
		if err := conn.Exec(context.Background(), insertQuery); err != nil {
			log.Printf("复制表 %s 数据失败: %v", table, err)
		} else {
			log.Printf("成功复制表 %s 的数据", table)
		}
	}

	fmt.Println("数据库复制完成")
}
