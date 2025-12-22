// Copyright (C) 2025 GY-GZW
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
	"regexp"
)

type PlayInfo struct {
	Data struct {
		Dash struct {
			Video []struct {
				BaseUrl string `json:"baseUrl"`
			} `json:"video"`
			Audio []struct {
				BaseUrl string `json:"baseUrl"`
			} `json:"audio"`
		} `json:"dash"`
	} `json:"data"`
}

func fetchBilibiliResponse(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 Bilibili 必需的防盗链
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36 Edg/118.0.2088.76")
	req.Header.Set("Referer", "https://www.bilibili.com/")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close() // 立即关闭，避免泄漏
		return nil, fmt.Errorf("HTTP %d: 访问被拒绝（可能防盗链失败）", resp.StatusCode)
	}
	return resp, nil
}
func main() {
	url := flag.String("url", "", "B站视频地址")
	flag.Parse()
	resp, err := fetchBilibiliResponse(*url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("读取失败:", err)
		return
	}
	// body 是 []byte
	html := string(body)
	re := regexp.MustCompile(`(?s)window\.__playinfo__=(.*?)</script>`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		log.Fatal("未找到 __playinfo__ 数据")
	}
	jsonStr := matches[1] // 提取出json
	var info PlayInfo
	json.Unmarshal([]byte(jsonStr), &info)
	videoURL := info.Data.Dash.Video[0].BaseUrl
	audioURL := info.Data.Dash.Audio[0].BaseUrl
	resp, err = fetchBilibiliResponse(videoURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	file, err := os.Create("video.mp4")
	if err != nil {
		log.Fatal("无法创建 video.mp4:", err)
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal("写入视频文件失败:", err)
	}
	resp, err = fetchBilibiliResponse(audioURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	file, err = os.Create("audio.mp3")
	if err != nil {
		log.Fatal("无法创建 audio.mp3:", err)
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal("写入视频文件失败:", err)
	}
	cmd := exec.Command(
		"ffmpeg",
		"-i", "video.mp4",
		"-i", "audio.mp3",
		"-vcodec", "copy",
		"-acodec", "copy",
		"output.mp4",
	)
	err = cmd.Run() // 阻塞执行，等待完成
	if err != nil {
		log.Fatal("ffmpeg 合并失败，可能是没安装ffmpeg:", err)
	}
}
