package filesys

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// 允许编辑的文件夹
var (
	ALLOWED_OPS_FOLDER = "C:/workspaces/python-projects/python-diff-img/output"
	folderMutex        sync.RWMutex
)

// 设置允许编辑的文件夹
func SetAllowedOpsFolder(folder string) {
	folderMutex.Lock()
	defer folderMutex.Unlock()

	// 规范化路径
	folder = filepath.Clean(folder)
	ALLOWED_OPS_FOLDER = folder
}

// 检查文件名是否符合系统命名规则
func IsValidFileName(fileName string) bool {
	// 检查文件名是否为空
	if fileName == "" {
		return false
	}

	// 检查文件名是否包含非法字符
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "..", "~"}
	for _, char := range invalidChars {
		if strings.Contains(fileName, char) {
			return false
		}
	}

	// 检查文件名是否以点开头
	if strings.HasPrefix(fileName, ".") {
		return false
	}

	// 检查文件名长度
	if len(fileName) > 255 {
		return false
	}

	return true
}

// 检查路径是否在允许的目录内
func isPathInAllowedDirectory(path string) bool {
	folderMutex.RLock()
	defer folderMutex.RUnlock()

	cleanPath := filepath.Clean(path)
	cleanAllowed := filepath.Clean(ALLOWED_OPS_FOLDER)

	rel, err := filepath.Rel(cleanAllowed, cleanPath)
	if err != nil {
		return false
	}

	return !strings.Contains(rel, "..")
}

// 检查文件是否为普通文件
func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// 安全的文件写入
func safeWriteFile(path string, content []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, "tmp_*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

// 转义文件名的特殊字符，将Windows路径转换为Unix格式
// 示例: C:\workspaces\python-projects\python-diff-img\output\hello world
// ==> C:/workspaces/python-projects/python-diff-img/output/hello world
func EscapeFileName(fileName string) string {
	// 规范化路径分隔符
	fileName = filepath.ToSlash(filepath.Clean(fileName))
	return fileName
}

// 获取指定文件在ALLOWED_OPS_FOLDER中的绝对路径
func GetAbsPathWithAllowedOpsFolder(targetfile string) string {
	folderMutex.RLock()
	defer folderMutex.RUnlock()

	if targetfile == "" {
		return ALLOWED_OPS_FOLDER
	}

	// 规范化路径
	fullPath := filepath.Join(ALLOWED_OPS_FOLDER, targetfile)
	fullPath = filepath.Clean(fullPath)

	// 检查路径是否在允许的目录内
	if !isPathInAllowedDirectory(fullPath) {
		return ""
	}

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); err != nil {
		return ""
	}

	return fullPath
}

// 检查指定目录是否在ALLOWED_OPS_FOLDER中
func IsAllowedOpsFolder(folder string) bool {
	// 检查folder是否是ALLOWED_OPS_FOLDER的子目录
	if strings.HasPrefix(folder, ALLOWED_OPS_FOLDER) {
		return true
	}
	return false
}

// 列出目录中的文件
func ListFilesInDirectory(directory string, includeSubdirectories bool) ([]string, error) {
	// 检查目录是否在允许的范围内
	if !isPathInAllowedDirectory(directory) {
		return nil, fmt.Errorf("directory access not allowed: %s", directory)
	}

	directory = filepath.Clean(directory)
	result := make([]string, 0)

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径
		relPath, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}

		// 如果不包含子目录且不是根目录，则跳过
		if !includeSubdirectories && filepath.Dir(relPath) != "." && info.IsDir() {
			return filepath.SkipDir
		}

		result = append(result, filepath.ToSlash(relPath))
		return nil
	}

	if err := filepath.Walk(directory, walkFn); err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return result, nil
}

