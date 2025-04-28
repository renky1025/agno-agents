package tools

import (
	"context"
	"fmt"
	"strings"

	"go-mcp-filesys/internal/filesys"

	"github.com/mark3labs/mcp-go/mcp"
)

// 创建一个工具，用于列出目录中的文件
func ListFilesInDirectoryTool() mcp.Tool {
	return mcp.NewTool("list_files_in_directory",
		mcp.WithDescription("List all files in the directory"),
		mcp.WithString("directory",
			mcp.Description("The directory to list files from"),
			mcp.DefaultString("."),
		),
		mcp.WithBoolean("includeSubdirectories",
			mcp.Description("Include subdirectories"),
			mcp.DefaultBool(true),
		),
	)
}

// --------------------------handle tools--------------------------------
func ListFilesInDirectoryHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		directory, _ := request.Params.Arguments["directory"].(string)
		includeSubdirectories, _ := request.Params.Arguments["includeSubdirectories"].(bool)
		absDirectory := filesys.GetAbsPathWithAllowedOpsFolder(directory)
		if absDirectory == "" {
			return nil, fmt.Errorf("%s directory is not allowed", directory)
		}
		files, err := filesys.ListFilesInDirectory(absDirectory, includeSubdirectories)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(strings.Join(files, "\n")), nil
	}
}

// 创建一个工具，用于读取文件内容
func ReadFileTool() mcp.Tool {
	return mcp.NewTool("read_file",
		mcp.WithDescription("Read the content of a file"),
		mcp.WithString("file",
			mcp.Description("The file to read"),
			mcp.DefaultString("."),
		),
	)
}

// --------------------------handle tools--------------------------------
func ReadFileToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, _ := request.Params.Arguments["file"].(string)
		absFile := filesys.GetAbsPathWithAllowedOpsFolder(file)
		if absFile == "" {
			return nil, fmt.Errorf("%s file is not allowed", file)
		}
		content, err := filesys.ReadFile(absFile)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(content), nil
	}
}

// 创建一个工具，用于写入文件内容
func WriteFileTool() mcp.Tool {
	return mcp.NewTool("write_file",
		mcp.WithDescription("Write the content to a file"),
		mcp.WithString("file",
			mcp.Description("The file to write to"),
			mcp.DefaultString("."),
		),
		mcp.WithString("content",
			mcp.Description("The content to write to the file"),
		),
	)
}

// --------------------------handle tools--------------------------------
func WriteFileToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, _ := request.Params.Arguments["file"].(string)
		content, _ := request.Params.Arguments["content"].(string)
		absFile := filesys.GetAbsPathWithAllowedOpsFolder(file)
		if absFile == "" {
			return nil, fmt.Errorf("%s file is not allowed", file)
		}
		err := filesys.WriteFile(absFile, content)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("write file success"), nil
	}
}

// 创建一个工具，用于删除文件
func DeleteFileTool() mcp.Tool {
	return mcp.NewTool("delete_file",
		mcp.WithDescription("Delete a file"),
		mcp.WithString("file",
			mcp.Description("The file to delete"),
			mcp.DefaultString("."),
		),
	)
}
func DeleteFileToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, _ := request.Params.Arguments["file"].(string)
		absFile := filesys.GetAbsPathWithAllowedOpsFolder(file)
		if absFile == "" {
			return nil, fmt.Errorf("%s file is not allowed", file)
		}
		err := filesys.DeleteFile(absFile)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("delete file success"), nil
	}
}

// 创建一个工具，用于移动文件
func MoveFileTool() mcp.Tool {
	return mcp.NewTool("move_file",
		mcp.WithDescription("Move a file"),
		mcp.WithString("file",
			mcp.Description("The file to move"),
			mcp.DefaultString("."),
		),
		mcp.WithString("destination",
			mcp.Description("The destination to move the file to"),
		),
	)
}
func MoveFileToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, _ := request.Params.Arguments["file"].(string)
		destination, _ := request.Params.Arguments["destination"].(string)
		absFile := filesys.GetAbsPathWithAllowedOpsFolder(file)
		absDestination := filesys.GetAbsPathWithAllowedOpsFolder(destination)
		if absFile == "" {
			return nil, fmt.Errorf("%s file is not allowed", file)
		}
		if absDestination == "" {
			return nil, fmt.Errorf("%s destination is not allowed", destination)
		}
		err := filesys.MoveFile(absFile, absDestination)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("move file success"), nil
	}
}

