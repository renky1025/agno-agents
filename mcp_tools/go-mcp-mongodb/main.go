package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	mongoconn_db *mongo.Database
	mongoconn    *mongo.Client
	user         string
	password     string
	host         string
	port         string
	dbname       string
	auth         string
)

type MongoDBConfig struct {
	user     string
	password string
	host     string
	port     string
	dbname   string
	auth     string
}

func main() {
	// 硬编码MongoDB连接信息，避免通过命令行参数传递
	// mongouri = "mongodb://myusername:mypassword@10.100.2.1:27017/?authSource=admin"
	// dbname = "fastgpt"
	flag.StringVar(&user, "user", "", "MONGO username")
	flag.StringVar(&password, "password", "", "MONGO password")
	flag.StringVar(&host, "host", "", "MONGO host")
	flag.StringVar(&port, "port", "", "MONGO port")
	flag.StringVar(&dbname, "dbname", "", "MONGO DBNAME")
	flag.StringVar(&auth, "auth", "", "MONGO auth")
	flag.Parse()

	dbconfig := MongoDBConfig{
		user:     user,
		password: password,
		host:     host,
		port:     port,
		dbname:   dbname,
		auth:     auth,
	}

	if err := initConnectionPool(dbconfig); err != nil {
		os.Exit(1)
	}

	defer func() {
		if err := mongoconn.Disconnect(context.TODO()); err != nil {
			os.Exit(1)
		}
	}()

	s := server.NewMCPServer(
		"mongo-mcp-server 🚀",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	s.AddTool(createReadQueryTool(), readQueryToolHandler)
	s.AddTool(createWriteQueryTool(), writeQueryToolHandler)
	s.AddTool(createCreateTableTool(), createTableToolHandler)
	s.AddTool(createListTablesTool(), listTableToolHandler)
	s.AddTool(createExplainQueryTool(), explainQueryToolHandler)
	s.AddTool(createCreateIndexTool(), createIndexToolHandler)
	s.AddTool(createDescribeTableTool(), describeTableToolHandler)

	if err := server.ServeStdio(s); err != nil {
		os.Exit(1)
	}
}

func createReadQueryTool() mcp.Tool {
	return mcp.NewTool("mongo_read_query",
		mcp.WithDescription("Execute a SELECT query on the mongodb database"),
		mcp.WithString("collection_name",
			mcp.Required(),
			mcp.Description("Collection name to query"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("SELECT SQL query to execute"),
		),
	)
}

func createWriteQueryTool() mcp.Tool {
	return mcp.NewTool("mongo_write_query",
		mcp.WithDescription("Execute an INSERT, UPDATE, or DELETE query on the mongodb database"),
		mcp.WithString("collection_name",
			mcp.Required(),
			mcp.Description("Collection name to query"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("SQL query to execute"),
		),
	)
}

func createCreateTableTool() mcp.Tool {
	return mcp.NewTool("mongo_create_table",
		mcp.WithDescription("Create a new table in the mongodb database"),
		mcp.WithString("collection_name",
			mcp.Required(),
			mcp.Description("Collection name to create"),
		),
	)
}

func createListTablesTool() mcp.Tool {
	return mcp.NewTool("mongo_list_tables",
		mcp.WithDescription("List all user tables in the mongodb database"),
		mcp.WithString("schema",
			mcp.Description("Optional schema name to filter tables"),
		),
	)
}

// 创建explain查询工具定义
func createExplainQueryTool() mcp.Tool {
	return mcp.NewTool("mongo_explain_query",
		mcp.WithDescription("Explain a query execution plan on the mongodb database"),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("SQL query to explain, start with EXPLAIN"),
		),
	)
}

func createCreateIndexTool() mcp.Tool {
	return mcp.NewTool("mongo_create_index",
		mcp.WithDescription("Create a new index on the mongodb database"),
		mcp.WithString("collection_name",
			mcp.Required(),
			mcp.Description("Collection name to create index"),
		),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("CREATE INDEX SQL statement"),
		),
	)
}

func createDescribeTableTool() mcp.Tool {
	return mcp.NewTool("mongo_describe_table",
		mcp.WithDescription("Describe a table in the mongodb database, query the table structure in the mongodb database"),
		mcp.WithString("collection_name",
			mcp.Required(),
			mcp.Description("Table name to describe"),
		),
	)
}

func initConnectionPool(dbconfig MongoDBConfig) error {
	if dbconfig.user == "" || dbconfig.password == "" || dbconfig.host == "" || dbconfig.port == "" || dbconfig.dbname == "" {
		return errors.New("mongouri and dbname are required")
	}
	mongouri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=%s", dbconfig.user, dbconfig.password, dbconfig.host, dbconfig.port, dbconfig.auth)
	dbname = dbconfig.dbname
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongouri).SetMaxPoolSize(20)) // 连接池
	if err != nil {
		fmt.Println(err)
		return err
	}
	mongoconn = client
	mongoconn_db = mongoconn.Database(dbname)
	return nil
}

func readQueryToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	collection_name, ok := request.Params.Arguments["collection_name"].(string)
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("invalid query parameter")
	}

	// string to bson.M
	filter := bson.M{}
	if err := json.Unmarshal([]byte(query), &filter); err != nil {
		return nil, fmt.Errorf("invalid query parameter")
	}
	cursors, err := mongoconn_db.Collection(collection_name).Find(ctx, filter)
	if err != nil {
		// fmt.Printf("Query error: %v\n", err)
		return nil, fmt.Errorf("query execution failed")
	}
	results, err := parseSQLRows(cursors)
	if err != nil {
		return nil, fmt.Errorf("result parsing failed")
	}

	return mcp.NewToolResultText(fmt.Sprintf("Query results: %v", results)), nil
}

func writeQueryToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	collection_name, ok := request.Params.Arguments["collection_name"].(string)
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("invalid query parameter")
	}
	// string to bson.M
	filter := bson.M{}
	if err := json.Unmarshal([]byte(query), &filter); err != nil {
		return nil, fmt.Errorf("invalid query parameter")
	}
	result, err := mongoconn_db.Collection(collection_name).InsertOne(ctx, filter)
	if err != nil {
		// fmt.Printf("Write error: %v\n", err)
		return nil, fmt.Errorf("write operation failed")
	}

	return mcp.NewToolResultText(fmt.Sprintf("Operation successful. Rows affected: %d", result.InsertedID)), nil
}

func createTableToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	collection_name, ok := request.Params.Arguments["collection_name"].(string)
	if !ok {
		return nil, errors.New("invalid schema parameter")
	}
	// create collection
	if err := mongoconn_db.CreateCollection(ctx, collection_name); err != nil {
		// fmt.Printf("Create table error: %v\n", err)
		return nil, fmt.Errorf("table creation failed")
	}

	return mcp.NewToolResultText("Table created successfully"), nil
}

func listTableToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// list collections in mongodb
	collections, err := mongoconn_db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		// fmt.Printf("List tables error: %v\n", err)
		return nil, fmt.Errorf("failed to list tables")
	}

	var tables []string
	for _, collection := range collections {
		tables = append(tables, collection)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Tables: %v", tables)), nil
}

// explain查询处理函数
func explainQueryToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("invalid schema parameter")
	}
	ctx = context.Background()
	// explain query
	var result bson.M
	// Runs the command and prints the database statistics
	err := mongoconn_db.RunCommand(context.TODO(), bson.M{"explain": query}).Decode(&result)
	if err != nil {
		// fmt.Printf("Explain error: %v\n", err)
		return nil, fmt.Errorf("explain execution failed: %v", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Execution plan:\n%s", result)), nil
}

func createIndexToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	collection_name, ok := request.Params.Arguments["collection_name"].(string)
	schema, ok := request.Params.Arguments["schema"].(string)
	if !ok {
		return nil, errors.New("invalid schema parameter")
	}
	// string to bson.M
	filter := bson.M{}
	if err := json.Unmarshal([]byte(schema), &filter); err != nil {
		return nil, fmt.Errorf("无效的查询参数")
	}

	// 创建索引模型
	indexModel := mongo.IndexModel{
		Keys: filter,
	}

	if _, err := mongoconn_db.Collection(collection_name).Indexes().CreateOne(ctx, indexModel); err != nil {
		return nil, fmt.Errorf("创建索引失败: %v", err)
	}

	return mcp.NewToolResultText("索引创建成功"), nil
}

func describeTableToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	collection_name, ok := request.Params.Arguments["collection_name"].(string)
	if !ok {
		return nil, errors.New("invalid schema parameter")
	}

	row := mongoconn_db.Collection(collection_name).FindOne(ctx, bson.M{})
	var columns map[string]interface{}
	if err := row.Decode(&columns); err != nil {
		return nil, fmt.Errorf("failed to describe table")
	}

	return mcp.NewToolResultText(fmt.Sprintf("Table columns: %v", columns)), nil
}

func parseSQLRows(cursor *mongo.Cursor) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}