// 读取文件内容
func ReadFile(filePath string) (string, error) {
	if !isPathInAllowedDirectory(filePath) {
		return "", fmt.Errorf("access denied: %s", filePath)
	}

	cleanPath := filepath.Clean(filePath)
	if !isRegularFile(cleanPath) {
		return "", fmt.Errorf("not a regular file: %s", filePath)
	}

	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// 写入文件内容
func WriteFile(filePath string, content string) error {
	if !isPathInAllowedDirectory(filePath) {
		return fmt.Errorf("access denied: %s", filePath)
	}

	cleanPath := filepath.Clean(filePath)
	if !isRegularFile(cleanPath) {
		return fmt.Errorf("not a regular file: %s", filePath)
	}

	return safeWriteFile(cleanPath, []byte(content), 0644)
}

// 删除文件
func DeleteFile(filePath string) error {
	if !isPathInAllowedDirectory(filePath) {
		return fmt.Errorf("access denied: %s", filePath)
	}

	cleanPath := filepath.Clean(filePath)
	if !isRegularFile(cleanPath) {
		return fmt.Errorf("not a regular file: %s", filePath)
	}

	if err := os.Remove(cleanPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// 移动文件
func MoveFile(oldPath, newPath string) error {
	if !isPathInAllowedDirectory(oldPath) || !isPathInAllowedDirectory(newPath) {
		return fmt.Errorf("access denied: source or destination path not allowed")
	}

	cleanOldPath := filepath.Clean(oldPath)
	cleanNewPath := filepath.Clean(newPath)

	if !isRegularFile(cleanOldPath) {
		return fmt.Errorf("not a regular file: %s", oldPath)
	}

	// 使用临时文件进行移动操作
	content, err := os.ReadFile(cleanOldPath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := safeWriteFile(cleanNewPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	if err := os.Remove(cleanOldPath); err != nil {
		// 如果删除源文件失败，尝试删除目标文件
		os.Remove(cleanNewPath)
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}

// 复制文件
func CopyFile(oldPath string, newPath string) error {
	if !isPathInAllowedDirectory(oldPath) || !isPathInAllowedDirectory(newPath) {
		return fmt.Errorf("access denied: source or destination path not allowed")
	}

	cleanOldPath := filepath.Clean(oldPath)
	cleanNewPath := filepath.Clean(newPath)

	if !isRegularFile(cleanOldPath) {
		return fmt.Errorf("not a regular file: %s", oldPath)
	}

	// 读取源文件
	content, err := os.ReadFile(cleanOldPath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// 使用安全写入方式复制到新文件
	return safeWriteFile(cleanNewPath, content, 0644)
}

// 修改文件权限
func ChangeFilePermissions(filePath string, permissions string) error {
	if !isPathInAllowedDirectory(filePath) {
		return fmt.Errorf("access denied: %s", filePath)
	}

	cleanPath := filepath.Clean(filePath)
	if !isRegularFile(cleanPath) {
		return fmt.Errorf("not a regular file: %s", filePath)
	}

	perm, err := strconv.ParseUint(permissions, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid permissions format: %s", permissions)
	}

	// 限制权限范围
	if perm > 0777 {
		return fmt.Errorf("invalid permissions value: %s", permissions)
	}

	if err := os.Chmod(cleanPath, os.FileMode(perm)); err != nil {
		return fmt.Errorf("failed to change permissions: %w", err)
	}

	return nil
}

// 创建目录
func CreateDirectory(directory string) error {
	if !isPathInAllowedDirectory(directory) {
		return fmt.Errorf("access denied: %s", directory)
	}

	cleanPath := filepath.Clean(directory)
	if err := os.MkdirAll(cleanPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// 删除目录
func DeleteDirectory(directory string) error {
	if !isPathInAllowedDirectory(directory) {
		return fmt.Errorf("access denied: %s", directory)
	}

	cleanPath := filepath.Clean(directory)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to access directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", directory)
	}

	if err := os.RemoveAll(cleanPath); err != nil {
		return fmt.Errorf("failed to delete directory: %w", err)
	}

	return nil
}

// 移动目录
func MoveDirectory(oldPath string, newPath string) error {
	// 检查目录是否存在
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", oldPath)
	}
	// 移动目录
	oldPath = EscapeFileName(oldPath)
	newPath = EscapeFileName(newPath)
	return os.Rename(oldPath, newPath)
}

// 复制目录
func CopyDirectory(oldPath string, newPath string) error {
	if !isPathInAllowedDirectory(oldPath) || !isPathInAllowedDirectory(newPath) {
		return fmt.Errorf("access denied: source or destination path not allowed")
	}

	cleanOldPath := filepath.Clean(oldPath)
	cleanNewPath := filepath.Clean(newPath)

	// 检查源目录
	srcInfo, err := os.Stat(cleanOldPath)
	if err != nil {
		return fmt.Errorf("failed to access source directory: %w", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", oldPath)
	}

	// 创建目标目录
	if err := os.MkdirAll(cleanNewPath, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 遍历源目录
	entries, err := os.ReadDir(cleanOldPath)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(cleanOldPath, entry.Name())
		dstPath := filepath.Join(cleanNewPath, entry.Name())

		if entry.IsDir() {
			if err := CopyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// 统计目录中的文件数量和文件大小
func CountFilesInDirectory(directory string) (int, int64, error) {
	// 检查目录是否存在
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return 0, 0, fmt.Errorf("directory does not exist: %s", directory)
	}
	// 统计目录中的文件数量和文件大小
	directory = EscapeFileName(directory)
	files, err := os.ReadDir(directory)
	if err != nil {
		return 0, 0, err
	}
	totalSize := int64(0)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
	}
	return len(files), totalSize, nil
}

// 搜索文件
func SearchFile(directory string, fileName string) ([]string, error) {
	// 检查目录是否存在
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil, fmt.Errorf("directory does not exist: %s", directory)
	}
	// 搜索文件
	result := make([]string, 0)
	directory = EscapeFileName(directory)
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(info.Name(), fileName) {
			// 获取相对路径
			relativePath, err := filepath.Rel(directory, path)
			if err != nil {
				return err
			}
			result = append(result, relativePath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 搜索文件内容
func SearchFileContent(directory string, content string) ([]string, error) {
	if !isPathInAllowedDirectory(directory) {
		return nil, fmt.Errorf("access denied: %s", directory)
	}

	cleanDir := filepath.Clean(directory)
	result := make([]string, 0)

	err := filepath.Walk(cleanDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过非常规文件
		if !info.Mode().IsRegular() {
			return nil
		}

		// 检查路径是否在允许的目录内
		if !isPathInAllowedDirectory(path) {
			return nil
		}

		// 读取文件内容
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if strings.Contains(string(fileContent), content) {
			relPath, err := filepath.Rel(cleanDir, path)
			if err != nil {
				return err
			}
			result = append(result, filepath.ToSlash(relPath))
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	return result, nil
}

// 替换文件内容
func ReplaceFileContent(filePath string, oldContent string, newContent string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}
	// 替换文件内容
	filePath = EscapeFileName(filePath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	// 替换文件内容
	tmpContent := strings.Replace(string(content), oldContent, newContent, -1)
	return os.WriteFile(filePath, []byte(tmpContent), 0644)
}

// 创建新文件
func CreateNewFile(tmppath string, content string) error {
	if !isPathInAllowedDirectory(tmppath) {
		return fmt.Errorf("access denied: %s", tmppath)
	}

	cleanPath := filepath.Clean(tmppath)

	// 检查文件名是否合法
	if !IsValidFileName(filepath.Base(cleanPath)) {
		return fmt.Errorf("invalid file name: %s", filepath.Base(cleanPath))
	}

	// 检查文件是否已存在
	if _, err := os.Stat(cleanPath); err == nil {
		return fmt.Errorf("file already exists: %s", tmppath)
	}

	// 确保目录存在
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 使用安全写入方式创建文件
	return safeWriteFile(cleanPath, []byte(content), 0644)
}

// 追加文件内容
func AppendFileContent(filePath string, content string) error {
	if !isPathInAllowedDirectory(filePath) {
		return fmt.Errorf("access denied: %s", filePath)
	}

	cleanPath := filepath.Clean(filePath)
	if !isRegularFile(cleanPath) {
		return fmt.Errorf("not a regular file: %s", filePath)
	}

	// 读取现有内容
	oldContent, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 准备新内容
	newContent := append(oldContent, []byte("\n"+content)...)

	// 使用安全写入方式更新文件
	return safeWriteFile(cleanPath, newContent, 0644)
}

// 编辑文件
func EditFile(filePath string, content string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}
	// 编辑文件
	return os.WriteFile(filePath, []byte(content), 0644)
}

// 修改目录权限
func ChangeDirectoryPermissions(directory string, permissions string) error {
	if !isPathInAllowedDirectory(directory) {
		return fmt.Errorf("access denied: %s", directory)
	}

	cleanPath := filepath.Clean(directory)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to access directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", directory)
	}

	perm, err := strconv.ParseUint(permissions, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid permissions format: %s", permissions)
	}

	// 限制权限范围
	if perm > 0777 {
		return fmt.Errorf("invalid permissions value: %s", permissions)
	}

	if err := os.Chmod(cleanPath, os.FileMode(perm)); err != nil {
		return fmt.Errorf("failed to change permissions: %w", err)
	}

	return nil
}
