package handler

import (
	"bytes"
	"html/template"
	"net/http"
)

const metricsHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="utf-8">
    <title>Метрики</title>
</head>
<body>
    <h1>Метрики сервера</h1>
    <table>
        <thead>
            <tr>
                <th>Тип</th>
                <th>Имя</th>
                <th>Значение</th>
            </tr>
        </thead>
        <tbody>
            {{range .}}
            <tr>
                <td>{{.MetricType}}</td>
                <td>{{.MetricName}}</td>
                <td>{{.Value}}</td>
            </tr>
            {{else}}
            <tr>
                <td colspan="3">Метрик пока нет</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</body>
</html>`

var metricsTemplate = template.Must(template.New("metrics").Parse(metricsHTML))

func HTMLHandler(rw http.ResponseWriter, r *http.Request, metricStorage MetricStorage) {
	var page bytes.Buffer
	if err := metricsTemplate.Execute(&page, metricStorage.GetAllMetrics()); err != nil {
		http.Error(rw, "Failed to render metrics", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = rw.Write(page.Bytes())
}
