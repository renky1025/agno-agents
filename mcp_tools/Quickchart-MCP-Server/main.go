package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"

	quickchartgo "github.com/henomis/quickchart-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const QUICKCHART_BASE_URL = "https://quickchart.io/chart"

type ChartConfig struct {
	Type string `json:"type"`
	Data struct {
		Labels   []string `json:"labels,omitempty"`
		Datasets []struct {
			Label            string                 `json:"label,omitempty"`
			Data             interface{}            `json:"data"`
			BackgroundColor  interface{}            `json:"backgroundColor,omitempty"`
			BorderColor      interface{}            `json:"borderColor,omitempty"`
			AdditionalConfig map[string]interface{} `json:"-"`
		} `json:"datasets"`
	} `json:"data"`
	Options struct {
		Title struct {
			Display bool   `json:"display,omitempty"`
			Text    string `json:"text,omitempty"`
		} `json:"title,omitempty"`
	} `json:"options,omitempty"`
}

func generateChartTool() mcp.Tool {
	return mcp.NewTool("generate_chart",
		mcp.WithDescription("Generate a chart using QuickChart, 使用QuickChart生成图表"),
		mcp.WithObject("config", mcp.Required(), mcp.Description("图表配置，使用JSON格式")),
	)
}

func main() {
	s := server.NewMCPServer(
		"quickchart-mcp-server 🚀",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithLogging(),
	)

	s.AddTool(generateChartTool(), generateChartToolHandler)

	if err := server.ServeStdio(s); err != nil {
		log.Printf("Server error: %v\n", err)
	}
}

func validateChartType(chartType string) {
	validTypes := []string{"bar", "line", "pie", "doughnut", "radar",
		"polarArea", "scatter", "bubble", "radialGauge", "speedometer"}
	if !slices.Contains(validTypes, chartType) {
		panic(fmt.Errorf("Invalid chart type. Must be one of: %v", validTypes))
	}
}

func generateChartToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := ChartConfig{}
	chartConfig, ok := request.Params.Arguments["config"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid chart configuration")
	}
	// Convert map[string]interface{} to ChartConfig
	jsonData, err := json.Marshal(chartConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshalling chart configuration: %w", err)
	}
	if err := json.Unmarshal(jsonData, &args); err != nil {
		return nil, fmt.Errorf("error unmarshalling chart configuration: %w", err)
	}
	if args.Type == "" {
		return nil, fmt.Errorf("chart type is required")
	}
	validateChartType(args.Type)
	if args.Data.Datasets == nil || len(args.Data.Datasets) == 0 {
		return nil, fmt.Errorf("datasets must be a non-empty array")
	}

	qc := quickchartgo.New()
	qc.Config = string(jsonData)
	// qc.Width = 600
	// qc.Height = 500
	qc.DevicePixelRation = 2.0
	qc.BackgroundColor = "transparent"
	quickchartURL, err := qc.GetShortUrl()
	if err != nil {
		return nil, fmt.Errorf("error generating chart: %w", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("%v", quickchartURL)), nil

}
