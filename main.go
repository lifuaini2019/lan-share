package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
)

// 全局变量
var (
	clients    = make(map[*websocket.Conn]bool)
	clientsMux = sync.RWMutex{}
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许跨域
		},
	}
)

// 数据结构
type Message struct {
	Time    string `json:"time"`
	Content string `json:"content"`
}

type Template struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Category struct {
	Icon      string     `json:"icon"`
	Name      string     `json:"name"`
	Templates []Template `json:"templates"`
}

type TemplatesConfig struct {
	Categories map[string]Category `json:"categories"`
}

type FileInfo struct {
	FileID   string  `json:"file_id"`
	Filename string  `json:"filename"`
	Size     int64   `json:"size"`
	SizeMB   float64 `json:"size_mb"`
	Type     string  `json:"type"`
	Data     string  `json:"data"`
	SenderIP string  `json:"sender_ip"`
	SendTime string  `json:"send_time"`
	Action   string  `json:"action"`
}

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// 常量
const (
	DataFile      = "messages.txt"
	TemplatesFile = "templates_config.json"
	Port          = 9405
)

var allowedExtensions = map[string]bool{
	// 文本文档
	"txt": true, "md": true, "markdown": true, "rtf": true,
	// PDF文档
	"pdf": true,
	// 图片格式
	"png": true, "jpg": true, "jpeg": true, "gif": true, "bmp": true, "webp": true, "svg": true,
	// Office文档
	"doc": true, "docx": true, "xls": true, "xlsx": true, "ppt": true, "pptx": true,
	// 代码文件
	"html": true, "htm": true, "css": true, "js": true, "json": true, "xml": true,
	"py": true, "go": true, "java": true, "cpp": true, "c": true, "h": true,
	// 配置文件
	"ini": true, "cfg": true, "conf": true, "yaml": true, "yml": true, "toml": true,
	// 压缩文件
	"zip": true, "rar": true, "7z": true, "tar": true, "gz": true,
	// 其他常用格式
	"csv": true, "log": true, "sql": true, "sh": true, "bat": true,
}

// 工具函数
func getLocalIP() string {
	// 方法1：尝试通过连接外部服务器获取本地IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		ip := localAddr.IP.String()
		// 如果获取到的不是回环地址，且是私有IP，则使用
		if ip != "127.0.0.1" && (strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.")) {
			return ip
		}
	}

	// 方法2：遍历网络接口获取私有IP
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip := ipnet.IP.String()
					// 优先返回192.168.x.x的IP
					if strings.HasPrefix(ip, "192.168.") {
						return ip
					}
					// 其次是10.x.x.x和172.16-31.x.x
					if strings.HasPrefix(ip, "10.") ||
						(strings.HasPrefix(ip, "172.") && len(strings.Split(ip, ".")) >= 2) {
						parts := strings.Split(ip, ".")
						if len(parts) >= 2 {
							second, _ := strconv.Atoi(parts[1])
							if second >= 16 && second <= 31 {
								return ip
							}
						}
					}
				}
			}
		}

		// 如果没找到私有IP，再次遍历返回任意非回环的IPv4地址
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ip := ipnet.IP.String()
					if ip != "127.0.0.1" {
						return ip
					}
				}
			}
		}
	}

	// 方法3：尝试通过命令行工具获取IP（仅在Linux系统上）
	if runtime.GOOS == "linux" {
		cmd := exec.Command("hostname", "-I")
		output, err := cmd.Output()
		if err == nil {
			ips := strings.Fields(strings.TrimSpace(string(output)))
			for _, ip := range ips {
				// 过滤IPv6地址和回环地址
				if net.ParseIP(ip) != nil && !strings.Contains(ip, ":") && ip != "127.0.0.1" {
					return ip
				}
			}
		}

		// 尝试使用ip route命令
		cmd = exec.Command("ip", "route", "get", "8.8.8.8")
		output, err = cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "src") {
					fields := strings.Fields(line)
					for i, field := range fields {
						if field == "src" && i+1 < len(fields) {
							ip := fields[i+1]
							if net.ParseIP(ip) != nil && ip != "127.0.0.1" {
								return ip
							}
						}
					}
				}
			}
		}
	}

	// 最后的选择：返回回环地址
	log.Printf("⚠️ 无法获取局域网IP，使用回环地址")
	return "127.0.0.1"
}

func generateQRCode(r *http.Request) (string, string, bool) {
	// 检测是否为域名访问
	host := r.Host
	isIPAccess := false

	// 提取主机名（去除端口）
	hostname := host
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		hostname = host[:colonIndex]
	}

	// 判断是否为IP地址访问
	if net.ParseIP(hostname) != nil {
		isIPAccess = true
	}

	// 根据访问类型选择协议
	var protocol string
	var url string
	if isIPAccess {
		// IP访问使用http
		protocol = "http"
		url = fmt.Sprintf("%s://%s", protocol, host)
	} else {
		// 域名访问使用https
		protocol = "https"
		url = fmt.Sprintf("%s://%s", protocol, host)
	}

	log.Printf("🔄 生成二维码URL: %s (IP访问: %v)", url, isIPAccess)

	// 使用与 Flask 版本相同的配置
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		log.Printf("❌ 创建二维码失败: %v", err)
		return "", url, isIPAccess
	}

	// 生成PNG格式，使用 512x512 尺寸
	pngData, err := qr.PNG(512)
	if err != nil {
		log.Printf("❌ 生成二维码PNG失败: %v", err)
		return "", url, isIPAccess
	}

	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData)
	log.Printf("✅ 二维码生成成功！DataURL长度: %d", len(dataURL))
	return dataURL, url, isIPAccess
}

func loadMessages() ([]Message, error) {
	var messages []Message

	if _, err := os.Stat(DataFile); os.IsNotExist(err) {
		return messages, nil
	}

	data, err := os.ReadFile(DataFile)
	if err != nil {
		return nil, err
	}

	// 改用JSON格式存储，避免换行符问题
	if len(data) > 0 {
		// 尝试JSON格式解析
		err = json.Unmarshal(data, &messages)
		if err != nil {
			// 如果JSON解析失败，尝试旧格式解析（兼容性）
			log.Printf("⚠️ JSON解析失败，尝试旧格式: %v", err)
			lines := strings.Split(string(data), "\n")
			for lineNum, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				parts := strings.SplitN(line, "|", 2)
				if len(parts) == 2 {
					messages = append(messages, Message{
						Time:    parts[0],
						Content: parts[1],
					})
				} else {
					log.Printf("⚠️ 第%d行格式不正确，已跳过: %s...", lineNum+1, line[:min(50, len(line))])
				}
			}
		}
	}

	return messages, nil
}