// 创建一个工具，用于复制文件
func CopyFileTool() mcp.Tool {
	return mcp.NewTool("copy_file",
		mcp.WithDescription("Copy a file"),
		mcp.WithString("file",
			mcp.Description("The file to copy"),
			mcp.DefaultString("."),
		),
		mcp.WithString("destination",
			mcp.Description("The destination to copy the file to"),
		),
	)
}
func CopyFileToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, _ := request.Params.Arguments["file"].(string)
		destination, _ := request.Params.Arguments["destination"].(string)
		absFile := filesys.GetAbsPathWithAllowedOpsFolder(file)
		absDestination := filesys.GetAbsPathWithAllowedOpsFolder(destination)
		if absFile == "" {
			return nil, fmt.Errorf("%s file is not allowed", file)
		}
		if absDestination == "" {
			return nil, fmt.Errorf("%s destination is not allowed", destination)
		}
		err := filesys.CopyFile(absFile, absDestination)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("copy file success"), nil
	}
}

// 创建一个工具，用于更改文件权限
func ChangeFilePermissionsTool() mcp.Tool {
	return mcp.NewTool("change_file_permissions",
		mcp.WithDescription("Change the permissions of a file"),
		mcp.WithString("file",
			mcp.Description("The file to change the permissions of"),
			mcp.DefaultString("."),
		),
		mcp.WithString("permissions",
			mcp.Description("The permissions to change the file to"),
		),
	)
}
func ChangeFilePermissionsToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, _ := request.Params.Arguments["file"].(string)
		permissions, _ := request.Params.Arguments["permissions"].(string)
		absFile := filesys.GetAbsPathWithAllowedOpsFolder(file)
		if absFile == "" {
			return nil, fmt.Errorf("%s file is not allowed", file)
		}
		err := filesys.ChangeFilePermissions(absFile, permissions)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("change file permissions success"), nil
	}
}

// 创建一个工具，用于创建目录
func CreateDirectoryTool() mcp.Tool {
	return mcp.NewTool("create_directory",
		mcp.WithDescription("Create a directory"),
		mcp.WithString("directory",
			mcp.Description("The directory to create"),
			mcp.DefaultString("."),
		),
	)
}
func CreateDirectoryToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		directory, _ := request.Params.Arguments["directory"].(string)
		//检查文件名是否合法
		if !filesys.IsValidFileName(directory) {
			return nil, fmt.Errorf("%s directory name is not allowed", directory)
		}
		absDirectory := filesys.GetAbsPathWithAllowedOpsFolder(directory)
		if absDirectory != "" {
			return nil, fmt.Errorf("%s directory already exists", directory)
		}
		newDirectory := fmt.Sprintf("%s/%s", filesys.ALLOWED_OPS_FOLDER, directory)
		err := filesys.CreateDirectory(newDirectory)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("create directory success"), nil
	}
}

// 创建一个工具，用于删除目录
func DeleteDirectoryTool() mcp.Tool {
	return mcp.NewTool("delete_directory",
		mcp.WithDescription("Delete a directory"),
		mcp.WithString("directory",
			mcp.Description("The directory to delete"),
			mcp.DefaultString("."),
		),
	)
}
func DeleteDirectoryToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		directory, _ := request.Params.Arguments["directory"].(string)
		absDirectory := filesys.GetAbsPathWithAllowedOpsFolder(directory)
		if absDirectory == "" {
			return nil, fmt.Errorf("%s directory is not allowed", directory)
		}
		err := filesys.DeleteDirectory(absDirectory)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("delete directory success"), nil
	}
}

// 创建一个工具，用于移动目录
func MoveDirectoryTool() mcp.Tool {
	return mcp.NewTool("move_directory",
		mcp.WithDescription("Move a directory"),
		mcp.WithString("directory",
			mcp.Description("The directory to move"),
			mcp.DefaultString("."),
		),
		mcp.WithString("destination",
			mcp.Description("The destination to move the directory to"),
		),
	)
}
func MoveDirectoryToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		directory, _ := request.Params.Arguments["directory"].(string)
		destination, _ := request.Params.Arguments["destination"].(string)
		absDirectory := filesys.GetAbsPathWithAllowedOpsFolder(directory)
		absDestination := filesys.GetAbsPathWithAllowedOpsFolder(destination)
		if absDirectory == "" {
			return nil, fmt.Errorf("%s directory is not allowed", directory)
		}
		if absDestination == "" {
			return nil, fmt.Errorf("%s destination is not allowed", destination)
		}
		err := filesys.MoveDirectory(absDirectory, absDestination)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("move directory success"), nil
	}
}

