from mcp.server.fastmcp import FastMCP
import datetime
import os
import shutil
import logging
from pathlib import Path
from typing import List, Union

# 设置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger("mcp_tools")

# 项目根目录作为安全边界
PROJECT_ROOT = Path(os.getcwd()).resolve()

mcp = FastMCP(name="File System MCP Server🚀")

# 总结prompt文本
@mcp.prompt("summarize")
async def summarize_prompt(text: str) -> list[dict]:
    """Generates a prompt to summarize the provided text."""
    return [
        {"role": "system", "content": "You are a helpful assistant skilled at summarization."},
        {"role": "user", "content": f"Please summarize the following text:\n\n{text}"}
    ]

# 添加一个动态问候资源
@mcp.resource("greeting://{name}")
def get_greeting(name: str) -> str:
    """获取个性化问候语"""
    return f"Hello, {name}!"

@mcp.tool(name="get_current_time", description="获取当前时间（ISO格式）")
def get_current_time() -> str:
    """获取当前时间（ISO格式）"""
    return datetime.datetime.now().isoformat()

@mcp.resource("files_in_directory://{directory}")
def list_files_resource(directory: str) -> list[str]:
    """列出指定目录中的文件"""
    try:
        return _safe_list_files(directory)
    except Exception as e:
        logger.error(f"列出目录内容失败: {str(e)}")
        return [f"错误: {str(e)}"]

### 文件操作策略
# Security First: 优先防止访问项目根目录外的文件
# Efficiency: 最小化AI交互的通信开销和Token使用
# Robustness: 提供批量操作的详细结果和错误报告
# Simplicity: 通过MCP提供清晰一致的API
# Standard Compliance: 严格遵守模型上下文协议

def _is_path_safe(path: Union[str, Path]) -> bool:
    """检查路径是否在项目根目录内（安全检查）"""
    try:
        path = Path(path).resolve()
        return PROJECT_ROOT in path.parents or path == PROJECT_ROOT
    except Exception:
        return False

def _safe_path(path: Union[str, Path]) -> Path:
    """确保路径安全并返回Path对象"""
    path = Path(path).resolve()
    if not _is_path_safe(path):
        raise ValueError(f"安全错误：路径 {path} 在项目根目录之外")
    return path
# 统一格式化mcp输出结果
def format_mcp_output(results: List[str]) -> List[str]:
    """格式化MCP输出结果"""
    return [f"成功: {result}" if result.startswith("成功") else f"失败: {result}" for result in results]

### Explore & Inspect: 列出文件/目录、获取详细状态
@mcp.tool(name="stat_items", description="获取多个项目的详细状态")
def stat_items(items: List[str]) -> List[str]:
    """获取多个项目的详细状态"""
    results = []
    for item in items:
        try:
            item_path = _safe_path(item)
            if not item_path.exists():
                results.append(f"文件不存在: {item}")
                continue
                
            stats = item_path.stat()
            modified_time = datetime.datetime.fromtimestamp(stats.st_mtime).isoformat()
            results.append(f"项目: {item}, 大小: {stats.st_size} 字节, "
                          f"修改时间: {modified_time}, "
                          f"类型: {'目录' if item_path.is_dir() else '文件'}")
        except Exception as e:
            results.append(f"获取 {item} 状态时出错: {str(e)}")
    
    return results

def _safe_list_files(directory: str) -> List[str]:
    """安全地列出指定目录中的文件（内部函数）"""
    dir_path = _safe_path(directory)
    if not dir_path.exists():
        raise FileNotFoundError(f"目录不存在: {directory}")
    if not dir_path.is_dir():
        raise NotADirectoryError(f"不是一个目录: {directory}")
    
    return [str(f.name) for f in dir_path.iterdir() if f.is_file()]

@mcp.tool(name="list_files", description="列出指定目录中的文件")
def list_files(directory: str) -> List[str]:
    """列出指定目录中的文件"""
    try:
        return _safe_list_files(directory)
    except Exception as e:
        logger.error(f"列出文件失败: {str(e)}")
        return [f"错误: {str(e)}"]

