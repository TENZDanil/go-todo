package browser

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type BrowserManager struct {
	browser *rod.Browser
	page    *rod.Page
	ctx     context.Context
}

func NewBrowserManager() (*BrowserManager, error) {
	launcher := launcher.New().
		Headless(true).
		NoSandbox(true)

	url, err := launcher.Launch()
	if err != nil {
		return nil, fmt.Errorf("не удалось запустить браузер: %w", err)
	}

	browser := rod.New().ControlURL(url).MustConnect()
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к браузеру: %w", err)
	}

	page := browser.MustPage()

	return &BrowserManager{
		browser: browser,
		page:    page,
		ctx:     context.Background(),
	}, nil
}

func (bm *BrowserManager) Navigate(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	err := bm.page.Timeout(30 * time.Second).Navigate(url)
	if err != nil {
		return fmt.Errorf("не удалось загрузить страницу %s: %w", url, err)
	}

	err = bm.page.WaitLoad()
	if err != nil {
		return fmt.Errorf("не удалось дождаться загрузки страницы %s: %w", url, err)
	}

	time.Sleep(2 * time.Second)
	return nil
}

func (bm *BrowserManager) GetPageContent() (string, error) {
	html, err := bm.page.HTML()
	if err != nil {
		return "", fmt.Errorf("не удалось получить HTML: %w", err)
	}
	return html, nil
}

func (bm *BrowserManager) GetPageText() (string, error) {
	text, err := bm.page.MustElement("body").Text()
	if err != nil {
		return "", fmt.Errorf("не удалось получить текст страницы: %w", err)
	}
	return text, nil
}

func (bm *BrowserManager) GetPageURL() string {
	return bm.page.MustInfo().URL
}

func (bm *BrowserManager) GetPageTitle() string {
	return bm.page.MustInfo().Title
}

func (bm *BrowserManager) ClickElement(selector string) error {
	element, err := bm.page.Timeout(10 * time.Second).Element(selector)
	if err != nil {
		return fmt.Errorf("не удалось найти элемент с селектором %s: %w", selector, err)
	}

	element.MustClick()
	time.Sleep(1 * time.Second)
	return nil
}

func (bm *BrowserManager) FillInput(selector string, text string) error {
	element, err := bm.page.Timeout(10 * time.Second).Element(selector)
	if err != nil {
		return fmt.Errorf("не удалось найти поле ввода с селектором %s: %w", selector, err)
	}

	err = element.Input(text)
	if err != nil {
		return fmt.Errorf("не удалось ввести текст в поле: %w", err)
	}

	return nil
}

func (bm *BrowserManager) GetElements(selector string) ([]ElementInfo, error) {
	elements, err := bm.page.Timeout(10 * time.Second).Elements(selector)
	if err != nil {
		return nil, fmt.Errorf("не удалось найти элементы с селектором %s: %w", selector, err)
	}

	var result []ElementInfo
	for i, elem := range elements {
		text, _ := elem.Text()

		tag := ""
		tagResult, err := elem.Eval(`() => this.tagName.toLowerCase()`)
		if err == nil && tagResult != nil {
			tagStr := tagResult.Value.String()
			tagStr = strings.Trim(tagStr, `"`)
			if tagStr != "" {
				tag = tagStr
			}
		}

		href, _ := elem.Attribute("href")
		id, _ := elem.Attribute("id")
		class, _ := elem.Attribute("class")
		visible, _ := elem.Visible()

		uniqueSelector := bm.generateSelector(elem, selector, i)

		result = append(result, ElementInfo{
			Selector: uniqueSelector,
			Tag:      tag,
			Text:     strings.TrimSpace(text),
			Href:     href,
			ID:       id,
			Class:    class,
			Visible:  visible,
			Index:    i,
		})
	}

	return result, nil
}

type ElementInfo struct {
	Selector string
	Tag      string
	Text     string
	Href     *string
	ID       *string
	Class    *string
	Visible  bool
	Index    int
}

func (bm *BrowserManager) generateSelector(elem *rod.Element, baseSelector string, index int) string {
	if id, err := elem.Attribute("id"); err == nil && id != nil && *id != "" {
		return fmt.Sprintf("#%s", *id)
	}

	if dataID, err := elem.Attribute("data-id"); err == nil && dataID != nil && *dataID != "" {
		return fmt.Sprintf("[data-id='%s']", *dataID)
	}

	return fmt.Sprintf("%s:nth-child(%d)", baseSelector, index+1)
}

func (bm *BrowserManager) Screenshot(path string) error {
	_, err := bm.page.Screenshot(false, nil)
	if err != nil {
		return fmt.Errorf("не удалось сделать скриншот: %w", err)
	}
	return nil
}

func (bm *BrowserManager) WaitForElement(selector string, timeout time.Duration) error {
	_, err := bm.page.Timeout(timeout).Element(selector)
	return err
}

func (bm *BrowserManager) ExecuteJavaScript(js string) (interface{}, error) {
	result, err := bm.page.Eval(js)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения JavaScript: %w", err)
	}
	return result.Value, nil
}

func (bm *BrowserManager) GetVisibleElements() ([]ElementInfo, error) {
	js := `
	() => {
		const elements = [];
		const selectors = ['a', 'button', 'input', 'textarea', 'select', '[onclick]', '[role="button"]'];
		
		selectors.forEach(selector => {
			document.querySelectorAll(selector).forEach((el, index) => {
				const rect = el.getBoundingClientRect();
				if (rect.width > 0 && rect.height > 0 && 
					window.getComputedStyle(el).visibility !== 'hidden' &&
					window.getComputedStyle(el).display !== 'none') {
					
					let selector = '';
					if (el.id) {
						selector = '#' + el.id;
					} else if (el.className) {
						selector = '.' + el.className.split(' ')[0];
					} else {
						selector = el.tagName.toLowerCase();
					}
					
					elements.push({
						selector: selector,
						tag: el.tagName.toLowerCase(),
						text: el.textContent.trim().substring(0, 100),
						href: el.href || null,
						id: el.id || null,
						class: el.className || null,
						visible: true,
						index: index
					});
				}
			});
		});
		
		return elements;
	}
	`

	_, err := bm.ExecuteJavaScript(js)
	if err != nil {
		return nil, err
	}

	return []ElementInfo{}, nil
}

func (bm *BrowserManager) Close() {
	if bm.page != nil {
		bm.page.Close()
	}
	if bm.browser != nil {
		bm.browser.Close()
	}
}
