#!/usr/bin/python
# -*- coding: utf-8 -*-
from textwrap import dedent
from agno.agent import Agent  # noqa
from agno.models.ollama.chat import Ollama
from agno.tools.duckduckgo import DuckDuckGoTools
from agno.tools.yfinance import YFinanceTools
from agno.tools.mcp import MCPTools
from agno.tools.mcp import MultiMCPTools

from agno.playground import Playground, serve_playground_app
from agno.storage.sqlite import SqliteStorage
from fastapi.middleware.cors import CORSMiddleware
from agno.tools.jira import JiraTools
import asyncio
import nest_asyncio

# Allow nested event loops
nest_asyncio.apply()

ollama_model = Ollama(id="mistral-small3.1:latest",
                      show_tool_calls=True,
                      timeout=120,
                      host= "http://10.100.1.1:11434")


datavcommand = "C:/workspaces/python-projects/agno-agents/mcp_tools/Quickchart-MCP-Server/go-mcp-quickchart.exe"
pgcommand = "C:/workspaces/python-projects/agno-agents/mcp_tools/go-mcp-postgres/go-mcp-postgres.exe --host 10.100.2.1 --port 5433 --name aiproxy --user username --password password --sslmode disable"
mongcommand="C:/workspaces/python-projects/agno-agents/mcp_tools/go-mcp-mongodb/go-mcp-mongodb.exe --user myusername --password mypassword --host 10.100.2.1 --port 27017 --auth admin --dbname fastgpt"

agent_storage: str = "tmp/agents.db"
web_agent = Agent(
    name="Web Agent",
    model=ollama_model,
    tools=[DuckDuckGoTools()],
    instructions=["Always include sources"],
    storage=SqliteStorage(table_name="web_agent", db_file=agent_storage),
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
    storage=SqliteStorage(table_name="finance_agent", db_file=agent_storage),
    add_datetime_to_instructions=True,
    add_history_to_messages=True,
    num_history_responses=5,
    markdown=True,
)
async def run_server() -> None:
    # Create a client session to connect to the MCP server
    async with MultiMCPTools([datavcommand, pgcommand, mongcommand]) as mcp_tools:
        # 创建MCP代理
        if mcp_tools:
            try:
                mcp_agent = Agent(
                    name="MCP Tools",
                    model=ollama_model,
                    tools=[mcp_tools],
                    instructions=dedent("""\

## 你是一个数据分析助手，使用 `mcp tools` 工具完成以下任务：

1. 根据用户需求选择合适的数据库进行查询：
   - postgres_*：用于连接并查询 PostgreSQL 数据库；
   - mongo_*：用于连接并查询 MongoDB 数据库;
   - 如果你不清楚数据表内部字段,请先使用对应数据库的*_describe_table工具查询表内部结构.

2. 构造查询语句，从数据库中提取满足条件的数据。
   - 构造数据查询语句，需要根据前面获取的表结构，构造出合适的查询语句。
   - 查询语句需要根据使用数据库类型,使用对应的SQL语法。

3. 返回结果要求
    - 如果要求生成图表，优先检查输入数据是否符合如下json格式，如果符合，则直接使用输入数据调用generate_chart生成图表，否则将查询结果转换为如下固定 JSON 格式：
```json
{
  "type": "bar" | "line" | "pie" | "scatter" | "area",
  "data": {
    "labels": ["January", "February", "March"],#X轴显示标签
    "datasets": [{
      "label": "Sales",
      "data": [y轴数据65, 59, 80],#y轴数据
      "backgroundColor": "rgb(75, 192, 192)"
    }]
  },
  "options": {
    "title": {
      "display": true,
      "text": "标题"
    }
  }
}
```
- 如果用户没有要求，默认数据一个表格形式返回

4. 使用以上 JSON 数据用工具generate_chart生成图表，自动选择合适的图表类型（如趋势数据用 `line`，分类数据用 `bar` 或 `pie`）。

**注意事项：**
- 查询结果要确保字段完整，数据清洗正确；
- 图表内容需清晰表达用户问题中的核心数据关系；
- 若用户未指定数据库类型，请根据查询内容智能判断；
- 查询完毕后，请将生成的图表链接嵌入为图片展示：
    - 图表链接应以 Markdown 语法格式插入对话框：`![](图表链接)`;
    - 不要只展示链接文字，请以图片形式直接渲染；
- 若有多个图表，可依次展示多个图片。
"""),
                    debug_mode= True,
                    storage=SqliteStorage(table_name="mcp_agents", db_file=agent_storage),
                    show_tool_calls=True,
                    markdown=True,
                    add_context=True,
                    add_datetime_to_instructions=True,
                    add_history_to_messages=True,
                    num_history_responses=5,
                )
                print("MCP代理创建成功")
                # 创建应用，同时包含所有代理
                agents = [web_agent, finance_agent]
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
        print("启动Playground应用...")
        asyncio.run(run_server())
    except Exception as e:
        print(f"启动Playground应用时出错: {e}")