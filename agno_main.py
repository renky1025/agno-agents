# -*- coding: utf-8 -*-
from textwrap import dedent
from agno.agent import Agent  # noqa
from agno.models.ollama.chat import Ollama
from agno.tools.duckduckgo import DuckDuckGoTools
from agno.tools.yfinance import YFinanceTools
#from agno.tools.mcp import MCPTools
from agno.tools.mcp import MultiMCPTools
#from agno.team import Team
from agno.playground import Playground, serve_playground_app
from agno.storage.sqlite import SqliteStorage
from fastapi.middleware.cors import CORSMiddleware
# from agno.tools.jira import JiraTools
import asyncio
import nest_asyncio
from chart_models import ChartConfig
import os

import logging
logging.basicConfig(level=logging.INFO)

logger = logging.getLogger(__name__)

# Allow nested event loops
nest_asyncio.apply()

ollama_model = Ollama(id="mistral-small3.1:latest",
                      show_tool_calls=True,
                      timeout=120,
                      host= "http://10.100.1.1:11434")

datavcommand = "C:/workspaces/python-projects/agno-agents/mcp_tools/Quickchart-MCP-Server/go-mcp-quickchart.exe"
pgcommand = "C:/workspaces/python-projects/agno-agents/mcp_tools/go-mcp-postgres/go-mcp-postgres.exe --host 10.100.2.1 --port 5433 --name aiproxy --user username --password password --sslmode disable"
mongcommand="C:/workspaces/python-projects/agno-agents/mcp_tools/go-mcp-mongodb/go-mcp-mongodb.exe --user myusername --password mypassword --host 10.100.2.1 --port 27017 --auth admin --dbname fastgpt"
lokicommand="C:/workspaces/python-projects/agno-agents/mcp_tools/loki-mpc/loki-mcp.exe"
k8scommand="C:/workspaces/python-projects/agno-agents/mcp_tools/go-mcp-k8s/go-mcp-k8s.exe --kubeconfig C:/workspaces/MCPCommand/10.100.0.4/202504111125/config"

env = {
    "LOKI_URL": "http://10.100.0.4:32004",
}

def get_agent_storage(table_name: str):
    agent_storage_file: str = "tmp/agno_agents.db"
    
    return SqliteStorage(
        table_name=table_name,
        db_file=agent_storage_file,
        auto_upgrade_schema=True,
    )

web_agent = Agent(
    name="Web Agent",
    model=ollama_model,
    tools=[DuckDuckGoTools()],
    instructions=["Always include sources"],
    storage=get_agent_storage("web_agent"),
    add_datetime_to_instructions=True,
    add_history_to_messages=True,
    num_history_responses=5,
    markdown=True,
)

finance_agent = Agent(
    name="Finance Agent",
    model=ollama_model,
    tools=[YFinanceTools(stock_price=True, analyst_recommendations=True, company_info=True, company_news=True)],
    instructions=["Always use tables to display data"],
    storage=get_agent_storage("finance_agent"),
    add_datetime_to_instructions=True,
    add_history_to_messages=True,
    num_history_responses=5,
    markdown=True,
)

# Agent that uses JSON mode
json_format_agent = Agent(
    name="JSON Format Agent",
    model=ollama_model,
    description=dedent(""" \n 你是数据转换大师, 在需要生成图表之前，你会将数据转换为json数据结构\n,再传递给图表生成工具。 """),
    instructions=dedent(""" \n 你是数据转换大师, 在需要生成图表之前，你会将数据转换为json数据结构，\n
固定格式如下：
```json
{
  "type":"line",
  "data": {
    "labels": ["Label 1", "Label 2", "Label 3"],
    "datasets": [{
      "type": "bar" | "line" | "pie" | "scatter" | "area",
      "fill": false,
      "label": "Series Name",
      "data": [10, 20, 30],
      "backgroundColor": "rgb(75, 192, 192)"
    }]
  },
  "options": {
    "title": {
      "display": true,
      "text": "Chart Title"
    }
  }
}
``` """),
    response_model=ChartConfig,
    use_json_mode=True,
    storage=get_agent_storage("json_format_agent"),
    add_datetime_to_instructions=True,
    add_history_to_messages=True,
    num_history_responses=5,
   
)