// 创建一个工具，用于复制目录
func CopyDirectoryTool() mcp.Tool {
	return mcp.NewTool("copy_directory",
		mcp.WithDescription("Copy a directory"),
		mcp.WithString("directory",
			mcp.Description("The directory to copy"),
			mcp.DefaultString("."),
		),
		mcp.WithString("destination",
			mcp.Description("The destination to copy the directory to"),
		),
	)
}
func CopyDirectoryToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		directory, _ := request.Params.Arguments["directory"].(string)
		destination, _ := request.Params.Arguments["destination"].(string)
		absDirectory := filesys.GetAbsPathWithAllowedOpsFolder(directory)
		absDestination := filesys.GetAbsPathWithAllowedOpsFolder(destination)
		if absDirectory == "" {
			return nil, fmt.Errorf("%s directory is not allowed", directory)
		}
		if absDestination == "" {
			// 如果目标目录不存在，则创建目标目录
			newfolder := fmt.Sprintf("%s/%s", filesys.ALLOWED_OPS_FOLDER, destination)
			filesys.CreateDirectory(newfolder)
			absDestination = newfolder
		}
		err := filesys.CopyDirectory(absDirectory, absDestination)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("copy directory success"), nil
	}
}

// 创建一个工具，用于统计目录中的文件数量和大小
func CountFilesInDirectoryTool() mcp.Tool {
	return mcp.NewTool("count_files_in_directory",
		mcp.WithDescription("Count the number of files and their total size in a directory"),
		mcp.WithString("directory",
			mcp.Description("The directory to count the files in"),
			mcp.DefaultString("."),
		),
	)
}
func CountFilesInDirectoryToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		directory, _ := request.Params.Arguments["directory"].(string)
		absDirectory := filesys.GetAbsPathWithAllowedOpsFolder(directory)
		if absDirectory == "" {
			return nil, fmt.Errorf("%s directory is not allowed", directory)
		}
		num, size, err := filesys.CountFilesInDirectory(absDirectory)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("number of files: %d, total size: %d", num, size)), nil
	}
}

// 创建一个工具，用于查找文件，文件内容查找，并返回文件路径
func FindFileTool() mcp.Tool {
	return mcp.NewTool("find_file",
		mcp.WithDescription("Find a file, file content search, and return the file path"),
		mcp.WithString("file",
			mcp.Description("The file to find"),
			mcp.DefaultString("."),
		),
		mcp.WithString("content",
			mcp.Description("The content to find in the file"),
		),
	)
}

// --------------------------handle tools--------------------------------
func FindFileToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		directory := filesys.GetAbsPathWithAllowedOpsFolder("")
		content, _ := request.Params.Arguments["content"].(string)
		filename, _ := request.Params.Arguments["file"].(string)
		if len(filename) > 0 {
			// search file in directory
			files, err := filesys.SearchFile(directory, filename)
			if err != nil {
				return nil, err
			}
			return mcp.NewToolResultText(strings.Join(files, "\n")), nil
		}
		if len(content) > 0 {
			filepaths, err := filesys.SearchFileContent(directory, content)
			if err != nil {
				return nil, err
			}
			return mcp.NewToolResultText(strings.Join(filepaths, "\n")), nil
		}
		return nil, fmt.Errorf("no file or content provided")
	}
}

// 创建一个工具，用于替换文件内容
func ReplaceFileContentTool() mcp.Tool {
	return mcp.NewTool("replace_file_content",
		mcp.WithDescription("Replace the content of a file"),
		mcp.WithString("file",
			mcp.Description("The file to replace the content of"),
			mcp.DefaultString("."),
		),
		mcp.WithString("content",
			mcp.Description("The content to replace the content of the file with"),
		),
	)
}

// --------------------------handle tools--------------------------------
func ReplaceFileContentToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, _ := request.Params.Arguments["file"].(string)
		content, _ := request.Params.Arguments["content"].(string)
		newcontent, _ := request.Params.Arguments["newcontent"].(string)
		absFile := filesys.GetAbsPathWithAllowedOpsFolder(file)
		if absFile == "" {
			return nil, fmt.Errorf("%s file is not allowed", file)
		}
		err := filesys.ReplaceFileContent(absFile, content, newcontent)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("replace file content success"), nil
	}
}

