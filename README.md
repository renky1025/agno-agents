# 多智能体框架 - AGNO

## 安装依赖

```shell
uv venv
# active venv
uv pip install -U agno duckduckgo-search httpx "mcp[cli]" ollama yfinance jira
```

## 创建agno agents

```python
    ollama_model = Ollama(id="qwen2.5:14b",
                      show_tool_calls=True,
                      timeout=30,
                      host= "http://192.168.100.80:11434")
    #command = "python C:/workspaces/python-projects/agno-agents/server.py"
    command = "C:/workspaces/python-projects/agno-agents/golang-mcp/main.exe"
    #mcp_tools = MCPTools(server_params=server_params)
    # MCP server to access the filesystem (via `npx`)
    async with MCPTools(command=command) as mcp_tools:
        agent = Agent(
            exponential_backoff=True,
            #debug_mode=True,
            model=ollama_model,
            name="My_Agent",
            tools=[mcp_tools],
            markdown=True,
            show_tool_calls=True,
        )

        await agent.aprint_response(
           ## "what's the current time?",
            "计算20乘以50等于多少?",
            stream=True,
        )

```

## 连接 MCP Server 获取mcp tools

## 多agent

```shell
#方式一
    Agent(
        name="多智能体Team",
        model=ollama_model,
        team=[
        json_format_agent, xxxx,xxx,
        ],
        instructions=[
        """prompt"""
        ],
        #success_criteria="图表生成成功，有quickchart的链接.",
        #enable_agentic_context=True,
        show_tool_calls=True,
        markdown=True,
        add_context=True,
        add_datetime_to_instructions=True,
        add_history_to_messages=True,
        num_history_responses=5,
        #show_members_responses=True,
    )
#方式二
    Team(
        name="多智能体Team",
        model=ollama_model,
        members=[
        json_format_agent, xxxx,xxx,
        ],
        instructions=[
        """prompt"""
        ],
        success_criteria="图表生成成功，有quickchart的链接.",
        enable_agentic_context=True,
        show_tool_calls=True,
        markdown=True,
        add_context=True,
        add_datetime_to_instructions=True,
        add_history_to_messages=True,
        num_history_responses=5,
        show_members_responses=True,
    )

# 载入多个mcp tools 目前测试只有这个方式可以工作
 async with MultiMCPTools([datavcommand, pgcommand]) as mcp_tools:
        # 创建MCP代理
        if mcp_tools:
            try:
                mcp_agent = Agent(
                   name="MCP Tools",
                    model=ollama_model,
                    tools=[mcp_tools],
                )

```

## Get Started with Agent UI

```shell

npx create-agent-ui@latest
cd agent-ui && pnpm run dev
## 修改启动端口， package.json ==>"dev": "next dev -p 4000",
##Open http://localhost:4000 to view the Agent UI,

#main.py
app = Playground(agents=[web_agent, finance_agent, mcp_agent, jira_agent]).get_app()
# cors issue
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:4000"],  # Add the URL of your Agent UI
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

if __name__ == "__main__":
    serve_playground_app("main:app", reload=False)
```

## test /debug mcp tools

```shell
pnpm install @modelcontextprotocol/inspector

npx @modelcontextprotocol/inspector go run main.go
```
