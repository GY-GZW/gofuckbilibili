// Copyright (C) 2026 GY-GZW
// This program is licensed under AGPLv3 or later.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// 定义 API 响应结构体
type ViewResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Cid   int    `json:"cid"`
		Bvid  string `json:"bvid"`
		Title string `json:"title"`
	} `json:"data"`
}

type PlayUrlResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Dash struct {
			Duration int      `json:"duration"`
			Video    []Stream `json:"video"`
			Audio    []Stream `json:"audio"`
		} `json:"dash"`
	} `json:"data"`
}

type Stream struct {
	ID         int      `json:"id"`
	BaseURL    string   `json:"baseUrl"`
	BaseURL2   string   `json:"base_url"` // 兼容下划线命名
	BackupURL  []string `json:"backupUrl"`
	BackupURL2 []string `json:"backup_url"`
	MimeType   string   `json:"mimeType"`
	Codecs     string   `json:"codecs"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Bandwidth  int      `json:"bandwidth"`
}

func (s *Stream) GetRealURL() string {
	if s.BaseURL != "" {
		return s.BaseURL
	}
	if s.BaseURL2 != "" {
		return s.BaseURL2
	}
	if len(s.BackupURL) > 0 {
		return s.BackupURL[0]
	}
	if len(s.BackupURL2) > 0 {
		return s.BackupURL2[0]
	}
	return ""
}

func fetchJSON(url string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// 必须设置 Referer 和 User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	// 不再设置 Cookie

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP 状态码错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}

func downloadFile(url string, filename string) error {
	fmt.Printf("正在下载: %s ...\n", filename)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0")
	req.Header.Set("Referer", "https://www.bilibili.com/")
	// 不再设置 Cookie

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("下载完成: %s\n", filename)
	return nil
}

func main() {
	url := flag.String("url", "", "B站视频地址")
	flag.Parse()

	if *url == "" {
		log.Fatal("请提供 -url 参数")
	}

	// 1. 从 URL 中提取 BV 号
	var bvid string
	if strings.Contains(*url, "BV") {
		parts := strings.Split(*url, "BV")
		if len(parts) > 1 {
			// 简单提取 BV 后的一串字符，直到遇到非 BV 字符或结束
			raw := parts[1]
			var sb strings.Builder
			for _, r := range raw {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
					sb.WriteRune(r)
				} else {
					break
				}
			}
			bvid = "BV" + sb.String()
		}
	}

	if bvid == "" {
		log.Fatal("无法从 URL 中提取 BV 号")
	}

	fmt.Printf("检测到 BV 号: %s\n", bvid)

	// 2. 获取 CID
	viewUrl := fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", bvid)
	var viewResp ViewResponse
	if err := fetchJSON(viewUrl, &viewResp); err != nil {
		log.Fatalf("获取视频信息失败: %v", err)
	}
	if viewResp.Code != 0 {
		log.Fatalf("API 错误: %s", viewResp.Message)
	}
	cid := viewResp.Data.Cid
	fmt.Printf("获取到 CID: %d, 标题: %s\n", cid, viewResp.Data.Title)

	// 3. 获取播放地址 (Dash 格式)
	// qn=80 表示请求 1080P (但无 Cookie 可能会降级), fnval=16 或 4048 表示 Dash 格式
	playUrl := fmt.Sprintf("https://api.bilibili.com/x/player/playurl?bvid=%s&cid=%d&qn=80&fnval=4048", bvid, cid)
	var playResp PlayUrlResponse
	if err := fetchJSON(playUrl, &playResp); err != nil {
		log.Fatalf("获取播放地址失败: %v", err)
	}
	if playResp.Code != 0 {
		log.Fatalf("播放地址 API 错误: %s", playResp.Message)
	}

	if len(playResp.Data.Dash.Video) == 0 || len(playResp.Data.Dash.Audio) == 0 {
		log.Fatal("未找到可用的视频或音频流，可能视频受限或需要登录")
	}

	// 4. 选择最佳画质和音质
	bestVideo := playResp.Data.Dash.Video[0]
	for _, v := range playResp.Data.Dash.Video {
		if v.ID > bestVideo.ID {
			bestVideo = v
		}
	}

	bestAudio := playResp.Data.Dash.Audio[0]
	for _, a := range playResp.Data.Dash.Audio {
		if a.ID > bestAudio.ID {
			bestAudio = a
		}
	}

	videoURL := bestVideo.GetRealURL()
	audioURL := bestAudio.GetRealURL()

	if videoURL == "" || audioURL == "" {
		log.Fatal("无法获取有效的下载链接")
	}

	fmt.Printf("选择视频: ID=%d, Resolution=%dx%d\n", bestVideo.ID, bestVideo.Width, bestVideo.Height)
	fmt.Printf("选择音频: ID=%d\n", bestAudio.ID)

	// 5. 下载
	if err := downloadFile(videoURL, "video.m4s"); err != nil {
		log.Fatalf("下载视频失败: %v", err)
	}
	if err := downloadFile(audioURL, "audio.m4s"); err != nil {
		log.Fatalf("下载音频失败: %v", err)
	}

	// 6. 合并
	fmt.Println("正在合并音视频...")
	cmd := exec.Command("ffmpeg", "-y", "-i", "video.m4s", "-i", "audio.m4s", "-c", "copy", "output.mp4")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("FFmpeg 合并失败: %v", err)
	}

	fmt.Println("成功！文件已保存为 output.mp4")
}