@mcp.tool(name="list_all", description="列出目录中的所有文件和子目录")
def list_all(directory: str, include_dirs: bool = True) -> List[str]:
    """列出目录中的所有文件和子目录"""
    try:
        dir_path = _safe_path(directory)
        if not dir_path.exists():
            raise FileNotFoundError(f"目录不存在: {directory}")
        if not dir_path.is_dir():
            raise NotADirectoryError(f"不是一个目录: {directory}")
        
        items = []
        for item in dir_path.iterdir():
            if include_dirs or item.is_file():
                items.append(f"{'[DIR]' if item.is_dir() else '[FILE]'} {item.name}")
        return items
    except Exception as e:
        logger.error(f"列出所有项目失败: {str(e)}")
        return [f"错误: {str(e)}"]

###  Read & Write Content: 读/写/追加多个文件，创建父目录
@mcp.tool(name="read_content", description="读取多个文件的内容")
def read_content(file: str) -> str:
    """读取多个文件的内容"""
    results = []
    try:
        file_path = _safe_path(file)
        if not file_path.exists():
            results.append(f"文件不存在: {file}")
        if not file_path.is_file():
            results.append(f"不是一个文件: {file}")
        with open(file_path, "r", encoding="utf-8") as f:
            results.append(f.read())
    except Exception as e:
        results.append(f"读取 {file} 时出错: {str(e)}")

    return results

@mcp.tool(name="write_content", description="将内容写入多个文件")
def write_content(file: str, content: str) -> str:
    """将内容写入多个文件"""
    results = []
    try:
        file_path = _safe_path(file)
        # 确保父目录存在
        file_path.parent.mkdir(parents=True, exist_ok=True)
        with open(file_path, "w", encoding="utf-8") as f:
            f.write(content)
        results.append(f"成功写入: {file}")
    except Exception as e:
        results.append(f"写入 {file} 时出错: {str(e)}")

    return results

@mcp.tool(name="append_content", description="将内容追加到多个文件中")
def append_content(file: str, content: str) -> str:
    """将内容追加到多个文件中"""
    results = []
    try:
        file_path = _safe_path(file)
        # 确保父目录存在
        file_path.parent.mkdir(parents=True, exist_ok=True)
            
        with open(file_path, "a", encoding="utf-8") as f:
            f.write(content)
        results.append(f"成功追加: {file}")
    except Exception as e:
        results.append(f"追加到 {file} 时出错: {str(e)}")
    
    return results

### Precision Editing & Searching: 精确编辑和搜索
@mcp.tool(name="edit_file", description="编辑文件的内容")
def edit_file(file: str, content: str) -> str:
    """编辑文件的内容"""
    return write_content(file, content)  # 复用写入功能

@mcp.tool(name="search_file", description="搜索包含查询内容的文件")
def search_file(file: str, query: str) -> str:
    """搜索包含查询内容的文件"""
    results = []
    try:
        file_path = _safe_path(file)
        if not file_path.exists():
            results.append(f"文件不存在: {file}")
        if not file_path.is_file():
            results.append(f"不是一个文件: {file}")
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()
            if query in content:
                results.append(file)
    except Exception as e:
        logger.error(f"搜索 {file} 时出错: {str(e)}")

    return results

@mcp.tool(name="replace_content", description="替换文件中的内容")
def replace_content(file: str, query: str, replacement: str) -> str:
    """替换文件中的内容"""
    results = []
    try:
        file_path = _safe_path(file)
        if not file_path.exists():
            results.append(f"文件不存在: {file}")
        if not file_path.is_file():
            results.append(f"不是一个文件: {file}")
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()
        # 执行替换
        new_content = content.replace(query, replacement)
        # 写回文件
        with open(file_path, "w", encoding="utf-8") as f:
            f.write(new_content)
        results.append(f"成功替换 {file} 中的内容")
    except Exception as e:
        results.append(f"替换 {file} 中的内容时出错: {str(e)}")

    return results

### 目录管理
@mcp.tool(name="create_directories", description="创建目录（包括中间父目录）")
def create_directories(directory: str) -> str:
    """创建目录（包括中间父目录）"""
    results = []
    try:
        dir_path = _safe_path(directory)
        dir_path.mkdir(parents=True, exist_ok=True)
        results.append(f"成功创建目录: {directory}")
    except Exception as e:
        results.append(f"创建目录 {directory} 时出错: {str(e)}")

    return results