func saveMessages(messages []Message) error {
	// 改用JSON格式存储，避免换行符问题
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(DataFile, data, 0644)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 创建默认模板配置
func createDefaultTemplates() TemplatesConfig {
	return TemplatesConfig{
		Categories: map[string]Category{
			"home": {
				Icon: "🏠",
				Name: "共享文字",
				Templates: []Template{
					{
						Title:   "欢迎使用",
						Content: "欢迎使用祖宇字文共享系统！您可以在这里快速分享文字内容。新的描述内容可以在这里写。",
					},
					{
						Title:   "使用提示",
						Content: "💡 小提示：电脑端按回车键可快速提交内容，支持多设备实时同步。",
					},
				},
			},
			"presale": {
				Icon: "💬",
				Name: "售前问题",
				Templates: []Template{
					{
						Title:   "在线客服",
						Content: "我们一直在线，如果没能第一时间回复您，说明我们正在忙碌中。有任何机器方面的问题直接咨询即可，我看到消息会马上回复您的。",
					},
				},
			},
			"express": {
				Icon:      "📦",
				Name:      "快递问题",
				Templates: []Template{},
			},
			"aftersale": {
				Icon: "🛠️",
				Name: "售后问题",
				Templates: []Template{
					{
						Title:   "保修说明",
						Content: "我们提供完善的保修服务：主机主板保修一年，打印头保修半年，电源附加线等保修一个月。",
					},
				},
			},
			"purchase": {
				Icon:      "🛒",
				Name:      "购买链接",
				Templates: []Template{},
			},
			"repair": {
				Icon:      "🔧",
				Name:      "维修问题",
				Templates: []Template{},
			},
			"settings": {
				Icon: "⚙️",
				Name: "系统设置",
				Templates: []Template{
					{
						Title:   "数据管理",
						Content: "使用下方的导入导出功能管理模板数据，支持全部或指定栏目的数据备份和恢复。",
					},
				},
			},
		},
	}
}

func ensureTemplatesFile() error {
	if _, err := os.Stat(TemplatesFile); os.IsNotExist(err) {
		log.Printf("⚠️ 模板文件 %s 不存在，正在创建默认配置...", TemplatesFile)
		defaultConfig := createDefaultTemplates()
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return err
		}

		err = os.WriteFile(TemplatesFile, data, 0644)
		if err != nil {
			return err
		}
		log.Printf("✅ 默认模板配置文件已创建: %s", TemplatesFile)
	}
	return nil
}

func loadTemplates() (TemplatesConfig, error) {
	var config TemplatesConfig

	data, err := os.ReadFile(TemplatesFile)
	if err != nil {
		log.Printf("⚠️ 模板文件 %s 不存在", TemplatesFile)
		return TemplatesConfig{Categories: make(map[string]Category)}, nil
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Printf("❌ 模板文件格式错误: %v", err)
		return TemplatesConfig{Categories: make(map[string]Category)}, nil
	}

	return config, nil
}

func saveTemplates(config TemplatesConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(TemplatesFile, data, 0644)
}

func allowedFile(filename string) bool {
	if !strings.Contains(filename, ".") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(filename)[1:])
	return allowedExtensions[ext]
}

// WebSocket 处理
func broadcastMessage(msgType string, data interface{}) {
	clientsMux.RLock()
	defer clientsMux.RUnlock()

	message := WebSocketMessage{
		Type: msgType,
		Data: data,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ 序列化消息失败: %v", err)
		return
	}

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, messageBytes)
		if err != nil {
			log.Printf("⚠️ WebSocket广播失败: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("❌ WebSocket连接升级失败: %v", err)
		return
	}
	defer conn.Close()

	clientsMux.Lock()
	clients[conn] = true
	clientsMux.Unlock()

	log.Println("✅ 新WebSocket客户端连接")

	// 发送连接确认
	conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"connected","data":{"message":"已连接到实时同步服务"}}`))

	// 处理客户端消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("❌ WebSocket读取消息失败: %v", err)
			break
		}

		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("❌ 解析WebSocket消息失败: %v", err)
			continue
		}

		// 处理不同类型的消息
		switch wsMsg.Type {
		case "request_sync":
			messages, err := loadMessages()
			if err != nil {
				log.Printf("❌ 同步数据错误: %v", err)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"sync_error","data":{"error":"同步失败"}}`))
			} else {
				// 创建消息副本并反转顺序，使最新的消息在数组前面
				messagesCopy := make([]Message, len(messages))
				for i, msg := range messages {
					messagesCopy[len(messages)-1-i] = msg
				}

				syncData := map[string]interface{}{
					"messages": messagesCopy,
				}
				syncMsg := WebSocketMessage{
					Type: "sync_data",
					Data: syncData,
				}
				syncBytes, _ := json.Marshal(syncMsg)
				conn.WriteMessage(websocket.TextMessage, syncBytes)
				log.Println("✅ 数据同步请求已处理")
			}
		}
	}

	// 清理连接
	clientsMux.Lock()
	delete(clients, conn)
	clientsMux.Unlock()
	log.Println("❌ WebSocket客户端断开连接")
}

// HTTP 路由处理函数
func indexHandler(c *gin.Context) {
	messages, err := loadMessages()
	if err != nil {
		log.Printf("❌ 加载消息失败: %v", err)
		messages = []Message{}
	}

	qrDataURL, serverURL, isIPAccess := generateQRCode(c.Request)
	log.Printf("🔍 传递给模板的二维码数据长度: %d", len(qrDataURL))
	log.Printf("🔍 传递给模板的服务器地址: %s", serverURL)

	// 判断网络类型
	var networkType string
	if isIPAccess {
		networkType = "局域网"
	} else {
		networkType = "广域网"
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"messages":     messages,
		"qr_data_url":  qrDataURL,
		"server_url":   serverURL,
		"network_type": networkType,
		"is_ip_access": isIPAccess,
	})
}