# # 创建图表生成代理
# async def create_team_with_agents():
#     """创建并初始化所有代理及团队
    
#     返回包含所有MCP工具代理的团队
#     """
#     # 创建团队
#     teamGroup = Agent(
#         name="图表生成Team",
      
#         model=ollama_model,
#         team=[
#             json_format_agent,
#         ],
#         instructions=[
#             "使用generate_chart工具创建图表。",
#             "使用Markdown语法将图表作为嵌入图像返回：`![](chart_link)`，而不是纯URL。",
#         ],
#         #success_criteria="图表生成成功，有quickchart的链接.",
#         #enable_agentic_context=True,
#             show_tool_calls=True,
#             markdown=True,
#             add_context=True,
#             add_datetime_to_instructions=True,
#             add_history_to_messages=True,
#             num_history_responses=5,
#         #show_members_responses=True,
#     )
    
#     # 使用上下文管理器同时创建所有MCP会话，这样它们在函数结束前都会保持打开状态
#     async with (
#         MCPTools(command=datavcommand) as datav_mcp_tools,
#         MCPTools(command=pgcommand) as postgres_mcp_tools,
#         # MCPTools(command=mongcommand) as mongo_mcp_tools,
#         # MCPTools(command=lokicommand, env=env) as loki_mcp_tools
#     ):
#         # 创建Datav图表代理
#         datav_agent = Agent(
#             name="Quickchart图表生成 MCP Tools",
#             model=ollama_model,
#             tools=[datav_mcp_tools],
#             instructions=dedent("""\
# ## 你是一位图表生成大师，使用工具完成以下任务：

# ### 你的任务是根据用户请求生成图表。

# ### 3. 结果格式和返回规则：
# - **如果用户请求图表**，请按照以下步骤操作：
# - 检查结果是否已经是期望的JSON格式；
# ```json
# {
#   "type": "line" ,
#   "data": {
#     "labels": ["Label 1", "Label 2", "Label 3"],
#     "datasets": [{
#       "fill": false,
#       "label": "Series Name",
#       "data": [10, 20, 30],
#       "backgroundColor": "rgb(75, 192, 192)"
#     }]
#   },
#   "options": {
#     "title": {
#       "display": true,
#       "text": "Chart Title"
#     }
#   }
# }
# ```
# - 自动选择最适合的图表类型（例如，`line`用于趋势，`bar`或`pie`用于类别）；
# - 使用generate_chart工具创建图表；
# - 使用Markdown语法将图表作为嵌入图像返回：`![](chart_link)`，而不是纯URL。

# ### 4. 附加指南：
# - 如果需要多个图表，请单独按顺序渲染每个图表。

# """),
#             storage=get_agent_storage("datav_mcp_agent"),
#             show_tool_calls=True,
#             markdown=True,
#             add_context=True,
#             add_datetime_to_instructions=True,
#             add_history_to_messages=True,
#             num_history_responses=5,
#         )
        
#         # 创建PostgreSQL代理
#         postgres_agent = Agent(
#             name="Postgres MCP Tools",
#             model=ollama_model,
#             tools=[postgres_mcp_tools],
#             instructions=dedent("""\
# ## 你是一位Postgres数据库大师，使用工具完成以下任务：

# ### 任务：
# - 根据用户请求生成postgres sql
# - 保证sql语句语法正确，并能够正确执行
# - 执行sql语句并返回结果
# - 完全按照用户请求生成sql语句，不要出现自己假设的数据
# - 如果你不知道表中的列，可以使用postgres_describe_table工具获取表结构
# - 如果你不知道表名，可以使用postgres_list_tables工具获取表名

# ### 结果格式：
# - json形式返回数据库查询结果

#               """),
#             storage=get_agent_storage("postgres_mcp_agent"),
#             show_tool_calls=True,
#             markdown=True,
#             add_context=True,
#             add_datetime_to_instructions=True,
#             add_history_to_messages=True,
#             num_history_responses=5,
#         )
        
#         # 添加代理到团队
#         teamGroup.team.append(datav_agent)
#         teamGroup.team.append(postgres_agent)
        
#         logger.info("团队和代理创建成功，开始处理请求...")
        
