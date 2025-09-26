package main

import (
    "io"       // Чтение тела HTTP-запроса полностью в память
    "log"      // Логирование событий сервера и ошибок
    "net/http" // Базовый HTTP-сервер, маршрутизация и ответы
    "os"       // Работа с файловой системой (создание/запись в data.txt)
)

// withCORS — простой middleware, добавляющий к ответам CORS-заголовки,
// чтобы фронтенд мог обращаться к нашему API из браузера.
func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Разрешаем запросы с любых источников (для учебной работы это допустимо)
		w.Header().Set("Access-Control-Allow-Origin", "*")
        // Разрешаем методы (минимально достаточно POST для нашей задачи)
        w.Header().Set("Access-Control-Allow-Methods", "POST")
        // Разрешаем заголовок Content-Type
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        // Просто передаём управление хендлеру
		h.ServeHTTP(w, r)
	})
}

// submitHandler — обрабатывает POST /submit. Принимает тело запроса как текст,
// валидирует, дописывает строку с временной меткой в файл data.txt и отвечает
// JSON-объектом {"status":"ok"} при успехе.
func submitHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем только POST. Любой другой метод — 405 Method Not Allowed
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Читаем тело запроса целиком
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Простейшая валидация: тело не должно быть пустым
	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// Открываем файл для дозаписи. Флаги:
	// - os.O_CREATE — создать файл, если он отсутствует
	// - os.O_WRONLY — открываем на запись
	// - os.O_APPEND — дописываем в конец файла
	file, err := os.OpenFile("data.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("open file error: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

    // Записываем только переданный текст + перевод строки
    if _, err := file.WriteString(string(body) + "\n"); err != nil {
		log.Printf("write error: %v", err)
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ в формате JSON
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// main — точка входа приложения. Создаём маршрутизатор, регистрируем
// обработчики и запускаем HTTP-сервер на порту 8080. Маршрутизатор оборачиваем
// в middleware withCORS, чтобы ко всем ответам добавлялись нужные заголовки.
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/submit", submitHandler)

	addr := ":8080" // адрес прослушивания (порт 8080 на всех интерфейсах)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
