package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"ai-browser-agent/agent"
	"ai-browser-agent/browser"
)

func main() {
	fmt.Println(" AI Browser Agent запущен!")
	fmt.Println("Введите задачу для агента (или 'quit' для выхода)")
	fmt.Println()

	// Инициализация браузера
	browserManager, err := browser.NewBrowserManager()
	if err != nil {
		fmt.Printf("Ошибка инициализации браузера: %v\n", err)
		return
	}
	defer browserManager.Close()

	// Инициализация AI агента
	aiAgent, err := agent.NewAIAgent(browserManager)
	if err != nil {
		fmt.Printf("Ошибка инициализации AI агента: %v\n", err)
		return
	}

	// Интерактивный цикл
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n Ваша задача: ")
		if !scanner.Scan() {
			break
		}

		task := strings.TrimSpace(scanner.Text())
		if task == "" {
			continue
		}

		if strings.ToLower(task) == "quit" || strings.ToLower(task) == "exit" {
			fmt.Println("До свидания!")
			break
		}

		fmt.Printf("\n Агент выполняет задачу: %s\n\n", task)

		// Выполнение задачи агентом
		result, err := aiAgent.ExecuteTask(task)
		if err != nil {
			fmt.Printf(" Ошибка выполнения задачи: %v\n", err)
		} else {
			fmt.Printf("\n Результат: %s\n", result)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf(" Ошибка чтения ввода: %v\n", err)
	}
}
