package main

type MapPoint struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Summary     string  `json:"summary"`     //地点简介
	Description string  `json:"description"` //地点详情
}

var mapPoints []MapPoint
