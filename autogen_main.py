## pip install pyautogen autogen-agentchat  autogen-ext[ollama] autogen-ext[mcp] json_schema_to_pydantic
import asyncio
from autogen_ext.tools.mcp import StdioServerParams, mcp_server_tools
from autogen_agentchat.agents import AssistantAgent
from autogen_core import CancellationToken
from autogen_agentchat.conditions import  TextMentionTermination, MaxMessageTermination
from autogen_agentchat.teams import RoundRobinGroupChat
from autogen_ext.models.ollama import OllamaChatCompletionClient
from autogen_agentchat.ui import Console

ollama_client = OllamaChatCompletionClient(
    model="mistral-small3.1:latest",
    host="http://10.100.1.1:11434",
    temperature=0.1,
    model_info = {
        "vision": False,
        "function_calling": True,
        "json_output": True,
        "family": "unknown",
        "structured_output": True,
    }
)

datavcommand = "C:/workspaces/python-projects/agno-agents/mcp_tools/Quickchart-MCP-Server/go-mcp-quickchart.exe"
pgcommand = "C:/workspaces/python-projects/agno-agents/mcp_tools/go-mcp-postgres/go-mcp-postgres.exe --host 10.100.2.1 --port 5433 --name aiproxy --user username --password password --sslmode disable"
mongcommand="C:/workspaces/python-projects/agno-agents/mcp_tools/go-mcp-mongodb/go-mcp-mongodb.exe --user myusername --password mypassword --host 10.100.2.1 --port 27017 --auth admin --dbname fastgpt"
lokicommand="C:/workspaces/python-projects/agno-agents/mcp_tools/loki-mpc/loki-mcp.exe"
env = {
    "LOKI_URL": "http://10.100.0.4:32004",
}

async def get_mcptools(command: str):
    pgarray = command.split(" ")
    pgcmd = pgarray[0]
    pgargs = pgarray[1:]
    server_params = StdioServerParams(
        command=pgcmd, args=pgargs, read_timeout_seconds=120
    )
    tools = await mcp_server_tools(server_params)
    return tools


async def main() -> None:
    # Setup server params for local filesystem access
    # Get all available tools from the server
    pgtools = await get_mcptools(pgcommand)
    datavtools = await get_mcptools(datavcommand)
    mongo_tools = await get_mcptools(mongcommand)
    lokitools = await get_mcptools(lokicommand)

    # Create an agent that can use all the tools
    pg_agent = AssistantAgent(
        name="postgres_manager",
        system_message="你是一个postgres数据库助手,你可以根据用户输入理解生成完美合理的Postgresql sql查询语句，并能够连接数据库执行查询数据返回给用户，用不需要提醒你去执行，你可以帮忙执行sql返回查询结果。",
        model_client=ollama_client,
        tools=pgtools,  # type: ignore
        reflect_on_tool_use=True,
    )
    # Create an agent that can use all the tools
    mongo_agent = AssistantAgent(
        name="mongo_manager",
        system_message="你是一个mongo数据库助手, 可以查询MongoDB数据库。",
        model_client=ollama_client,
        tools=mongo_tools,  # type: ignore
        reflect_on_tool_use=True,
    )
        # Create an agent that can use all the tools
    datav_agent = AssistantAgent(
        name="datav_manager",
        system_message="""你是一个数据生成图表助手, 验证输入json格式是否符合quickchart工具要求， 如果符合， 调用generate_chart工具生成图表。
```json
{
  "type": "line" ,
  "data": {
    "labels": ["Label 1", "Label 2", "Label 3"],
    "datasets": [{
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
```, 并保证json数据完整性，不要随意截断json数据， 调用generate_chart工具生成图表。并返回图片的markdown代码""",
        model_client=ollama_client,
        tools=datavtools,  # type: ignore
        reflect_on_tool_use=True,
    )
        # Create an agent that can use all the tools
    loki_agent = AssistantAgent(
        name="loki_manager",
        system_message="你是一个loki日志管理助手, 可以查询loki数据。",
        model_client=ollama_client,
        tools=lokitools,  # type: ignore
        reflect_on_tool_use=True,
    )
    # Create rewriter Agent (unchanged)
    json_agent = AssistantAgent(
        name="json_parser",
        system_message="""你是一个json解析专家。将提供给你的内容解析为json格式， quickchart工具需要json格式数据如下：\n
```json
{
  "type": "line" ,
  "data": {
    "labels": ["Label 1", "Label 2", "Label 3"],
    "datasets": [{
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
```        
        """,
        model_client=ollama_client,
    )
    # 查询logs表表结构
    # result = await pg_agent.run(task="查询logs表表结构", cancellation_token=CancellationToken())
    # print(result.messages[-1].content)
    # 初始化用户代理


    
    termination = TextMentionTermination("TERMINATE") | MaxMessageTermination(10)

    team = RoundRobinGroupChat([pg_agent, json_agent, datav_agent], termination_condition=termination)
    config = team.dump_component()
    print(config.model_dump_json())
    try:
    # Start the team and wait for it to terminate.
        # result = await team.run(
        #     task="查询Postgresql数据库的logs 表，统计最近10天，model字段模型每天使用量， 和平均响应时间（ttfb_milliseconds）， 并生成图表",
        #     cancellation_token=CancellationToken()
        # )
        message="查询Postgresql数据库的logs 表，统计最近10天，model字段模型每天使用量，并调用工具生成quickchart图表"
        stream = team.run_stream(task=message)
    # `Console` is a simple UI to display the stream.
        await Console(stream)
    finally:
        await team.close()
        await ollama_client.close()
        

    # return result

if __name__ == "__main__":
    asyncio.run(main())