// 创建一个工具，用于创建新文件
func CreateNewFileTool() mcp.Tool {
	return mcp.NewTool("create_new_file",
		mcp.WithDescription("Create a new file"),
		mcp.WithString("file",
			mcp.Description("The file to create"),
			mcp.DefaultString("."),
		),
		mcp.WithString("content",
			mcp.Description("The content to write to the file"),
		),
	)
}

// --------------------------handle tools--------------------------------
func CreateNewFileToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, _ := request.Params.Arguments["file"].(string)
		content, _ := request.Params.Arguments["content"].(string)
		//检查文件名是否合法
		splitarr := strings.Split(file, "/")
		if len(splitarr) > 1 {
			if !filesys.IsValidFileName(splitarr[len(splitarr)-1]) {
				return nil, fmt.Errorf("%s file name is not allowed", splitarr[len(splitarr)-1])
			}
		} else {
			if !filesys.IsValidFileName(file) {
				return nil, fmt.Errorf("%s file name is not allowed", file)
			}
		}
		absFile := filesys.GetAbsPathWithAllowedOpsFolder(file)
		if absFile != "" {
			return nil, fmt.Errorf("%s file already exists, please use another name", file)
		}
		absFile = fmt.Sprintf("%s/%s", filesys.ALLOWED_OPS_FOLDER, file)
		err := filesys.CreateNewFile(absFile, content)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText(fmt.Sprintf("create new file success %s", file)), nil
	}
}

// 创建一个工具，用于追加文件内容
func AppendFileContentTool() mcp.Tool {
	return mcp.NewTool("append_file_content",
		mcp.WithDescription("Append content to a file"),
		mcp.WithString("file",
			mcp.Description("The file to append the content to"),
			mcp.DefaultString("."),
		),
		mcp.WithString("content",
			mcp.Description("The content to append to the file"),
		),
	)
}

// --------------------------handle tools--------------------------------
func AppendFileContentToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, _ := request.Params.Arguments["file"].(string)
		content, _ := request.Params.Arguments["content"].(string)
		absFile := filesys.GetAbsPathWithAllowedOpsFolder(file)
		// 多层级目录的文件
		splitarr := strings.Split(file, "/")
		if len(splitarr) > 1 {
			absFile = fmt.Sprintf("%s/%s", filesys.ALLOWED_OPS_FOLDER, file)
		} else {
			absFile = fmt.Sprintf("%s/%s", filesys.ALLOWED_OPS_FOLDER, file)
		}

		err := filesys.AppendFileContent(absFile, content)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("append file content success"), nil
	}
}

// 创建一个工具，用于编辑文件
func EditFileTool() mcp.Tool {
	return mcp.NewTool("edit_file",
		mcp.WithDescription("Edit a file"),
		mcp.WithString("file",
			mcp.Description("The file to edit"),
			mcp.DefaultString("."),
		),
		mcp.WithString("content",
			mcp.Description("The content to edit the file with"),
		),
	)
}

// --------------------------handle tools--------------------------------
func EditFileToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file, _ := request.Params.Arguments["file"].(string)
		content, _ := request.Params.Arguments["content"].(string)
		absFile := filesys.GetAbsPathWithAllowedOpsFolder(file)
		if absFile == "" {
			return nil, fmt.Errorf("%s file is not allowed", file)
		}
		err := filesys.EditFile(absFile, content)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("edit file success"), nil
	}
}

// 创建一个工具，用于修改文件夹权限
func ChangeDirectoryPermissionsTool() mcp.Tool {
	return mcp.NewTool("change_directory_permissions",
		mcp.WithDescription("Change the permissions of a directory"),
		mcp.WithString("directory",
			mcp.Description("The directory to change the permissions of"),
		),
		mcp.WithString("permissions",
			mcp.Description("The permissions to change the directory to"),
		),
	)
}

// --------------------------handle tools--------------------------------
func ChangeDirectoryPermissionsToolHandle() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		directory, _ := request.Params.Arguments["directory"].(string)
		permissions, _ := request.Params.Arguments["permissions"].(string)
		absDirectory := filesys.GetAbsPathWithAllowedOpsFolder(directory)
		if absDirectory == "" {
			return nil, fmt.Errorf("%s directory is not allowed", directory)
		}
		err := filesys.ChangeDirectoryPermissions(absDirectory, permissions)
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("change directory permissions success"), nil
	}
}