func addMessageHandler(c *gin.Context) {
	content := strings.TrimSpace(c.PostForm("content"))
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "内容不能为空"})
		return
	}

	messages, err := loadMessages()
	if err != nil {
		log.Printf("❌ 加载消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "加载消息失败"})
		return
	}

	timestamp := time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	newMessage := Message{
		Time:    timestamp,
		Content: content,
	}

	// 将新消息插入到开头而不是末尾，使其显示在最上面
	messages = append([]Message{newMessage}, messages...)
	err = saveMessages(messages)
	if err != nil {
		log.Printf("❌ 保存消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存消息失败"})
		return
	}

	// 广播新消息
	broadcastData := map[string]interface{}{
		"time":    timestamp,
		"content": content,
		"action":  "add",
	}
	broadcastMessage("new_message", broadcastData)
	log.Printf("✅ 消息已广播: %s", timestamp)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"time":    timestamp,
		"content": content,
	})
}

// 二维码API接口
func qrCodeHandler(c *gin.Context) {
	qrDataURL, serverURL, isIPAccess := generateQRCode(c.Request)

	// 判断网络类型
	var networkType string
	if isIPAccess {
		networkType = "局域网"
	} else {
		networkType = "广域网"
	}

	c.JSON(http.StatusOK, gin.H{
		"qr_data_url":  qrDataURL,
		"server_url":   serverURL,
		"network_type": networkType,
		"is_ip_access": isIPAccess,
	})
}

// 测试二维码生成
func testQRHandler(c *gin.Context) {
	qrDataURL, serverURL, isIPAccess := generateQRCode(c.Request)
	c.JSON(http.StatusOK, gin.H{
		"qr_data_url":  qrDataURL,
		"server_url":   serverURL,
		"qr_length":    len(qrDataURL),
		"is_ip_access": isIPAccess,
	})
}

// 局域网检测测试页面
func testLANHandler(c *gin.Context) {
	// 读取测试页面文件
	content, err := os.ReadFile("templates/test-lan-detection.html")
	if err != nil {
		c.String(http.StatusNotFound, "测试页面不存在")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// 域名检测测试页面
func testDomainHandler(c *gin.Context) {
	// 读取测试页面文件
	content, err := os.ReadFile("templates/test-domain.html")
	if err != nil {
		c.String(http.StatusNotFound, "域名测试页面不存在")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// 诊断工具页面
func diagnosticHandler(c *gin.Context) {
	// 读取诊断工具页面文件
	content, err := os.ReadFile("templates/diagnostic_tool.html")
	if err != nil {
		c.String(http.StatusNotFound, "诊断工具页面不存在")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// 调试检测页面
func debugDetectionHandler(c *gin.Context) {
	// 读取调试检测页面文件
	content, err := os.ReadFile("templates/debug_detection.html")
	if err != nil {
		c.String(http.StatusNotFound, "调试检测页面不存在")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// 高级调试页面
func advancedDebugHandler(c *gin.Context) {
	// 读取高级调试页面文件
	content, err := os.ReadFile("templates/advanced_debug.html")
	if err != nil {
		c.String(http.StatusNotFound, "高级调试页面不存在")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// Host头行为分析页面
func hostAnalysisHandler(c *gin.Context) {
	// 读取Host头行为分析页面文件
	content, err := os.ReadFile("templates/host_analysis.html")
	if err != nil {
		c.String(http.StatusNotFound, "Host头行为分析页面不存在")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// 智能检测帮助页面
func smartDetectionHelpHandler(c *gin.Context) {
	// 读取智能检测帮助页面文件
	content, err := os.ReadFile("templates/smart-detection-help.html")
	if err != nil {
		c.String(http.StatusNotFound, "智能检测帮助页面不存在")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

// 局域网检测深度调试页面
func debugLanDetectionHandler(c *gin.Context) {
	// 读取局域网检测深度调试页面文件
	content, err := os.ReadFile("templates/debug_lan_detection.html")
	if err != nil {
		c.String(http.StatusNotFound, "局域网检测深度调试页面不存在")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, string(content))
}

func deleteMessageHandler(c *gin.Context) {
	timestamp := c.PostForm("time")
	if timestamp == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "未提供时间戳"})
		return
	}

	messages, err := loadMessages()
	if err != nil {
		log.Printf("❌ 加载消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "加载消息失败"})
		return
	}

	originalCount := len(messages)
	var filteredMessages []Message
	for _, msg := range messages {
		if msg.Time != timestamp {
			filteredMessages = append(filteredMessages, msg)
		}
	}

	if len(filteredMessages) == originalCount {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "未找到要删除的消息"})
		return
	}

	err = saveMessages(filteredMessages)
	if err != nil {
		log.Printf("❌ 保存消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存消息失败"})
		return
	}

	// 广播删除消息
	broadcastData := map[string]interface{}{
		"time":   timestamp,
		"action": "delete",
	}
	broadcastMessage("message_deleted", broadcastData)
	log.Printf("✅ 删除消息已广播: %s", timestamp)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// 文件上传处理
func uploadFileHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "没有选择文件"})
		return
	}
	defer file.Close()

	if header.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "没有选择文件"})
		return
	}

	if !allowedFile(header.Filename) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "不支持的文件类型"})
		return
	}

	// 读取文件内容
	fileContent, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取文件失败"})
		return
	}

	fileSize := int64(len(fileContent))
	if fileSize > 16*1024*1024 { // 16MB限制
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "文件过大，最大支持16MB"})
		return
	}

	// 转换为Base64
	fileBase64 := base64.StdEncoding.EncodeToString(fileContent)

	// 生成文件ID
	fileID := fmt.Sprintf("%s_%s", time.Now().In(time.Local).Format("20060102_150405"), header.Filename)

	// 获取发送者IP
	senderIP := c.ClientIP()

	// 创建文件信息
	fileInfo := FileInfo{
		FileID:   fileID,
		Filename: header.Filename,
		Size:     fileSize,
		SizeMB:   float64(fileSize) / 1024 / 1024,
		Type:     header.Header.Get("Content-Type"),
		Data:     fileBase64,
		SenderIP: senderIP,
		SendTime: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		Action:   "file_incoming",
	}

	// 实时广播文件给所有设备
	broadcastMessage("file_incoming", fileInfo)
	log.Printf("✅ 文件实时共享已广播: %s (%d bytes) from %s", header.Filename, fileSize, senderIP)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  fmt.Sprintf("文件 \"%s\" 已发送给局域网所有设备！", header.Filename),
		"file_id":  fileID,
		"filename": header.Filename,
		"size":     fileSize,
	})
}

