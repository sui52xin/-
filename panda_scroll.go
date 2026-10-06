package main

import (
	"fmt"
	"html/template"
	"math"
	"strings"
)

// 熊猫的位置
type PandaShot struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Subtitle string  `json:"subtitle"`
	Effect   string  `json:"effect"`    //特效名
	PandaPos float64 `json:"panda_pos"` //位置：0-最左，1-最右
	Accent   string  `json:"accent"`    //强调色
}

var allowedEffects = map[string]struct{}{
	"rise":       {}, //下方升起
	"slide-left": {}, //右侧划入
	"zoom-in":    {}, //由远及近放大
	"blur-in":    {}, //模糊变清晰
	"tilt-in":    {}, //带倾斜角进入
	"split-in":   {}, //从右往左展开
}

func pandaTimeline() []PandaShot {
	return []PandaShot{
		{ID: "bamboo", Title: "熊猫从竹林醒来", Subtitle: "向下滚动鼠标，它就会从画面左侧出发。", Effect: "rise", PandaPos: 0.00, Accent: "#3f7d4f"},
		{ID: "river", Title: "顺着岷江往前走", Subtitle: "水汽升起来，标题从侧面滑入画面。", Effect: "slide-left", PandaPos: 0.20, Accent: "#2f7f8f"},
		{ID: "teahouse", Title: "在茶馆坐一小会儿", Subtitle: "盖碗茶端上桌，文字由远及近地放大。", Effect: "zoom-in", PandaPos: 0.40, Accent: "#b4552f"},
		{ID: "opera", Title: "脸谱与三弦同时登场", Subtitle: "锣鼓响起，标题像被灯光慢慢照亮。", Effect: "blur-in", PandaPos: 0.60, Accent: "#a0711f"},
		{ID: "snow", Title: "翻过雪山抵达高原", Subtitle: "风把经幡吹斜，标题也跟着倾斜入场。", Effect: "tilt-in", PandaPos: 0.80, Accent: "#5b6b7a"},
		{ID: "home", Title: "再回到出发的地方", Subtitle: "长卷收尾，文字从右向左展开。", Effect: "split-in", PandaPos: 1.00, Accent: "#3f7d4f"},
	}
}

// 检查特效的合法性
func validatePandaTimeline(shots []PandaShot) error {
	if len(shots) < 2 {
		return fmt.Errorf("滑动的标题太少！")
	}
	for i, shot := range shots {
		if strings.TrimSpace(shot.ID) == "" {
			return fmt.Errorf("第%d个缺少ID", i+1)
		}
		if _, ok := allowedEffects[shot.Effect]; !ok {
			return fmt.Errorf("第%d个%q特效不在白名单里", i+1, shot.Effect)
		}
		if shot.PandaPos < 0 || shot.PandaPos > 1 {
			return fmt.Errorf("第%d个的范围超出0~1", i+1)
		}
	}
	return nil
}

// 某个滚动进度下，后端算出来的动画帧"。
type PandaFrame struct {
	Progress float64 `json:"progress"` //当前滚动进度
	Index    int     `json:"index"`    //第几个
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Subtitle string  `json:"subtitle"`
	Effect   string  `json:"effect"`
	PandaPos float64 `json:"panda_pos"`
	Local    float64 `json:"local"` //局部进度0~1
	Accent   string  `json:"accent"`
}

// 它先做 clamp（把 t 限制在 0~1），再用 3t²-2t³ 这条经典曲线平滑过渡。
func smoothstep(t float64) float64 {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return t * t * (3 - 2*t)
}

// resolvePandaFrame 把"滚动进度"映射成"熊猫位置 + 该显示哪一幕"。
//如果不用插值：
//- 直接跳到下一个关键帧 → 熊猫会一格一格地「闪」，不是滑动

func resolvePandaFrame(progress float64) PandaFrame {
	shots := pandaTimeline()
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	last := len(shots) - 1
	scaled := progress * float64(last) //放大为0~last的坐标
	idx := int(math.Floor(scaled))     //向下取整数
	if idx >= last {
		idx = last - 1
	}
	local := smoothstep(scaled - float64(idx))
	from := shots[idx]
	to := shots[idx+1]
	return PandaFrame{
		Progress: progress,
		Index:    idx,
		ID:       from.ID,
		Title:    from.Title,
		Subtitle: from.Subtitle,
		Effect:   from.Effect,
		PandaPos: from.PandaPos + (to.PandaPos-from.PandaPos)*local,
		//线性插值实现滑动
		Local:  local,
		Accent: from.Accent,
	}

}

var pandaScrollTemplate = template.Must(
	template.New("panda-scroll.html").
		Funcs(template.FuncMap{"inc": func(n int) int { return n + 1 }}).
		ParseFiles("./templates/panda.html"),
)
