#!/usr/bin/python
# -*- coding: utf-8 -*-
from textwrap import dedent
from agno.agent import Agent  # noqa
from agno.models.ollama.chat import Ollama
from agno.tools.duckduckgo import DuckDuckGoTools
from agno.tools.yfinance import YFinanceTools
from agno.tools.mcp import MCPTools
from agno.tools.mcp import MultiMCPTools
import asyncio
from agno.playground import Playground, serve_playground_app
from agno.storage.sqlite import SqliteStorage
import os
from fastapi.middleware.cors import CORSMiddleware
import anyio
from agno.tools.jira import JiraTools

ollama_model = Ollama(id="mistral-small3.1:latest",
                      show_tool_calls=True,
                      timeout=120,
                      host= "http://10.100.1.1:11434")

# 创建一个惰性加载的MCP代理
class LazyMCPAgent(Agent):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._mcp_agent = None
        self._lock = asyncio.Lock()
        self._mcp_tools = None
        
    async def arun(self, message, *args, **kwargs):
        max_retries = 3
        retry_count = 0
        
        while retry_count < max_retries:
            try:
                print(f"开始处理消息: {message[:30]}...")
                async with self._lock:
                    # 检查连接状态并在需要时重新初始化
                    if self._mcp_agent is None or not await self._check_and_reconnect():
                        print(f"尝试初始化MCP代理 (尝试 {retry_count + 1}/{max_retries})")
                        success = await self._initialize_agent()
                        if not success:
                            retry_count += 1
                            if retry_count < max_retries:
                                print("初始化失败，等待1秒后重试...")
                                await asyncio.sleep(1)
                                continue
                            return f"无法初始化MCP代理，请检查MCP服务是否正常运行"
                    
                    try:
                        result = await self._mcp_agent.arun(message=message, *args, **kwargs)
                        print(f"MCP代理成功处理消息: {message[:30]}...")
                        return result
                    except anyio.ClosedResourceError:
                        print("检测到连接已关闭，尝试重新初始化...")
                        # 强制重新初始化
                        self._mcp_agent = None
                        retry_count += 1
                        if retry_count < max_retries:
                            continue
                        raise
            except Exception as e:
                print(f"MCP代理处理消息时出错: {str(e)}")
                import traceback
                traceback.print_exc()
                retry_count += 1
                if retry_count < max_retries:
                    print(f"等待1秒后重试... ({retry_count}/{max_retries})")
                    await asyncio.sleep(1)
                    continue
                # 所有重试都失败后返回错误消息
                return f"处理查询时发生错误: {str(e)}"
        
        return "达到最大重试次数，操作失败"
    
    async def _initialize_agent(self):
        print("延迟初始化MCP代理...")
        # 设置环境变量
        pgcommand = "C:/workspaces/MCPCommand/go-mcp-postgres/go-mcp-postgres-windows-64.exe --dsn postgresql://username:password@10.100.2.1:5433/aiproxy"
        k8scommand = "C:/workspaces/MCPCommand/mcp-k8s/mcp-k8s_windows_amd64.exe --kubeconfig=C:/workspaces/MCPCommand/10.100.0.4/202504111125/config"
        try:
            mcp_tools = MultiMCPTools(commands=[k8scommand, pgcommand])
            # 添加超时处理
            print("开始初始化MCP工具...")
            await asyncio.wait_for(mcp_tools.__aenter__(), timeout=30)
            print("MCP工具初始化成功")
            self._mcp_agent = Agent(
                name="MCP Tools",
                model=ollama_model,
                tools=[mcp_tools],
                instructions=dedent("""\
                    ## Using the mcp tools
                    Before taking any action or responding to the user after receiving tool results, use the mcp tools as a scratchpad to:
                    - List the specific rules that apply to the current request
                    - 数据库使用的是 postgresql, 注意生成sql语句要适配postgresql
                    - 表格形式返回结果

                    ## Rules
                    - Use tables to display data where possible\
                    """),
                show_tool_calls=True,
                markdown=True,
                add_context=True,
                add_datetime_to_instructions=True,
                add_history_to_messages=True,
                num_history_responses=5,
                # debug_mode=True,
            )
            if hasattr(self._mcp_agent, 'tools') and self._mcp_agent.tools:
                print(f"MCP代理初始化成功，可用工具: {len(self._mcp_agent.tools)}")
            else:
                print("警告: MCP代理初始化完成，但工具列表可能为空")
                
            # 保存MCP工具实例以便后续检查状态
            self._mcp_tools = mcp_tools
            return True
        except asyncio.TimeoutError:
            print("初始化MCP工具超时，请检查外部进程是否能正常运行")
            # 尝试清理资源
            try:
                if hasattr(mcp_tools, '_stdio_context') and mcp_tools._stdio_context:
                    await mcp_tools._stdio_context.__aexit__(None, None, None)
            except Exception as cleanup_err:
                print(f"清理资源时出错: {cleanup_err}")
            return False
        except Exception as e:
            print(f"初始化MCP工具时出现其他错误: {e}")
            self._mcp_agent = None
            self._mcp_tools = None
            return False
            
    async def _check_and_reconnect(self):
        """检查MCP连接状态，如果已关闭则重新连接"""
        if not hasattr(self, '_mcp_tools') or self._mcp_tools is None:
            print("MCP工具尚未初始化，将尝试初始化")
            return await self._initialize_agent()
            
        # 检查MCP工具连接状态
        try:
            # 简单的连接状态检查
            if hasattr(self._mcp_tools, '_session') and self._mcp_tools._session:
                if getattr(self._mcp_tools._session, '_closed', False):
                    print("检测到MCP连接已关闭，将重新初始化")
                    # 先尝试清理旧资源
                    try:
                        await self._mcp_tools.__aexit__(None, None, None)
                    except:
                        pass  # 忽略清理错误
                    # 重新初始化
                    return await self._initialize_agent()
            return True
        except Exception as e:
            print(f"检查MCP连接状态时出错: {e}")
            return False


