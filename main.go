package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type StoryContent struct {
	Heading  string `json:"heading"`
	Subtitle string `json:"subtitle"`
	Video    string `json:"video"`
	Map      string `json:"map"`
	MapPage  string `json:"mapPage"`
}

func storyContent() StoryContent {
	return StoryContent{
		Heading:  "走进四川",
		Subtitle: "从一段影像开始，穿过山川与世界相遇。",
		Video:    "/static/media/intro.mp4",
		Map:      "/static/media/sichuan-map.png",
		MapPage:  "/map/sichuan",
	}
}

const (
	defaultPort     = "8080"   //默认监听
	templatePattern = ""       //匹配的前端
	staticDir       = "static" //静态匹配

)

type PageConfig struct {
	PageTitle    string     `json:"PageTitle"`    //最大标题
	VideoTitle   string     `json:"VideoTitle"`   //视频上的标题
	VideoURL     string     `json:"VideoUrl"`     //视频地址
	PosterURL    string     `json:"PosterUrl"`    //视频封面地址
	SideImageURL string     `json:"SideImageUrl"` //图片地址
	MapImageURL  string     `json:"MapImageUrl"`  //地图地址
	MapPointURL  []MapPoint `json:"MapPointUrl"`  //地图点击地址
	CurrentPlace MapPoint   `json:"CurrentPlace"` //对应的地点
}

// template是一个Go的包
// 第 32 行修改为：
var pageTemplates = template.Must(template.ParseFiles("index.html", "about.html", "template.html")) // 保存已经加载的 HTML 模板。
func main() {
	mux1 := http.NewServeMux()
	registerPandaScrollRoutes(mux1)
	server := &http.Server{
		Addr:        "8080",
		Handler:     withRequestLog(mux1), //日志中间件包裹
		ReadTimeout: 5 * time.Second,
	}
	port := os.Getenv("PORT") //读取环境变量
	if port == "" {
		port = defaultPort //默认
	}
	var err error
	//解析
	pageTemplates, err = template.ParseGlob("templates/*")
	if err != nil {
		log.Fatal("加载 HTML 模板失败！", err)
	}
	staticRoot, err := filepath.Abs(staticDir)
	//转化为绝对路径
	if err != nil {
		log.Fatal("解析静态资源目录失败！", err)
	}
	if _, err := os.Stat(staticRoot); err != nil {
		log.Fatal("静态资源不存在:%s,错误：%V", staticRoot, err)
		//目录不存在停止启动
	}
	mux := http.NewServeMux() //创建路由分发器
	//托管静态资源且支持视频Range请求
	// http.StripPrefix去除URL路径钱买你没用的部分
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticRoot))))
	mux.HandleFunc("/api/health", healthHandler)          //注册健康检查接口
	mux.HandleFunc("/api/config", configHandler)          // 注册页面配置接口。
	mux.HandleFunc("/api/map-points", mapPointsHandler)   //注册地图地点接口
	mux.HandleFunc("/go/place/", placeRedirectHandler)    //注册点击跳转接口
	mux.HandleFunc("/place/", placePageHandler)           //注册详情网页接口
	mux.HandleFunc("/go/sichuan", sichuanRedirectHandler) // 注册兼容旧入口的四川跳转接口
	mux.HandleFunc("/", homeHandler)                      //注册首页
	server = &http.Server{ //创建服务器实例
		Addr:              ":" + port,
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,  // 限制读取请求头的时间。
		ReadTimeout:       15 * time.Second, // 限制读取请求的时间。
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second, // 限制空闲连接保持时间。
	}
	log.Println("服务器已经启动：http://localhost:%s", port)
	if err1 := server.ListenAndServe(); err1 != nil && err1 != http.ErrServerClosed {
		log.Fatal("服务器启动失败！%v", err1)
	}
}
func logRequests(next *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Println("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// 展示网页
func pageData() PageConfig {
	return PageConfig{
		PageTitle:    "",
		VideoTitle:   "",
		VideoURL:     "",
		PosterURL:    "",
		SideImageURL: "",
		MapImageURL:  "",
		MapPointURL:  mapPoints,
	}
}

// 首页1home.html没写
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r) //返回404
		return
	}
	
	if r.Method != http.MethodGet {
		//返回405
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}
	renderTemplate(w, "home.html", pageData()) //渲染
}
func renderTemplate(w http.ResponseWriter, name string, data PageConfig) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//响应类型为 HTML
	if err := pageTemplates.ExcuteTemplate(w, name, data); err != nil {
		log.Println("渲染模板%s失败%v", name, err)
	}

}

// 首页2
func configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		//返回405
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, pageData())
}

// 首页3
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		//返回405
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339)}) //返回服务器时间
}
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Println("写入JSON响应失败！%v", err)
	}
}

// 首页4
func sichuanRedirectHandler(w http.ResponseWriter, r *http.Request) { // 定义兼容跳转接口。
	if r.Method != http.MethodGet {
		http.Error(w, "只允许 GET 请求", http.StatusMethodNotAllowed)
		// 返回 405。
		return
	}

	name := r.URL.Query().Get("id")
	if name == "" {
		// 如果没有传地点 ID。
		http.Redirect(w, r, "/#map", http.StatusFound)
		// 回到首页地图区域。
		return
	}

	point, ok := findMapPoint(name)
	if !ok {
		http.NotFound(w, r) // 返回 404。
		return
	}
	http.Redirect(w, r, "/place/"+point.ID+"/", http.StatusFound) // 跳转到对应地点详情页。
}

func findMapPoint(ID string) (MapPoint, bool) {
	for _, point := range mapPoints {
		if point.ID == ID {
			return point, true
		}
	}
	return MapPoint{}, false
}
