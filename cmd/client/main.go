package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	fmt.Println("Введите длинный URL")
	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Не удалось считать длинный URL: %+v", err)
		return
	}
	long = strings.TrimSuffix(long, "\n")

	if len(long) == 0 {
		fmt.Println("Длинный URL не может быть пустым")
		return
	}

	reqData := url.Values{}
	reqData.Set("url", long)

	// todo: вынести во флаги
	endpoint := "http://localhost:8080/"
	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqData.Encode()))
	if err != nil {
		fmt.Printf("Не удалось создать запрос: %+v", err)
		return
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("ошибка выполнения запроса: %+v", err)
		return
	}

	fmt.Println("Статус-код ", response.Status)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Ошибка чтения тела ответа: %+v", err)
		return
	}
	fmt.Println(string(body))
}
