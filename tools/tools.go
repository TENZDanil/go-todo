package tools

import (
	"encoding/json"
	"fmt"
)

type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

func GetBrowserTools() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "navigate",
				Description: "Переходит на указанный URL",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{
							"type":        "string",
							"description": "URL страницы для перехода",
						},
					},
					"required": []string{"url"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "get_page_content",
				Description: "Получает содержимое текущей страницы",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "click_element",
				Description: "Кликает на элемент страницы по CSS селектору",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"selector": map[string]interface{}{
							"type":        "string",
							"description": "CSS селектор элемента",
						},
					},
					"required": []string{"selector"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "fill_input",
				Description: "Заполняет поле ввода текстом",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"selector": map[string]interface{}{
							"type":        "string",
							"description": "CSS селектор поля ввода",
						},
						"text": map[string]interface{}{
							"type":        "string",
							"description": "Текст для ввода",
						},
					},
					"required": []string{"selector", "text"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "get_elements",
				Description: "Получает информацию об элементах на странице по селектору",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"selector": map[string]interface{}{
							"type":        "string",
							"description": "CSS селектор для поиска элементов",
						},
					},
					"required": []string{"selector"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "get_page_info",
				Description: "Получает информацию о текущей странице",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "wait_for_element",
				Description: "Ждет появления элемента на странице",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"selector": map[string]interface{}{
							"type":        "string",
							"description": "CSS селектор элемента",
						},
						"timeout": map[string]interface{}{
							"type":        "integer",
							"description": "Время ожидания в секундах",
						},
					},
					"required": []string{"selector"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "complete_task",
				Description: "Завершает задачу",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"result": map[string]interface{}{
							"type":        "string",
							"description": "Результат выполнения задачи",
						},
					},
					"required": []string{"result"},
				},
			},
		},
	}
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func ParseArguments(argumentsJSON string, target interface{}) error {
	return json.Unmarshal([]byte(argumentsJSON), target)
}

type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Role       string `json:"role"`
	Content    string `json:"content"`
}

func NewToolResult(toolCallID string, content string) ToolResult {
	return ToolResult{
		ToolCallID: toolCallID,
		Role:       "tool",
		Content:    content,
	}
}

func FormatToolsForOpenAI(tools []Tool) []Tool {
	return tools
}

type NavigateArgs struct {
	URL string `json:"url"`
}

type ClickElementArgs struct {
	Selector string `json:"selector"`
}

type FillInputArgs struct {
	Selector string `json:"selector"`
	Text     string `json:"text"`
}

type GetElementsArgs struct {
	Selector string `json:"selector"`
}

type WaitForElementArgs struct {
	Selector string `json:"selector"`
	Timeout  int    `json:"timeout,omitempty"`
}

type CompleteTaskArgs struct {
	Result string `json:"result"`
}

func FormatError(err error) string {
	return fmt.Sprintf("Ошибка: %v", err)
}
