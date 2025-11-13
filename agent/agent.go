package agent

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"ai-browser-agent/browser"
	"ai-browser-agent/tools"

	"github.com/sashabaranov/go-openai"
)

type AIAgent struct {
	client        *openai.Client
	browser       *browser.BrowserManager
	conversation  []openai.ChatCompletionMessage
	maxIterations int
}

func NewAIAgent(browserManager *browser.BrowserManager) (*AIAgent, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY не установлен. Установите переменную окружения")
	}

	client := openai.NewClient(apiKey)

	agent := &AIAgent{
		client:        client,
		browser:       browserManager,
		conversation:  []openai.ChatCompletionMessage{},
		maxIterations: 20,
	}

	agent.initializeSystemPrompt()

	return agent, nil
}

func (a *AIAgent) initializeSystemPrompt() {
	a.conversation = []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
		},
	}
}

func (a *AIAgent) ExecuteTask(task string) (string, error) {
	fmt.Printf(" Задача: %s\n", task)
	fmt.Println("Агент начинает выполнение...\n")

	a.conversation = append(a.conversation, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: task,
	})

	availableTools := convertToolsToOpenAI(tools.GetBrowserTools())

	for iteration := 0; iteration < a.maxIterations; iteration++ {
		fmt.Printf(" Итерация %d/%d\n", iteration+1, a.maxIterations)

		req := openai.ChatCompletionRequest{
			Model:       openai.GPT4TurboPreview,
			Messages:    a.conversation,
			Tools:       availableTools,
			ToolChoice:  nil,
			Temperature: 0.7,
		}

		resp, err := a.client.CreateChatCompletion(context.Background(), req)
		if err != nil {
			return "", fmt.Errorf("ошибка запроса к OpenAI: %w", err)
		}

		assistantMessage := resp.Choices[0].Message
		a.conversation = append(a.conversation, assistantMessage)

		if len(assistantMessage.ToolCalls) == 0 {
			fmt.Println(" Агент завершил задачу без вызовов инструментов")
			if assistantMessage.Content != "" {
				return assistantMessage.Content, nil
			}
			continue
		}

		for _, toolCall := range assistantMessage.ToolCalls {
			fmt.Printf(" Вызов инструмента: %s\n", toolCall.Function.Name)

			result := a.executeTool(toolCall)

			a.conversation = append(a.conversation, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result.Content,
				ToolCallID: toolCall.ID,
			})

			fmt.Printf(" Результат: %s\n\n", truncateString(result.Content, 200))

			if toolCall.Function.Name == "complete_task" {
				var args tools.CompleteTaskArgs
				if err := tools.ParseArguments(toolCall.Function.Arguments, &args); err == nil {
					return args.Result, nil
				}
			}
		}
	}

	return "", fmt.Errorf("достигнуто максимальное количество итераций (%d). Задача может быть слишком сложной или требовать дополнительной информации", a.maxIterations)
}

func (a *AIAgent) executeTool(toolCall openai.ToolCall) tools.ToolResult {
	switch toolCall.Function.Name {
	case "navigate":
		var args tools.NavigateArgs
		if err := tools.ParseArguments(toolCall.Function.Arguments, &args); err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		err := a.browser.Navigate(args.URL)
		if err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		return tools.NewToolResult(toolCall.ID, fmt.Sprintf("Успешно перешел на страницу: %s", args.URL))

	case "get_page_content":
		html, err := a.browser.GetPageContent()
		if err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		text, err := a.browser.GetPageText()
		if err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		content := fmt.Sprintf("HTML (первые 5000 символов): %s\n\nТекст страницы (первые 3000 символов): %s",
			truncateString(html, 5000),
			truncateString(text, 3000))
		return tools.NewToolResult(toolCall.ID, content)

	case "get_page_info":
		url := a.browser.GetPageURL()
		title := a.browser.GetPageTitle()
		info := fmt.Sprintf("URL: %s\nЗаголовок: %s", url, title)
		return tools.NewToolResult(toolCall.ID, info)

	case "click_element":
		var args tools.ClickElementArgs
		if err := tools.ParseArguments(toolCall.Function.Arguments, &args); err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		err := a.browser.ClickElement(args.Selector)
		if err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		return tools.NewToolResult(toolCall.ID, fmt.Sprintf("Успешно кликнул на элемент: %s", args.Selector))

	case "fill_input":
		var args tools.FillInputArgs
		if err := tools.ParseArguments(toolCall.Function.Arguments, &args); err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		err := a.browser.FillInput(args.Selector, args.Text)
		if err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		return tools.NewToolResult(toolCall.ID, fmt.Sprintf("Успешно заполнил поле %s текстом: %s", args.Selector, args.Text))

	case "get_elements":
		var args tools.GetElementsArgs
		if err := tools.ParseArguments(toolCall.Function.Arguments, &args); err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		elements, err := a.browser.GetElements(args.Selector)
		if err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		if len(elements) == 0 {
			return tools.NewToolResult(toolCall.ID, fmt.Sprintf("Элементы с селектором '%s' не найдены", args.Selector))
		}
		var info strings.Builder
		info.WriteString(fmt.Sprintf("Найдено элементов: %d\n", len(elements)))
		for i, elem := range elements {
			if i >= 10 {
				info.WriteString(fmt.Sprintf("... и еще %d элементов\n", len(elements)-10))
				break
			}
			elemInfo := fmt.Sprintf("%d. Селектор: %s, Тег: %s, Текст: %s",
				i+1, elem.Selector, elem.Tag, truncateString(elem.Text, 100))
			if elem.Href != nil {
				elemInfo += fmt.Sprintf(", Ссылка: %s", *elem.Href)
			}
			if elem.ID != nil {
				elemInfo += fmt.Sprintf(", ID: %s", *elem.ID)
			}
			info.WriteString(elemInfo + "\n")
		}
		return tools.NewToolResult(toolCall.ID, info.String())

	case "wait_for_element":
		var args tools.WaitForElementArgs
		if err := tools.ParseArguments(toolCall.Function.Arguments, &args); err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		timeout := 10 * time.Second
		if args.Timeout > 0 {
			timeout = time.Duration(args.Timeout) * time.Second
		}
		err := a.browser.WaitForElement(args.Selector, timeout)
		if err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		return tools.NewToolResult(toolCall.ID, fmt.Sprintf("Элемент %s появился", args.Selector))

	case "complete_task":
		var args tools.CompleteTaskArgs
		if err := tools.ParseArguments(toolCall.Function.Arguments, &args); err != nil {
			return tools.NewToolResult(toolCall.ID, tools.FormatError(err))
		}
		return tools.NewToolResult(toolCall.ID, fmt.Sprintf("Задача завершена: %s", args.Result))

	default:
		return tools.NewToolResult(toolCall.ID, fmt.Sprintf("Неизвестный инструмент: %s", toolCall.Function.Name))
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (a *AIAgent) GetConversationHistory() []openai.ChatCompletionMessage {
	return a.conversation
}

func (a *AIAgent) ClearHistory() {
	a.conversation = []openai.ChatCompletionMessage{}
	a.initializeSystemPrompt()
}

func convertToolsToOpenAI(tools []tools.Tool) []openai.Tool {
	result := make([]openai.Tool, len(tools))
	for i, tool := range tools {
		result[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		}
	}
	return result
}
