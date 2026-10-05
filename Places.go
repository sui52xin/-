package main

import (
	"net/http"
	"strings"
)

type MapPoint struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Summary     string  `json:"summary"`     //地点简介
	Description string  `json:"description"` //地点详情
}

// 放可以点击的点
var mapPoints = []MapPoint{
	{
		ID:          "chengdu",
		Name:        "成都",
		X:           42.5,
		Y:           54.0,
		Summary:     "",
		Description: "",
	},
	{
		ID:          "aiba",
		Name:        "阿坝",
		X:           50.0,
		Y:           40.0,
		Summary:     "",
		Description: "",
	},
	{
		ID:          "emeishan",
		Name:        "峨眉山",
		X:           47.0,
		Y:           64.0,
		Summary:     "",
		Description: "",
	},
	{
		ID:          "daochengyading",
		Name:        "稻城亚丁",
		X:           20.0,
		Y:           66.0,
		Summary:     "",
		Description: "",
	},
}

func mapPointsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只允许 GET 请求", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, mapPoints)
}
func placeRedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只允许 GET 请求", http.StatusMethodNotAllowed)
		return
	}
	//r.URL.Path获取当前 HTTP 请求的 URL 路径部分，即你要访问的资源，网页等
	id := extractPlaceID(r.URL.Path, "/go/place/") //提取ID
	point, ok := findMapPoint(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	//StatusNotFound 是 404，
	http.Redirect(w, r, "/place/"+point.ID+"/", http.StatusFound)
}
func placePageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只允许 GET 请求", http.StatusMethodNotAllowed)
		return
	}
	id := extractPlaceID(r.URL.Path, "/place/")
	point, ok := findMapPoint(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	data := pageData()
	data.CurrentPlace = point
	renderTemplate(w, "place.html", data)
}
func extractPlaceID(path string, prefix string) string {
	return strings.Trim(strings.TrimPrefix(path, prefix), "/")
}
