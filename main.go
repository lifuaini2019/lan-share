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
	"txt": true, "pdf": true, "png": true, "jpg": true, "jpeg": true,
	"gif": true, "doc": true, "docx": true, "xls": true, "xlsx": true,
	"ppt": true, "pptx": true, "zip": true, "rar": true,
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

	return messages, nil
}

func saveMessages(messages []Message) error {
	var lines []string
	for _, msg := range messages {
		lines = append(lines, fmt.Sprintf("%s|%s", msg.Time, msg.Content))
	}

	return os.WriteFile(DataFile, []byte(strings.Join(lines, "\n")+"\n"), 0644)
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
						Content: "欢迎使用祖宇字文共享系统！您可以在这里快速分享文字内容。",
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
				syncData := map[string]interface{}{
					"messages": messages,
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

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	newMessage := Message{
		Time:    timestamp,
		Content: content,
	}

	messages = append(messages, newMessage)
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
	content, err := os.ReadFile("test_lan_detection.html")
	if err != nil {
		c.String(http.StatusNotFound, "测试页面不存在")
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
	fileID := fmt.Sprintf("%s_%s", time.Now().Format("20060102_150405"), header.Filename)

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
		SendTime: time.Now().Format("2006-01-02 15:04:05"),
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
		"receive_time": time.Now().Format("2006-01-02 15:04:05"),
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

	timestamp := time.Now().Format("20060102_150405")

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
				line := fmt.Sprintf("%s,%s,%s", category.Name, template.Title, template.Content)
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
		default:
			icon = "📝"
		}

		config.Categories[key] = Category{
			Icon:      icon,
			Name:      chineseName,
			Templates: []Template{},
		}
	}

	// 解析文本内容
	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 按逗号分割：分类名,模板标题,模板内容
		parts := strings.SplitN(line, ",", 3)
		if len(parts) != 3 {
			return config, fmt.Errorf("第%d行格式错误，应为: 分类名,模板标题,模板内容", lineNum+1)
		}

		categoryName := strings.TrimSpace(parts[0])
		templateTitle := strings.TrimSpace(parts[1])
		templateContent := strings.TrimSpace(parts[2])

		// 查找对应的分类键
		categoryKey, exists := allowedCategories[categoryName]
		if !exists {
			return config, fmt.Errorf("第%d行中的分类'%s'不受支持，支持的分类有：%v", lineNum+1, categoryName, getMapKeys(allowedCategories))
		}

		// 验证必要字段
		if templateTitle == "" {
			templateTitle = "未说明"
		}
		if templateContent == "" {
			return config, fmt.Errorf("第%d行中的模板内容不能为空", lineNum+1)
		}

		// 添加模板到对应分类
		category := config.Categories[categoryKey]
		category.Templates = append(category.Templates, Template{
			Title:   templateTitle,
			Content: templateContent,
		})
		config.Categories[categoryKey] = category
	}

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

// 局域网检测API处理函数
func lanCheckHandler(c *gin.Context) {
	// 获取请求头信息
	host := c.Request.Host
	userAgent := c.Request.Header.Get("User-Agent")
	referrer := c.Request.Header.Get("Referer")
	clientIP := c.ClientIP()

	// 获取本机局域网IP
	localIP := getLocalIP()

	// 判断是否为域名访问
	hostname := host
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		hostname = host[:colonIndex]
	}

	// 检查是否为IP地址访问
	isIPAccess := net.ParseIP(hostname) != nil

	// 检查客户端IP是否在局域网范围内
	isClientInLAN := false
	if strings.HasPrefix(clientIP, "192.168.") ||
		strings.HasPrefix(clientIP, "10.") ||
		strings.HasPrefix(clientIP, "172.") ||
		clientIP == "127.0.0.1" || clientIP == "::1" {
		isClientInLAN = true
	}

	// 生成局域网访问地址
	lanURL := fmt.Sprintf("http://%s:%d", localIP, Port)

	// 判断是否需要提示切换
	needSwitchPrompt := false
	if !isIPAccess && isClientInLAN {
		// 域名访问且客户端在局域网内，需要提示
		needSwitchPrompt = true
	}

	log.Printf("🔍 局域网检测: Host=%s, ClientIP=%s, IsIPAccess=%v, IsClientInLAN=%v, NeedPrompt=%v",
		host, clientIP, isIPAccess, isClientInLAN, needSwitchPrompt)

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
	})
}

func main() {
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
	r.GET("/test-qr", testQRHandler)
	r.GET("/test-lan", testLANHandler) // 新增：局域网检测测试页面
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

	log.Printf("\n🚀 启动祖宇局域网共享服务器...")
	log.Printf("📱 本地访问: http://127.0.0.1:%d", Port)
	log.Printf("🌐 局域网访问: http://%s:%d", localIP, Port)
	log.Printf("⚡ 实时同步功能已启用")
	log.Printf("\n按 Ctrl+C 停止服务器\n")

	// 启动服务器
	log.Fatal(r.Run(fmt.Sprintf(":%d", Port)))
}
