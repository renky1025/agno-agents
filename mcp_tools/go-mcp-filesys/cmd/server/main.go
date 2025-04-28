package main

import (
	"flag"
	"go-mcp-filesys/internal/filesys"
	"go-mcp-filesys/internal/tools"
	"log"

	"github.com/mark3labs/mcp-go/server"
)

// 文件操作mcp服务
// 主要功能有：
// 1. 列出目录中的文件
// 2. 读取文件内容
// 3. 写入文件内容
// 4. 删除文件
// 5. 移动文件
// 6. 复制文件
// 7. 更改文件权限
// 8. 创建目录
// 9. 删除目录
// 10. 移动目录
// 11. 复制目录
// 12. 更改目录权限
// 13. 创建文件
// 14. 替换文件内容
// 15. 搜索文件
// 16. 统计文件
// 17. 编辑文件
// 18. 追加文件内容

func main() {
	// Parse command line arguments
	rootPath := flag.String("root", "", "Path to the root directory (uses default config if not specified)")
	transport := flag.String("transport", "stdio", "Transport to use (stdio, sse)")
	flag.Parse()

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"File System MCP Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)
	if *rootPath != "" {
		filesys.SetAllowedOpsFolder(*rootPath)
	}

	// Add basic tools
	//fmt.Println("Registering basic tools...")
	mcpServer.AddTool(tools.ListFilesInDirectoryTool(), tools.ListFilesInDirectoryHandle())
	mcpServer.AddTool(tools.ReplaceFileContentTool(), tools.ReplaceFileContentToolHandle())
	mcpServer.AddTool(tools.CreateNewFileTool(), tools.CreateNewFileToolHandle())
	mcpServer.AddTool(tools.AppendFileContentTool(), tools.AppendFileContentToolHandle())
	mcpServer.AddTool(tools.EditFileTool(), tools.EditFileToolHandle())
	mcpServer.AddTool(tools.FindFileTool(), tools.FindFileToolHandle())
	mcpServer.AddTool(tools.CountFilesInDirectoryTool(), tools.CountFilesInDirectoryToolHandle())
	mcpServer.AddTool(tools.DeleteFileTool(), tools.DeleteFileToolHandle())
	mcpServer.AddTool(tools.MoveFileTool(), tools.MoveFileToolHandle())
	mcpServer.AddTool(tools.CopyFileTool(), tools.CopyFileToolHandle())
	mcpServer.AddTool(tools.ChangeFilePermissionsTool(), tools.ChangeFilePermissionsToolHandle())
	mcpServer.AddTool(tools.CreateDirectoryTool(), tools.CreateDirectoryToolHandle())
	mcpServer.AddTool(tools.DeleteDirectoryTool(), tools.DeleteDirectoryToolHandle())
	mcpServer.AddTool(tools.MoveDirectoryTool(), tools.MoveDirectoryToolHandle())
	mcpServer.AddTool(tools.CopyDirectoryTool(), tools.CopyDirectoryToolHandle())
	mcpServer.AddTool(tools.ChangeDirectoryPermissionsTool(), tools.ChangeDirectoryPermissionsToolHandle())

	// Start stdio server
	// if err := server.ServeStdio(mcpServer); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	// 	os.Exit(1)
	// }
	// start sse server

	// mcpServer := server.NewMCPServer(
	// 	"File System MCP Server",
	// 	"1.0.0",
	// 	server.WithResourceCapabilities(true, true),
	// 	server.WithLogging(),
	// 	server.WithRecovery(),
	// )

	// Only check for "sse" since stdio is the default
	if *transport == "sse" {
		sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL("http://localhost:8080"))
		log.Printf("SSE server listening on :8080")
		if err := sseServer.Start(":8080"); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