# 创建惰性加载的MCP代理
mcp_agent = LazyMCPAgent(
    name="MCP Tools",
    model=ollama_model,
    tools=[],  # 工具将在第一次使用时添加
    instructions="Using MCP tools to help user",
    show_tool_calls=True,
    markdown=True,
    add_context=True,
    add_datetime_to_instructions=True,
    add_history_to_messages=True,
    num_history_responses=5,
    # debug_mode=True,
)



agent_storage: str = "tmp/agents.db"

web_agent = Agent(
    name="Web Agent",
    model=ollama_model,
    tools=[DuckDuckGoTools()],
    instructions=["Always include sources"],
    # Store the agent sessions in a sqlite database
    storage=SqliteStorage(table_name="web_agent", db_file=agent_storage),
    # Adds the current date and time to the instructions
    add_datetime_to_instructions=True,
    # Adds the history of the conversation to the messages
    add_history_to_messages=True,
    # Number of history responses to add to the messages
    num_history_responses=5,
    # Adds markdown formatting to the messages
    markdown=True,
)

finance_agent = Agent(
    name="Finance Agent",
    model=ollama_model,
    tools=[YFinanceTools(stock_price=True, analyst_recommendations=True, company_info=True, company_news=True)],
    instructions=["Always use tables to display data"],
    storage=SqliteStorage(table_name="finance_agent", db_file=agent_storage),
    add_datetime_to_instructions=True,
    add_history_to_messages=True,
    num_history_responses=5,
    markdown=True,
)


env ={
    "JIRA_SERVER_URL": "http://jira.xxx.com",## https 需要添加证书
    "JIRA_USERNAME": "kangyao.ren",
    "JIRA_PASSWORD": "changeme"
}

jira_agent = Agent(
    name="Jira Agent",
    model=ollama_model,
    tools=[JiraTools(server_url=env["JIRA_SERVER_URL"], username=env["JIRA_USERNAME"], password=env["JIRA_PASSWORD"])],
    instructions=["Always use tables to display data"],
    storage=SqliteStorage(table_name="jira_agent", db_file=agent_storage),
    add_datetime_to_instructions=True,
    add_history_to_messages=True,
    num_history_responses=5,
    markdown=True,
)


app = Playground(agents=[web_agent, finance_agent, mcp_agent, jira_agent]).get_app()
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:4000"],  # Add the URL of your Agent UI
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

if __name__ == "__main__":
    # 首先测试MCP代理是否正常工作
    # async def test_mcp_agent():
    #     try:
    #         print("开始测试MCP代理...")
    #         response = await mcp_agent.arun(message="查询logs前十条数据")
    #         print("MCP代理测试结果:", response)
    #         return True
    #     except Exception as e:
    #         print(f"测试MCP代理时出错: {e}")
    #         return False

    # # 运行测试
    # import asyncio
    # test_success = asyncio.run(test_mcp_agent())
    
    # # 如果测试成功，启动Playground应用
    # if test_success:
    #     print("测试成功，启动Playground应用...")
    #     serve_playground_app("main:app", reload=True)
    # else:
    #     print("测试失败，请检查MCP代理配置...")
    #     # 仍然启动应用，但显示警告
    try:
        print("启动Playground应用，MCP代理将在第一次使用时初始化...")
        serve_playground_app("main:app", reload=False)
    except Exception as e:
        print(f"启动Playground应用时出错: {e}")
        import traceback
        traceback.print_exc()