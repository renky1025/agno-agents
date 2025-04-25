package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	db          *sql.DB
	writeStmt   = regexp.MustCompile(`(?i)^\s*(INSERT|UPDATE|DELETE)`)
	selectStmt  = regexp.MustCompile(`(?i)^\s*SELECT`)
	createStmt  = regexp.MustCompile(`(?i)^\s*CREATE TABLE`)
	explainStmt = regexp.MustCompile(`(?i)^\s*EXPLAIN`)
)

type DB struct {
	*sql.DB
}

func init() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file:", err)
		}
	}
}

var (
	host     string
	port     string
	name     string
	user     string
	password string
	sslmode  string
)

func main() {
	// 初始化数据库连接池

	flag.StringVar(&host, "host", "", "POSTGRES HOST")
	flag.StringVar(&port, "port", "", "POSTGRES PORT")
	flag.StringVar(&name, "name", "", "POSTGRES NAME")
	flag.StringVar(&user, "user", "", "POSTGRES USER")
	flag.StringVar(&password, "password", "", "POSTGRES PASSWORD")
	flag.StringVar(&sslmode, "sslmode", "", "POSTGRES SSLMODE")
	flag.Parse()

	dbconfig := PDBCONNECTION{
		Host:     host,
		Port:     port,
		Name:     name,
		User:     user,
		Password: password,
		SSLMODE:  sslmode,
	}
	if err := initConnectionPool(dbconfig); err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	s := server.NewMCPServer(
		"postgresql-mcp-server 🚀",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	s.AddTool(createReadQueryTool(), readQueryToolHandler)
	// s.AddTool(createWriteQueryTool(), writeQueryToolHandler)
	// s.AddTool(createCreateTableTool(), createTableToolHandler)
	s.AddTool(createListTablesTool(), listTableToolHandler)
	// s.AddTool(createExplainQueryTool(), explainQueryToolHandler)
	// s.AddTool(createCreateIndexTool(), createIndexToolHandler)
	s.AddTool(createDescribeTableTool(), describeTableToolHandler)

	if err := server.ServeStdio(s); err != nil {
		log.Printf("Server error: %v\n", err)
	}
}

func createDescribeTableTool() mcp.Tool {
	return mcp.NewTool("postgres_describe_table",
		mcp.WithDescription("Describe a table in the postgres database, query the table structure in the postgres database, 描述一个表在postgres数据库中，查询一个表的结构在postgres数据库中"),
		mcp.WithString("table_name",
			mcp.Required(),
			mcp.Description("The table name to describe, 要描述的表名"),
		),
	)
}

func createReadQueryTool() mcp.Tool {
	return mcp.NewTool("postgres_execute_query",
		mcp.WithDescription("Execute a SELECT query on the postgres database, 执行一个SELECT查询在postgres数据库上"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("SELECT SQL query to execute, 执行一个SELECT查询"),
		),
	)
}

func createListTablesTool() mcp.Tool {
	return mcp.NewTool("postgres_list_tables",
		mcp.WithDescription("List all user tables in the database, 列出数据库中的所有用户表"),
		mcp.WithString("table_name",
			mcp.Description("Optional table name to filter tables, 可选的表名来过滤表"),
		),
	)
}

type PDBCONNECTION struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMODE  string
}

// func getDefaultDBConnection() map[string]string {
// 	return map[string]string{
// 		"DB_HOST":     "10.100.2.1",
// 		"DB_PORT":     "5433",
// 		"DB_NAME":     "aiproxy",
// 		"DB_USER":     "username",
// 		"DB_PASSWORD": "password",
// 		"DB_SSLMODE":  "disable",
// 	}

// }

func initConnectionPool(dbconfig PDBCONNECTION) error {
	port, err := strconv.Atoi(dbconfig.Port)
	if err != nil {
		return fmt.Errorf("invalid DB_PORT: %w", err)
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbconfig.Host,
		port,
		dbconfig.User,
		dbconfig.Password,
		dbconfig.Name,
		dbconfig.SSLMODE,
	)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	if err = db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	log.Println("Successfully connected to database")
	return nil
}

func readQueryToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("invalid query parameter")
	}

	if !selectStmt.MatchString(query) {
		return nil, errors.New("only SELECT queries are allowed")
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Query error: %v\n", err)
		return nil, fmt.Errorf("query execution failed")
	}
	defer rows.Close()

	results, err := parseSQLRows(rows)
	if err != nil {
		return nil, fmt.Errorf("result parsing failed")
	}

	return mcp.NewToolResultText(fmt.Sprintf("Query results: %v", results)), nil
}

func writeQueryToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("invalid query parameter")
	}

	if !writeStmt.MatchString(query) {
		return nil, errors.New("only INSERT/UPDATE/DELETE queries are allowed")
	}

	result, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Write error: %v\n", err)
		return nil, fmt.Errorf("write operation failed")
	}

	rowsAffected, _ := result.RowsAffected()
	return mcp.NewToolResultText(fmt.Sprintf("Operation successful. Rows affected: %d", rowsAffected)), nil
}

func createTableToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	schema, ok := request.Params.Arguments["schema"].(string)
	if !ok {
		return nil, errors.New("invalid schema parameter")
	}

	if !createStmt.MatchString(schema) {
		return nil, errors.New("invalid CREATE TABLE statement")
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		log.Printf("Create table error: %v\n", err)
		return nil, fmt.Errorf("table creation failed")
	}

	return mcp.NewToolResultText("Table created successfully"), nil
}

func listTableToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	schemaFilter := ""
	if schema, ok := request.Params.Arguments["schema"].(string); ok {
		schemaFilter = fmt.Sprintf(" AND schemaname = '%s'", sanitizeInput(schema))
	}

	query := fmt.Sprintf(`
		SELECT tablename 
		FROM pg_catalog.pg_tables 
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema') %s
	`, schemaFilter)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("List tables error: %v\n", err)
		return nil, fmt.Errorf("failed to list tables")
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("error scanning table name")
		}
		tables = append(tables, table)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Tables: %v", tables)), nil
}

// explain查询处理函数
func explainQueryToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["schema"].(string)
	if !ok {
		return nil, errors.New("invalid schema parameter")
	}

	if !explainStmt.MatchString(query) {
		return nil, errors.New("invalid explain schema parameter")
	}

	// 执行EXPLAIN查询
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Explain error: %v\n", err)
		return nil, fmt.Errorf("explain execution failed: %v", err)
	}
	defer rows.Close()

	// 解析结果
	var plan strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return nil, fmt.Errorf("error scanning explain result: %v", err)
		}
		plan.WriteString(line)
		plan.WriteString("\n")
	}

	return mcp.NewToolResultText(fmt.Sprintf("Execution plan:\n%s", plan.String())), nil
}

func createIndexToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	schema, ok := request.Params.Arguments["schema"].(string)
	if !ok {
		return nil, errors.New("invalid schema parameter")
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, fmt.Errorf("创建索引失败: %v", err)
	}

	return mcp.NewToolResultText("Index created successfully"), nil
}

func describeTableToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	table_name, ok := request.Params.Arguments["table_name"].(string)
	if !ok {
		return nil, errors.New("invalid schema parameter")
	}

	query := fmt.Sprintf(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = '%s'
	`, table_name)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Describe table error: %v\n", err)
		return nil, fmt.Errorf("failed to describe table")
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var column string
		var dataType string
		if err := rows.Scan(&column, &dataType); err != nil {
			return nil, fmt.Errorf("error scanning table columns")
		}
		columns = append(columns, fmt.Sprintf("%s (%s)", column, dataType))
	}

	return mcp.NewToolResultText(fmt.Sprintf("Table columns: %v", columns)), nil
}

func parseSQLRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		pointers := make([]interface{}, len(cols))
		for i := range values {
			pointers[i] = &values[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}
	return results, nil
}

func sanitizeInput(input string) string {
	return strings.ReplaceAll(input, "'", "''")
}