// 文件接收确认处理
func fileReceivedHandler(c *gin.Context) {
	var requestData struct {
		FileID string `json:"file_id"`
		Mode   string `json:"mode"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "请求数据格式错误"})
		return
	}

	if requestData.FileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "缺少文件ID"})
		return
	}

	receiverIP := c.ClientIP()
	if requestData.Mode == "" {
		requestData.Mode = "exclusive"
	}

	// 广播接收确认消息
	notificationData := map[string]interface{}{
		"file_id":      requestData.FileID,
		"receiver_ip":  receiverIP,
		"receive_time": time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		"mode":         requestData.Mode,
		"action":       "file_received",
	}
	broadcastMessage("file_received_notification", notificationData)
	log.Printf("✅ 文件接收确认: %s by %s (mode: %s)", requestData.FileID, receiverIP, requestData.Mode)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "接收确认已发送",
	})
}

// 获取模板数据
func getTemplatesHandler(c *gin.Context) {
	templatesData, err := loadTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "加载模板失败"})
		return
	}
	c.JSON(http.StatusOK, templatesData)
}

// 更新模板数据
func updateTemplatesHandler(c *gin.Context) {
	var templatesData TemplatesConfig
	if err := c.ShouldBindJSON(&templatesData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "数据格式错误"})
		return
	}

	if err := saveTemplates(templatesData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存模板数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "模板数据更新成功"})
}

// 向指定分类添加模板
func addTemplateToCategoryHandler(c *gin.Context) {
	categoryKey := c.Param("categoryKey")
	var templateData Template
	if err := c.ShouldBindJSON(&templateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "数据格式错误"})
		return
	}

	templatesConfig, err := loadTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "加载模板失败"})
		return
	}

	category, exists := templatesConfig.Categories[categoryKey]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "分类不存在"})
		return
	}

	newTitle := strings.TrimSpace(templateData.Title)
	newContent := strings.TrimSpace(templateData.Content)

	if newTitle == "" {
		newTitle = "未说明"
	}

	if newContent == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "模板内容不能为空"})
		return
	}

	// 检查重复：只比较内容
	for _, existingTemplate := range category.Templates {
		if strings.TrimSpace(existingTemplate.Content) == newContent {
			c.JSON(http.StatusOK, gin.H{
				"success":      true,
				"message":      "模板内容已存在，未重复添加",
				"is_duplicate": true,
			})
			return
		}
	}

	// 不重复，添加新模板
	category.Templates = append(category.Templates, Template{
		Title:   newTitle,
		Content: newContent,
	})
	templatesConfig.Categories[categoryKey] = category

	if err := saveTemplates(templatesConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "模板添加成功",
		"is_duplicate": false,
	})
}

// 导出模板数据
func exportTemplatesHandler(c *gin.Context) {
	formatType := c.Param("formatType")
	categoriesParam := c.Query("categories")

	templatesData, err := loadTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "加载模板失败"})
		return
	}

	var filteredData TemplatesConfig
	if categoriesParam != "" {
		// 导出指定栏目
		selectedCategories := strings.Split(categoriesParam, ",")
		filteredData.Categories = make(map[string]Category)
		for _, key := range selectedCategories {
			if category, exists := templatesData.Categories[key]; exists {
				filteredData.Categories[key] = category
			}
		}
	} else {
		// 导出全部
		filteredData = templatesData
	}

	timestamp := time.Now().In(time.Local).Format("20060102_150405")

	if formatType == "txt" {
		// TXT格式导出
		filename := fmt.Sprintf("templates_%s.txt", timestamp)
		txtContent := convertToTxtFormat(filteredData)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(txtContent))
	} else {
		// JSON格式导出
		filename := fmt.Sprintf("templates_%s.json", timestamp)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.JSON(http.StatusOK, filteredData)
	}
}

// 导入模板数据
func importTemplatesHandler(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "没有选择文件"})
		return
	}
	defer file.Close()

	if header.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "没有选择文件"})
		return
	}

	// 检查文件格式
	fileExtension := strings.ToLower(filepath.Ext(header.Filename)[1:])
	if fileExtension != "txt" && fileExtension != "json" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "只支持 TXT 和 JSON 格式的文件"})
		return
	}

	// 读取文件内容
	fileContent, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "读取文件失败"})
		return
	}

	// 获取参数
	importMode := c.PostForm("mode")
	importRange := c.PostForm("range")
	categoriesParam := c.PostForm("categories")

	if importMode == "" {
		importMode = "replace" // 默认为替换模式
	}

	// 加载当前模板数据
	currentTemplates, err := loadTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "加载当前模板失败"})
		return
	}

	// 解析导入的数据
	var importedData TemplatesConfig
	var duplicateCount int
	var addedCount int

	if fileExtension == "json" {
		// JSON格式解析
		err = json.Unmarshal(fileContent, &importedData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "JSON文件格式错误"})
			return
		}
	} else {
		// TXT格式解析
		importedData, err = parseTxtTemplates(string(fileContent))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "TXT文件格式错误: " + err.Error()})
			return
		}
	}

	// 处理导入范围
	allowedCategories := []string{"aftersale", "express", "presale", "purchase", "repair"}
	var targetCategories []string

	if importRange == "selected" && categoriesParam != "" {
		targetCategories = strings.Split(categoriesParam, ",")
	} else {
		targetCategories = allowedCategories
	}

	// 应用导入数据
	for _, categoryKey := range targetCategories {
		if importedCategory, exists := importedData.Categories[categoryKey]; exists {
			if currentTemplates.Categories == nil {
				currentTemplates.Categories = make(map[string]Category)
			}

			currentCategory := currentTemplates.Categories[categoryKey]

			if importMode == "replace" {
				// 替换模式：完全替换栏目内容
				currentCategory.Templates = importedCategory.Templates
				addedCount += len(importedCategory.Templates)
			} else {
				// 追加模式：添加新内容，跳过重复
				for _, newTemplate := range importedCategory.Templates {
					isExisting := false
					// 检查是否重复（只比较内容）
					for _, existingTemplate := range currentCategory.Templates {
						if strings.TrimSpace(existingTemplate.Content) == strings.TrimSpace(newTemplate.Content) {
							isExisting = true
							duplicateCount++
							break
						}
					}
					if !isExisting {
						currentCategory.Templates = append(currentCategory.Templates, newTemplate)
						addedCount++
					}
				}
			}

			currentTemplates.Categories[categoryKey] = currentCategory
		}
	}

	// 保存更新后的模板数据
	err = saveTemplates(currentTemplates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "保存模板数据失败"})
		return
	}

	// 构建成功消息
	var message string
	if importMode == "replace" {
		message = fmt.Sprintf("导入成功！共导入 %d 个模板", addedCount)
	} else {
		message = fmt.Sprintf("导入成功！新增 %d 个模板", addedCount)
		if duplicateCount > 0 {
			message += fmt.Sprintf("，跳过 %d 个重复项", duplicateCount)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         message,
		"added_count":     addedCount,
		"duplicate_count": duplicateCount,
	})
}

// 将模板数据转换为TXT格式
func convertToTxtFormat(templatesData TemplatesConfig) string {
	var lines []string
	allowedCategories := []string{"aftersale", "express", "presale", "purchase", "repair"}

	for _, categoryKey := range allowedCategories {
		if category, exists := templatesData.Categories[categoryKey]; exists {
			for _, template := range category.Templates {
				line := fmt.Sprintf("%s#%s，%s", category.Name, template.Title, template.Content)
				lines = append(lines, line)
			}
		}
	}

	return strings.Join(lines, "\n")
}

// 解析TXT格式的模板数据
func parseTxtTemplates(content string) (TemplatesConfig, error) {
	config := TemplatesConfig{
		Categories: make(map[string]Category),
	}

	// 初始化允许的分类
	allowedCategories := map[string]string{
		"售后问题": "aftersale",
		"快递问题": "express",
		"售前问题": "presale",
		"购买链接": "purchase",
		"维修问题": "repair",
	}

	// 初始化分类结构
	for chineseName, key := range allowedCategories {
		var icon string
		switch key {
		case "aftersale":
			icon = "🛠️"
		case "express":
			icon = "📦"
		case "presale":
			icon = "💬"
		case "purchase":
			icon = "🛒"
		case "repair":
			icon = "🔧"
		}
		config.Categories[key] = Category{
			Icon:      icon,
			Name:      chineseName,
			Templates: []Template{},
		}
	}

	// 解析内容，按行分割
	lines := strings.Split(content, "\n")

	// 用于累积多行内容
	currentTitle := ""
	currentContent := ""
	currentCategoryName := ""

	// 添加调试计数器
	totalTemplates := 0

	for _, line := range lines {
		// 不进行strings.TrimSpace处理，保留原始格式
		// line = strings.TrimSpace(line)

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			// 如果有累积的内容，保存它
			if currentCategoryName != "" && currentTitle != "" {
				// 查找对应的栏目键
				categoryKey := ""
				// 使用包含匹配而不是完全匹配，以处理"服务"和"问题"的差异
				if strings.Contains(currentCategoryName, "售后问题") || strings.Contains("售后问题", currentCategoryName) {
					categoryKey = "aftersale"
				} else if strings.Contains(currentCategoryName, "快递问题") || strings.Contains("快递问题", currentCategoryName) {
					categoryKey = "express"
				} else if strings.Contains(currentCategoryName, "售前问题") || strings.Contains("售前问题", currentCategoryName) {
					categoryKey = "presale"
				} else if strings.Contains(currentCategoryName, "购买链接") || strings.Contains("购买链接", currentCategoryName) {
					categoryKey = "purchase"
				} else if strings.Contains(currentCategoryName, "维修问题") || strings.Contains("维修问题", currentCategoryName) {
					categoryKey = "repair"
				}

				// 如果没有找到精确匹配，尝试模糊匹配
				if categoryKey == "" {
					// 检查是否包含关键字
					if strings.Contains(currentCategoryName, "售后") {
						categoryKey = "aftersale"
					} else if strings.Contains(currentCategoryName, "快递") {
						categoryKey = "express"
					} else if strings.Contains(currentCategoryName, "售前") {
						categoryKey = "presale"
					} else if strings.Contains(currentCategoryName, "购买") {
						categoryKey = "purchase"
					} else if strings.Contains(currentCategoryName, "维修") {
						categoryKey = "repair"
					}
				}

				if categoryKey != "" {
					// 创建新的模板
					newTemplate := Template{
						Title:   currentTitle,
						Content: currentContent,
					}

					// 添加到对应栏目
					category := config.Categories[categoryKey]
					category.Templates = append(category.Templates, newTemplate)
					config.Categories[categoryKey] = category

					// 增加计数器
					totalTemplates++
				}

				// 重置累积变量
				currentTitle = ""
				currentContent = ""
				currentCategoryName = ""
			}
			continue
		}

		// 检查是否是新的条目开始（包含#和中文逗号）
		if strings.Contains(line, "#") && strings.Contains(line, "，") {
			// 如果有之前累积的内容，先保存它
			if currentCategoryName != "" && currentTitle != "" {
				// 查找对应的栏目键
				categoryKey := ""
				// 使用包含匹配而不是完全匹配，以处理"服务"和"问题"的差异
				if strings.Contains(currentCategoryName, "售后问题") || strings.Contains("售后问题", currentCategoryName) {
					categoryKey = "aftersale"
				} else if strings.Contains(currentCategoryName, "快递问题") || strings.Contains("快递问题", currentCategoryName) {
					categoryKey = "express"
				} else if strings.Contains(currentCategoryName, "售前问题") || strings.Contains("售前问题", currentCategoryName) {
					categoryKey = "presale"
				} else if strings.Contains(currentCategoryName, "购买链接") || strings.Contains("购买链接", currentCategoryName) {
					categoryKey = "purchase"
				} else if strings.Contains(currentCategoryName, "维修问题") || strings.Contains("维修问题", currentCategoryName) {
					categoryKey = "repair"
				}

				// 如果没有找到精确匹配，尝试模糊匹配
				if categoryKey == "" {
					// 检查是否包含关键字
					if strings.Contains(currentCategoryName, "售后") {
						categoryKey = "aftersale"
					} else if strings.Contains(currentCategoryName, "快递") {
						categoryKey = "express"
					} else if strings.Contains(currentCategoryName, "售前") {
						categoryKey = "presale"
					} else if strings.Contains(currentCategoryName, "购买") {
						categoryKey = "purchase"
					} else if strings.Contains(currentCategoryName, "维修") {
						categoryKey = "repair"
					}
				}

				if categoryKey != "" {
					// 创建新的模板
					newTemplate := Template{
						Title:   currentTitle,
						Content: currentContent,
					}

					// 添加到对应栏目
					category := config.Categories[categoryKey]
					category.Templates = append(category.Templates, newTemplate)
					config.Categories[categoryKey] = category

					// 增加计数器
					totalTemplates++
				}
			}

			// 解析新的条目
			parts := strings.SplitN(line, "，", 2)
			if len(parts) == 2 {
				// 解析前面的部分：栏目名称#标题
				header := parts[0]
				content := parts[1]

				headerParts := strings.SplitN(header, "#", 2)
				if len(headerParts) == 2 {
					currentCategoryName = headerParts[0]
					currentTitle = headerParts[1]
					currentContent = content
				} else {
					// 格式不正确，跳过
					currentTitle = ""
					currentContent = ""
					currentCategoryName = ""
				}
			} else {
				// 格式不正确，跳过
				currentTitle = ""
				currentContent = ""
				currentCategoryName = ""
			}
		} else {
			// 这是内容的延续行
			if currentContent != "" {
				currentContent += "\n" + line
			} else {
				currentContent = line
			}
		}
	}

	// 保存最后一条记录
	if currentCategoryName != "" && currentTitle != "" {
		// 查找对应的栏目键
		categoryKey := ""
		// 使用包含匹配而不是完全匹配，以处理"服务"和"问题"的差异
		if strings.Contains(currentCategoryName, "售后问题") || strings.Contains("售后问题", currentCategoryName) {
			categoryKey = "aftersale"
		} else if strings.Contains(currentCategoryName, "快递问题") || strings.Contains("快递问题", currentCategoryName) {
			categoryKey = "express"
		} else if strings.Contains(currentCategoryName, "售前问题") || strings.Contains("售前问题", currentCategoryName) {
			categoryKey = "presale"
		} else if strings.Contains(currentCategoryName, "购买链接") || strings.Contains("购买链接", currentCategoryName) {
			categoryKey = "purchase"
		} else if strings.Contains(currentCategoryName, "维修问题") || strings.Contains("维修问题", currentCategoryName) {
			categoryKey = "repair"
		}

		// 如果没有找到精确匹配，尝试模糊匹配
		if categoryKey == "" {
			// 检查是否包含关键字
			if strings.Contains(currentCategoryName, "售后") {
				categoryKey = "aftersale"
			} else if strings.Contains(currentCategoryName, "快递") {
				categoryKey = "express"
			} else if strings.Contains(currentCategoryName, "售前") {
				categoryKey = "presale"
			} else if strings.Contains(currentCategoryName, "购买") {
				categoryKey = "purchase"
			} else if strings.Contains(currentCategoryName, "维修") {
				categoryKey = "repair"
			}
		}

		if categoryKey != "" {
			// 创建新的模板
			newTemplate := Template{
				Title:   currentTitle,
				Content: currentContent,
			}

			// 添加到对应栏目
			category := config.Categories[categoryKey]
			category.Templates = append(category.Templates, newTemplate)
			config.Categories[categoryKey] = category

			// 增加计数器
			totalTemplates++
		}
	}

	// 输出调试信息
	log.Printf("✅ 解析TXT模板完成，总共解析到 %d 个模板", totalTemplates)

	return config, nil
}

// 获取map的所有键
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// 🔧 修复后的客户端IP获取函数
func getRealClientIP(c *gin.Context) string {
	// 获取服务器本机IP，用于排除
	localIP := getLocalIP()

	// 获取各种可能的IP来源
	xForwardedFor := c.Request.Header.Get("X-Forwarded-For")
	xRealIP := c.Request.Header.Get("X-Real-IP")
	cfConnectingIP := c.Request.Header.Get("CF-Connecting-IP")
	trueClientIP := c.Request.Header.Get("True-Client-IP")
	remoteAddr := c.Request.RemoteAddr
	ginClientIP := c.ClientIP()

	log.Printf("🔍 IP获取调试信息:")
	log.Printf("  - 服务器本机IP: %s", localIP)
	log.Printf("  - X-Forwarded-For: %s", xForwardedFor)
	log.Printf("  - X-Real-IP: %s", xRealIP)
	log.Printf("  - CF-Connecting-IP: %s", cfConnectingIP)
	log.Printf("  - True-Client-IP: %s", trueClientIP)
	log.Printf("  - RemoteAddr: %s", remoteAddr)
	log.Printf("  - Gin ClientIP: %s", ginClientIP)

	// 🔧 新的智能IP获取策略：优先处理X-Forwarded-For
	var candidateIPs []string

	// 1. 优先处理X-Forwarded-For（最常见的代理头）
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		for i, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" && net.ParseIP(ip) != nil {
				candidateIPs = append(candidateIPs, ip)
				log.Printf("  - X-Forwarded-For[%d]: %s", i, ip)
			}
		}
	}

	// 2. 其他代理头作为备选
	if cfConnectingIP != "" && net.ParseIP(cfConnectingIP) != nil {
		candidateIPs = append(candidateIPs, cfConnectingIP)
		log.Printf("  - CF-Connecting-IP: %s", cfConnectingIP)
	}

	if trueClientIP != "" && net.ParseIP(trueClientIP) != nil {
		candidateIPs = append(candidateIPs, trueClientIP)
		log.Printf("  - True-Client-IP: %s", trueClientIP)
	}

	if xRealIP != "" && net.ParseIP(xRealIP) != nil {
		candidateIPs = append(candidateIPs, xRealIP)
		log.Printf("  - X-Real-IP: %s", xRealIP)
	}

	// 3. 直连IP作为最后备选
	if remoteAddr != "" {
		if colonIndex := strings.LastIndex(remoteAddr, ":"); colonIndex != -1 {
			ip := remoteAddr[:colonIndex]
			if net.ParseIP(ip) != nil {
				candidateIPs = append(candidateIPs, ip)
				log.Printf("  - RemoteAddr: %s", ip)
			}
		}
	}

	if ginClientIP != "" && net.ParseIP(ginClientIP) != nil {
		candidateIPs = append(candidateIPs, ginClientIP)
		log.Printf("  - Gin ClientIP: %s", ginClientIP)
	}

	log.Printf("  - 候选IP列表: %v", candidateIPs)

	// 4. 智能选择最佳IP
	var bestIP string

	for _, ip := range candidateIPs {
		// 跳过无效IP
		if ip == "127.0.0.1" || ip == "::1" {
			log.Printf("  - 跳过IP: %s (回环地址)", ip)
			continue
		}

		// 跳过服务器本机IP
		if ip == localIP {
			log.Printf("  - 跳过IP: %s (服务器本机IP)", ip)
			continue
		}

		// 跳过明显的CDN内部IP
		if strings.HasPrefix(ip, "15.") || strings.HasPrefix(ip, "10.") {
			log.Printf("  - 跳过IP: %s (CDN内部IP)", ip)
			continue
		}

		// 🔧 关键修复：优先选择局域网IP
		if isPrivateIPAddress(ip) {
			bestIP = ip
			log.Printf("  - ✅ 选择局域网IP: %s", bestIP)
			break
		}

		// 如果没有局域网IP，选择第一个有效的公网IP
		if bestIP == "" {
			bestIP = ip
			log.Printf("  - ✅ 选择公网IP: %s", bestIP)
		}
	}

	// 最后的备选方案
	if bestIP == "" {
		if ginClientIP != "" {
			bestIP = ginClientIP
			log.Printf("  - 使用备选ClientIP: %s", bestIP)
		} else {
			bestIP = "unknown"
			log.Printf("  - 无法确定客户端IP")
		}
	}

	log.Printf("  - 🎯 最终客户端IP: %s", bestIP)
	return bestIP
}

// 检查IP是否为私有地址
func isPrivateIPAddress(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// IPv4私有地址范围
	if parsedIP.To4() != nil {
		// 192.168.0.0/16
		if parsedIP.IsPrivate() {
			return true
		}
		// 手动检查私有范围（以防IsPrivate方法不可用）
		ipParts := strings.Split(ip, ".")
		if len(ipParts) == 4 {
			first, _ := strconv.Atoi(ipParts[0])
			second, _ := strconv.Atoi(ipParts[1])

			// 192.168.x.x
			if first == 192 && second == 168 {
				return true
			}
			// 10.x.x.x
			if first == 10 {
				return true
			}
			// 172.16-31.x.x
			if first == 172 && second >= 16 && second <= 31 {
				return true
			}
		}
	}

	return false
}

// 局域网检测API处理函数
func lanCheckHandler(c *gin.Context) {
	host := c.Request.Host
	userAgent := c.Request.Header.Get("User-Agent")
	referrer := c.Request.Header.Get("Referer")

	// 使用增强的IP获取函数
	clientIP := getRealClientIP(c)

	// 检查是否强制显示提示框的参数
	forcePrompt := c.Query("force_prompt") == "true"

	// 获取各种可能的代理头信息（用于调试）
	xForwardedFor := c.Request.Header.Get("X-Forwarded-For")
	xRealIP := c.Request.Header.Get("X-Real-IP")
	xForwardedHost := c.Request.Header.Get("X-Forwarded-Host")
	xOriginalHost := c.Request.Header.Get("X-Original-Host")

	log.Printf("🔍 局域网检测请求开始:")
	log.Printf("  - Host: %s", host)
	log.Printf("  - User-Agent: %s", userAgent)
	log.Printf("  - Referer: %s", referrer)
	log.Printf("  - ClientIP: %s", clientIP)
	log.Printf("  - X-Forwarded-For: %s", xForwardedFor)
	log.Printf("  - X-Real-IP: %s", xRealIP)
	log.Printf("  - X-Forwarded-Host: %s", xForwardedHost)
	log.Printf("  - X-Original-Host: %s", xOriginalHost)
	log.Printf("  - CF-Connecting-IP: %s", c.Request.Header.Get("CF-Connecting-IP"))
	log.Printf("  - True-Client-IP: %s", c.Request.Header.Get("True-Client-IP"))
	log.Printf("  - ForcePrompt: %v", forcePrompt)

	// 获取本机局域网IP
	localIP := getLocalIP()
	log.Printf("  - 本机IP: %s", localIP)

	// 判断是否为IP地址访问（增强检测）
	hostname := host
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		hostname = host[:colonIndex]
	}
	log.Printf("  - 主机名: %s", hostname)

	// 检查是否为IP地址访问（增强检测）
	isIPAccess := false
	if net.ParseIP(hostname) != nil {
		isIPAccess = true
		log.Printf("  - 检测结果: IP地址访问")
	} else {
		log.Printf("  - 检测结果: 域名访问")
	}
	log.Printf("  - 是否IP访问: %v", isIPAccess)

	// 🔧 新的智能检测策略：对于域名访问，提供智能切换选项
	isClientInLAN := false

	// 方法1：检查客户端IP是否为真实的局域网IP
	if isPrivateIPAddress(clientIP) && !strings.HasPrefix(clientIP, "15.") {
		isClientInLAN = true
		log.Printf("  - 方法1: 检测到真实局域网IP: %s", clientIP)
	}

	// 方法2：对于域名访问，采用智能提示策略
	if !isIPAccess && !isClientInLAN {
		// 🔧 关键改进：对于域名访问，直接提供切换选项
		// 让用户自己判断是否在局域网内，这样避免了CDN IP检测的技术限制
		isClientInLAN = true
		log.Printf("  - 方法2: 域名访问智能提示策略 - 提供切换选项供用户选择")
		log.Printf("    * 检测到的IP: %s (可能是CDN代理IP)", clientIP)
		log.Printf("    * 策略: 显示智能切换提示，让用户自行判断")
	}

	// 特殊处理：IPv6回环地址
	if strings.Contains(clientIP, ":") {
		trimmedIP := strings.Trim(clientIP, "[]")
		if trimmedIP == "::1" {
			isClientInLAN = true
			log.Printf("  - IPv6回环地址特殊处理: 视为局域网")
		} else if strings.HasPrefix(trimmedIP, "fe80:") {
			isClientInLAN = true
			log.Printf("  - 匹配IPv6链路本地地址")
		} else if strings.HasPrefix(trimmedIP, "fc") || strings.HasPrefix(trimmedIP, "fd") {
			isClientInLAN = true
			log.Printf("  - 匹配IPv6唯一本地地址")
		}
	}

	// 特殊处理：如果强制显示提示框
	if forcePrompt {
		isClientInLAN = true
		isIPAccess = false
		log.Printf("  - 特殊处理: 强制显示提示框")
	}

	log.Printf("  - 客户端在局域网: %v", isClientInLAN)

	// 生成局域网访问地址
	lanURL := fmt.Sprintf("http://%s:%d", localIP, Port)
	log.Printf("  - 局域网地址: %s", lanURL)

	// 判断是否需要提示切换（改进的逻辑）
	needSwitchPrompt := false
	log.Printf("  - 开始判断是否需要提示切换...")
	log.Printf("  - 条件1 (域名访问): %v", !isIPAccess)
	log.Printf("  - 条件2 (局域网客户端): %v", isClientInLAN)

	// 改进的判断逻辑：适用于所有域名
	// 只要不是IP地址访问且客户端在局域网环境，就提示切换
	// 不再检查特定域名，而是适用于所有域名
	if !isIPAccess && !isClientInLAN && !forcePrompt {
		// 检查是否在同一局域网网段
		localIPParts := strings.Split(localIP, ".")
		clientIPParts := strings.Split(clientIP, ".")
		if len(localIPParts) == 4 && len(clientIPParts) == 4 {
			// 比较前三个部分是否相同（C类网络）
			if localIPParts[0] == clientIPParts[0] &&
				localIPParts[1] == clientIPParts[1] &&
				localIPParts[2] == clientIPParts[2] {
				isClientInLAN = true
				log.Printf("  - 特殊处理: 客户端IP与本机IP在同一网段，视为局域网")
			}
		}
	}

	if (!isIPAccess && isClientInLAN) || forcePrompt {
		log.Printf("  - 满足条件，建议切换到局域网地址...")

		// 测试局域网地址是否可访问（增加详细日志）
		log.Printf("  - 开始测试局域网地址可达性...")
		lanURL := fmt.Sprintf("http://%s:%d", localIP, Port)
		log.Printf("  - 测试地址: %s", lanURL)

		// 测试局域网地址是否可访问
		log.Printf("  - 开始HTTP客户端测试...")
		client := &http.Client{
			Timeout: 3 * time.Second,
		}
		resp, err := client.Get(lanURL)
		if err != nil {
			log.Printf("❌ 局域网地址测试失败: %v", err)
			// 即使测试失败，如果本机IP有效且不是回环地址，仍然提示切换
			if localIP != "127.0.0.1" && localIP != "::1" && net.ParseIP(localIP) != nil {
				log.Printf("⚠️ 测试失败但IP有效，仍然提示切换")
				needSwitchPrompt = true
			} else {
				log.Printf("❌ IP无效或为回环地址，不提示切换")
			}
		} else {
			defer resp.Body.Close()
			log.Printf("✅ 局域网地址测试成功，状态码: %d", resp.StatusCode)
			// 局域网地址可访问，显示提示
			needSwitchPrompt = true
		}
	} else {
		log.Printf("  - 不满足基本条件:")
		log.Printf("    * 域名访问: %v", !isIPAccess)
		log.Printf("    * 局域网客户端: %v", isClientInLAN)
	}

	// 特殊处理：如果强制显示提示框
	if forcePrompt {
		needSwitchPrompt = true
		log.Printf("  - 特殊处理: 强制显示提示框")
	}

	log.Printf("🔍 局域网检测完成: IsIPAccess=%v, IsClientInLAN=%v, NeedPrompt=%v",
		isIPAccess, isClientInLAN, needSwitchPrompt)

	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"current_host":       host,
		"client_ip":          clientIP,
		"local_ip":           localIP,
		"is_ip_access":       isIPAccess,
		"is_client_in_lan":   isClientInLAN,
		"need_switch_prompt": needSwitchPrompt,
		"lan_url":            lanURL,
		"user_agent":         userAgent,
		"referrer":           referrer,
		"x_forwarded_for":    xForwardedFor,
		"x_real_ip":          xRealIP,
		"x_forwarded_host":   xForwardedHost,
		"x_original_host":    xOriginalHost,
		"force_prompt":       forcePrompt,
	})
}

func main() {
	// 设置中国时区 (UTC+8) - 强制设置
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果无法加载Asia/Shanghai，尝试使用固定偏移
		log.Printf("⚠️ 无法加载Asia/Shanghai时区: %v", err)
		loc = time.FixedZone("CST", 8*3600) // UTC+8
		log.Printf("✅ 使用固定时区偏移 UTC+8")
	} else {
		log.Printf("✅ 时区已设置为中国时区 (Asia/Shanghai)")
	}
	time.Local = loc

	// 验证时区设置
	now := time.Now()
	log.Printf("🕐 当前时间: %s (时区: %s)", now.Format("2006-01-02 15:04:05"), now.Location())

	// 检查是否已有实例运行
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", Port))
	if err != nil {
		log.Printf("⚠️ 程序已经在运行中，端口 %d 被占用", Port)
		return
	}
	ln.Close()

	// 确保模板文件存在
	if err := ensureTemplatesFile(); err != nil {
		log.Printf("❌ 创建模板文件失败: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 配置CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	r.Use(cors.New(config))

	// 加载HTML模板
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseGlob("templates/*"))
	r.SetHTMLTemplate(tmpl)

	// 静态文件服务
	r.Static("/static", "./static")

	// WebSocket路由
	r.GET("/ws", handleWebSocket)

	// HTTP路由
	r.GET("/", indexHandler)
	r.GET("/qr-code", qrCodeHandler)
	r.GET("/test-qr", testQRHandler)                          // 新增：二维码测试页面
	r.GET("/test-lan", testLANHandler)                        // 新增：局域网检测测试页面
	r.GET("/test-domain", testDomainHandler)                  // 新增：域名检测测试页面
	r.GET("/diagnostic", diagnosticHandler)                   // 新增：诊断工具页面
	r.GET("/debug-detection", debugDetectionHandler)          // 新增：调试检测页面
	r.GET("/advanced-debug", advancedDebugHandler)            // 新增：高级调试页面
	r.GET("/host-analysis", hostAnalysisHandler)              // 新增：Host头行为分析页面
	r.GET("/debug-lan-detection", debugLanDetectionHandler)   // 新增：局域网检测深度调试页面
	r.GET("/smart-detection-help", smartDetectionHelpHandler) // 新增：智能检测帮助页面
	r.POST("/add", addMessageHandler)
	r.POST("/delete", deleteMessageHandler)
	r.POST("/upload", uploadFileHandler)
	r.POST("/file_received", fileReceivedHandler)

	// API路由
	r.GET("/api/templates", getTemplatesHandler)
	r.POST("/api/templates", updateTemplatesHandler)
	r.POST("/api/templates/category/:categoryKey", addTemplateToCategoryHandler)
	r.GET("/api/templates/export/:formatType", exportTemplatesHandler)
	r.POST("/api/templates/import", importTemplatesHandler)
	r.GET("/api/lan-check", lanCheckHandler) // 新增局域网检测API

	// 获取本机IP
	localIP := getLocalIP()

	log.Printf("\n🚀 启动祖宇字文共享服务器...")
	log.Printf("📱 本地访问: http://127.0.0.1:%d", Port)
	log.Printf("🌐 局域网访问: http://%s:%d", localIP, Port)
	log.Printf("⚡ 实时同步功能已启用")
	log.Printf("\n按 Ctrl+C 停止服务器\n")

	// 启动服务器
	log.Fatal(r.Run(fmt.Sprintf(":%d", Port)))
}