### 安全删除
@mcp.tool(name="delete_items", description="递归删除多个文件/目录")
def delete_items(items: List[str]) -> List[str]:
    """递归删除多个文件/目录"""
    results = []
    for item in items:
        try:
            item_path = _safe_path(item)
            if not item_path.exists():
                results.append(f"项目不存在: {item}")
                continue
                
            if item_path.is_file():
                item_path.unlink()
                results.append(f"成功删除文件: {item}")
            elif item_path.is_dir():
                shutil.rmtree(item_path)
                results.append(f"成功删除目录: {item}")
        except Exception as e:
            results.append(f"删除 {item} 时出错: {str(e)}")
    
    return results

### 移动和复制
@mcp.tool(name="move_items", description="移动/重命名多个文件/目录")
def move_items(items: List[str], destination: str) -> List[str]:
    """移动/重命名多个文件/目录"""
    results = []
    dest_path = _safe_path(destination)
    
    # 确保目标目录存在
    if not dest_path.is_dir() and len(items) > 1:
        try:
            dest_path.mkdir(parents=True, exist_ok=True)
        except Exception as e:
            return [f"创建目标目录失败: {str(e)}"]
    
    for item in items:
        try:
            item_path = _safe_path(item)
            if not item_path.exists():
                results.append(f"项目不存在: {item}")
                continue
                
            if dest_path.is_dir():
                target = dest_path / item_path.name
            else:
                target = dest_path
                
            # 执行移动
            shutil.move(str(item_path), str(target))
            results.append(f"成功移动: {item} -> {target}")
        except Exception as e:
            results.append(f"移动 {item} 时出错: {str(e)}")
    
    return results

@mcp.tool(name="copy_items", description="复制多个文件/目录")
def copy_items(items: List[str], destination: str) -> List[str]:
    """复制多个文件/目录"""
    results = []
    dest_path = _safe_path(destination)
    
    # 确保目标目录存在
    if not dest_path.is_dir() and len(items) > 1:
        try:
            dest_path.mkdir(parents=True, exist_ok=True)
        except Exception as e:
            return [f"创建目标目录失败: {str(e)}"]
    
    for item in items:
        try:
            item_path = _safe_path(item)
            if not item_path.exists():
                results.append(f"项目不存在: {item}")
                continue
                
            if dest_path.is_dir():
                target = dest_path / item_path.name
            else:
                target = dest_path
                
            # 根据类型执行不同的复制
            if item_path.is_file():
                shutil.copy2(str(item_path), str(target))
            elif item_path.is_dir():
                shutil.copytree(str(item_path), str(target))
                
            results.append(f"成功复制: {item} -> {target}")
        except Exception as e:
            results.append(f"复制 {item} 时出错: {str(e)}")
    
    return results

### 权限控制
@mcp.tool(name="chmod_items", description="更改多个项目的POSIX权限")
def chmod_items(items: List[str], permissions: int) -> List[str]:
    """更改多个项目的POSIX权限"""
    results = []
    for item in items:
        try:
            item_path = _safe_path(item)
            if not item_path.exists():
                results.append(f"项目不存在: {item}")
                continue
                
            # 使用整数权限值
            os.chmod(item_path, permissions)
            results.append(f"成功更改权限: {item} -> {oct(permissions)}")
        except Exception as e:
            results.append(f"更改 {item} 的权限时出错: {str(e)}")
    
    return results

if __name__ == "__main__":
    mcp.run(transport="stdio")
    

# 本地调试
# from fastmcp import Client # Import the client

# async def test_server_locally():
#     print("\n--- Testing Server Locally ---")
#     # Point the client directly at the server object
#     client = Client(mcp)

#     # Clients are asynchronous, so use an async context manager
#     async with client:
#         # Call the 'greet' tool
#         greet_result = await client.call_tool("greet", {"name": "FastMCP User"})
#         print(f"greet result: {greet_result}")

#         # Get the 'summarize' prompt structure (doesn't execute the LLM call here)
#         prompt_messages = await client.get_prompt("summarize", {"text": "This is some text."})
#         print(f"Summarize prompt structure: {prompt_messages}")

# Run the local test function
# asyncio.run(test_server_locally())
# Commented out for now, we'll focus on running the server next