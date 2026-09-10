package render

import (
	"bytes"
	"fmt"
	"html/template"
	"le-grimoire/internal/library"
	"le-grimoire/internal/state"
)

const baseCSS = `
* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
    -webkit-font-smoothing: none;
}
body {
    width: 400px;
    background-color: #ffffff;
    color: #000000;
    font-family: 'Literata', 'Georgia', serif;
}
.header {
    width: 100%;
    height: 24px;
    border-bottom: 2px solid #000000;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 8px;
    font-family: sans-serif;
    font-size: 12px;
    font-weight: bold;
}
.list-container {
    padding: 6px 8px;
}
.list-item {
    display: flex;
    align-items: center;
    padding: 4px 6px;
    margin-bottom: 2px;
    font-family: sans-serif;
    font-size: 14px;
    line-height: 18px;
}
.list-item.selected {
    background-color: #000000;
    color: #ffffff;
    font-weight: bold;
}
.reader-content {
    width: 400px;
    padding: 12px 14px;
    font-size: 16px;
    line-height: 24px;
    text-align: justify;
    word-break: break-word;
}
.reader-content p {
    margin-bottom: 12px;
    text-indent: 16px;
}
.reader-content img {
    max-width: 100%;
    height: auto;
    display: block;
    margin: 8px auto;
}
`

var (
	listTemplate = template.Must(template.New("list").Parse(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>` + baseCSS + `</style>
</head>
<body>
    <div class="header">
        <span>{{.Title}}</span>
        <span>{{.PageIndicator}}</span>
    </div>
    <div class="list-container">
        {{range $index, $item := .Items}}
        <div class="list-item {{if eq $index $.SelectedIndex}}selected{{end}}">
            {{$item}}
        </div>
        {{end}}
    </div>
</body>
</html>`))

	readerTemplate = template.Must(template.New("reader").Parse(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Literata:ital,opsz,wght@0,7..72,400;0,7..72,700;1,7..72,400&display=swap" rel="stylesheet">
    <style>` + baseCSS + `</style>
</head>
<body>
    <div class="reader-content" id="content">
        {{.ContentHTML}}
    </div>
</body>
</html>`))
)

type ListViewModel struct {
	Title         string
	PageIndicator string
	Items         []string
	SelectedIndex int
}

func BuildLibraryHTML(libraries []library.Library, cursor int) (string, error) {
	items := make([]string, len(libraries))
	for i, l := range libraries {
		items[i] = l.Name
	}
	return executeListTemplate("Libraries", items, cursor)
}

func BuildSeriesHTML(seriesList []library.Book, cursor int) (string, error) {
	items := make([]string, len(seriesList))
	for i, s := range seriesList {
		items[i] = s.Title
	}
	return executeListTemplate("Series", items, cursor)
}

func BuildBookListHTML(chapters []library.Chapter, cursor int) (string, error) {
	items := make([]string, len(chapters))
	for i, c := range chapters {
		items[i] = c.Title
	}
	return executeListTemplate("Chapters", items, cursor)
}

func BuildReaderHTML(chapterHTML string) (string, error) {
	var buf bytes.Buffer
	err := readerTemplate.Execute(&buf, map[string]template.HTML{
		"ContentHTML": template.HTML(chapterHTML),
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func BuildPlaceholderHTML(p state.Page) string {
	return fmt.Sprintf("<!DOCTYPE html><html><body style='width:400px;font-family:sans-serif;padding:20px;'><h1>%s</h1><p>Cursor: %d</p></body></html>", p.Type, p.State["cursor"])
}

func executeListTemplate(title string, items []string, cursor int) (string, error) {
	const pageSize = 10
	total := len(items)
	if total == 0 {
		items = []string{"(Empty)"}
	}

	start := (cursor / pageSize) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	visibleItems := items[start:end]
	selectedIndex := cursor - start

	var pageIndicator string
	if total > 0 {
		totalPages := (total + pageSize - 1) / pageSize
		currentPage := (cursor / pageSize) + 1
		pageIndicator = fmt.Sprintf("%d/%d", currentPage, totalPages)
	}

	data := ListViewModel{
		Title:         title,
		PageIndicator: pageIndicator,
		Items:         visibleItems,
		SelectedIndex: selectedIndex,
	}

	var buf bytes.Buffer
	if err := listTemplate.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
