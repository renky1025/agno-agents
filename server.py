from mcp.server.fastmcp import FastMCP
import datetime
import os

mcp = FastMCP(name="test_tools")


# 添加一个动态问候资源
@mcp.resource("greeting://{name}")
def get_greeting(name: str) -> str:
    """获取个性化问候语"""
    return f"Hello, {name}!"

@mcp.tool()
def get_handbook_section(section: str) -> str:
    return f"You asked for section: {section}"


@mcp.tool()
def get_current_time() -> str:
    """Get the current time in ISO format"""
    return datetime.datetime.now().isoformat()

@mcp.resource("files_in_directory://{directory}")
def list_files(directory: str) -> list[str]:
    """List files in the specified directory"""
    return [file for file in os.listdir(directory) if os.path.isfile(os.path.join(directory, file))]


if __name__ == "__main__":
    mcp.run(transport="stdio")