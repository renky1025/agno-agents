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
lokicommand="C:/workspaces/python-projects/agno-agents/mcp_tools/loki-mpc/loki-mcp.exe"
env = {
    "LOKI_URL": "http://10.100.0.4:32004",
}
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
    async with MultiMCPTools([datavcommand, pgcommand, mongcommand, lokicommand], env=env) as mcp_tools:
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
- Ensure the selected fields and logic match the user’s requirements;
- Execute the query using the appropriate `mcptools` command.

### 3. Result formatting and return rules:

- **By default, return data in a tabular format**;
- **If the user requests a chart**, follow these steps:
  - Check whether the result is already in the expected JSON format (as shown below);
  - If not, convert the data to the following standardized JSON structure:

```json
{
  "type": "bar" | "line" | "pie" | "scatter" | "area",
  "data": {
    "labels": ["Label 1", "Label 2", "Label 3"],
    "datasets": [{
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
```

  - Automatically choose the most appropriate chart type (e.g., `line` for trends, `bar` or `pie` for categories);
  - Use the `generate_chart` tool to create the chart;
  - Return the chart as an embedded image using Markdown syntax: `![](chart_link)`, not as a plain URL.

### 4. Additional guidelines:

- Ensure the query results are complete and properly cleaned;
- The chart must clearly represent the key data relationship in the user's question;
- If the user does not specify the database type, infer it from the query intent;
- If multiple charts are needed, render each one separately and sequentially.

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