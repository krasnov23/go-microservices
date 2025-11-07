package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 //Устанавливаем ограничение на размер тела запроса — 1 МБ (1048576 байт).

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes)) // Оборачиваем тело запроса специальным ридером, который автоматически прерывает чтение, если размер превышает 1 МБ.

	dec := json.NewDecoder(r.Body) // Создаём новый JSON-декодер, который будет читать тело запроса (r.Body) и преобразовывать JSON → Go-структуру.

	err := dec.Decode(data) // Пытаемся распарсить JSON в переменную data.

	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{}) // Пробуем декодировать ещё один JSON-объект. Чтобы убедиться, что в теле только один JSON, а не несколько подряд.

	// Попадаем в ошибку если не конец файла, то есть второй json будет распарсен без ошибок и будет <nil>
	if err != io.EOF {
		return errors.New("body must have only a single json value")
	}

	return nil
}

func (app *Config) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data) // Преобразуем data в JSON-строку (байты).

	if err != nil {
		return err
	}

	// Если переданы заголовки (headers), добавляем их к HTTP-ответу.
	// headers — это срез, но обычно передают один элемент, поэтому используется headers[0].
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	// Устанавливаем тип содержимого — application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Отправляем сам JSON-текст клиенту.
	_, err = w.Write(out)

	if err != nil {
		return err
	}

	return nil
}

func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	// Если пользователь передал другой статус — используем его.
	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)

}