#         # # 在上下文管理器内执行团队任务，确保MCP连接保持打开
#         # await teamGroup.aprint_response(
#         #     "查询Postgresql数据库的表名logs, model字段存放模型名称，查询最近10天模型每日使用量，和每个模型响应平均时间，查询出来数据转换为json格式，再生成图表", 
#         #     stream=False, 
#         #     markdown=True, 
#         #     show_reasoning=True
#         # )
    
#     return teamGroup

# async def run_server():
#     # 启用debug模式以便查看详细错误信息
#     os.environ["AGNO_DEBUG"] = "true"
    
#     try:
#         # 创建团队并执行任务
#       teamGroup = await create_team_with_agents()
#            #创建应用，同时包含所有代理
#       agents = [web_agent, finance_agent, json_format_agent, teamGroup]
#       app = Playground(agents=agents).get_app()
#       app.add_middleware(
#           CORSMiddleware,
#           allow_origins=["http://localhost:4000"],
#           allow_credentials=True,
#           allow_methods=["*"],
#           allow_headers=["*"],
#       )
      
#       # 使用nest_asyncio允许嵌套事件循环
#       #print("Playground应用已准备就绪，服务开始运行...")
#       serve_playground_app(app)
#     except Exception as e:
#         logger.error(f"处理请求时出错: {e}")
    



async def run_server() -> None:
    os.environ["AGNO_DEBUG"] = "true"
    # Create a client session to connect to the MCP server
    async with MultiMCPTools([datavcommand, pgcommand]) as mcp_tools:
        # 创建MCP代理
        if mcp_tools:
            try:
                mcp_agent = Agent(
                    name="MCP Tools",
                    model=ollama_model,
                    tools=[mcp_tools],
                    instructions=dedent("""\
## You are a data analysis assistant that uses tools to complete the following tasks:

### 1. Intelligently select the appropriate data source based on the user's request:

- **PostgreSQL databases**: Use the `postgres_*` tools;
- **MongoDB databases**: Use the `mongo_*` tools;
- **Server logs**: Use the `loki_*` tools;
- If the internal structure of a table is unknown, first call the appropriate `*_describe_table` tool to retrieve the table schema.

### 2. Construct and execute accurate query statements:

- Based on the table schema and user intent, construct valid query statements using the correct syntax for the target database (SQL for PostgreSQL/MongoDB, LogQL for Loki);
- Ensure the selected fields and logic match the user's requirements;
- Execute the query using the appropriate `mcptools` command.

### 3. Result formatting and return rules:

- **By default, return data in a tabular format**;
- **If the user requests a chart**, follow these steps:
  - Check whether the result is already in the expected JSON format (as shown below);
  - Automatically choose the most appropriate chart type (e.g., `line` for trends, `bar` or `pie` for categories);
  - Use the `generate_chart` tool to create the chart;
  - Return the chart as an embedded image using Markdown syntax: `![](chart_link)`, not as a plain URL.

### 4. Additional guidelines:

- Ensure the query results are complete and properly cleaned;
- The chart must clearly represent the key data relationship in the user's question;
- If the user does not specify the database type, infer it from the query intent;
- If multiple charts are needed, render each one separately and sequentially.

"""),
                    storage=get_agent_storage("mcp_agent"),
                    show_tool_calls=True,
                    markdown=True,
                    add_context=True,
                    add_datetime_to_instructions=True,
                    add_history_to_messages=True,
                    num_history_responses=5,
                )
                #print("MCP代理创建成功")
                # 创建应用，同时包含所有代理
                agents = [web_agent, finance_agent, json_format_agent]
                if mcp_agent:
                    agents.append(mcp_agent)

                app = Playground(agents=agents).get_app()
                app.add_middleware(
                    CORSMiddleware,
                    allow_origins=["http://localhost:4000"],
                    allow_credentials=True,
                    allow_methods=["*"],
                    allow_headers=["*"],
                )
                serve_playground_app(app)
            except Exception as e:
                print(f"创建MCP代理时出错: {e}")
        else:
            print("MCP工具初始化失败，无法创建MCP代理")




if __name__ == "__main__":
    try:
        logger.info("启动Playground应用...")
        # 确保我们有一个适当的事件循环
        asyncio.run(run_server())
    except Exception as e:
        logger.error(f"启动Playground应用时出错: {e}")