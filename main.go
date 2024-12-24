package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Функция для создания директории, если её нет
func ensureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// Функция для скачивания одного файла
func downloadFile(targetURL, savePath string) error {
	resp, err := http.Get(targetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// Функция для извлечения ссылок из HTML-контента
func extractLinks(baseURL string, content string) ([]string, error) {
	var links []string
	linkRegex := regexp.MustCompile(`href=\"(.*?)\"`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		link := match[1]
		parsedLink, err := url.Parse(link)
		if err != nil {
			continue
		}

		base, err := url.Parse(baseURL)
		if err != nil {
			continue
		}

		resolvedLink := base.ResolveReference(parsedLink).String()
		if strings.HasPrefix(resolvedLink, base.Scheme+"://"+base.Host) {
			links = append(links, resolvedLink)
		}
	}

	return links, nil
}

// Функция для рекурсивного скачивания сайта
func downloadSite(targetURL, saveDir string, visited map[string]bool) error {
	if visited[targetURL] {
		return nil
	}
	visited[targetURL] = true

	resp, err := http.Get(targetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("не удалось загрузить  %s: %s", targetURL, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	localPath := filepath.Join(saveDir, parsedURL.Path)
	if strings.HasSuffix(localPath, "/") {
		localPath = filepath.Join(localPath, "index.html")
	}

	err = ensureDir(filepath.Dir(localPath))
	if err != nil {
		return err
	}

	err = os.WriteFile(localPath, body, 0644)
	if err != nil {
		return err
	}

	links, err := extractLinks(targetURL, string(body))
	if err != nil {
		return err
	}

	for _, link := range links {
		err := downloadSite(link, saveDir, visited)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", link, err)
		}
	}

	return nil
}

// Основная функция
func main() {
	if len(os.Args) < 3 {
		fmt.Println("Использование: go run main.go <url> <директория>")
		return
	}

	targetURL := os.Args[1]
	saveDir := os.Args[2]

	err := ensureDir(saveDir)
	if err != nil {
		fmt.Printf("Ошибка создания директории: %v\n", err)
		return
	}

	visited := make(map[string]bool)
	err = downloadSite(targetURL, saveDir, visited)
	if err != nil {
		fmt.Printf("Ошибка загрузки сайта: %v\n", err)
		return
	}

	fmt.Println("Сайт успешно загружен.")
}
