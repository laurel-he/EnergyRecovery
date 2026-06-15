package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type ChatMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type deepSeekReq struct {
    Model    string        `json:"model"`
    Messages []ChatMessage `json:"messages"`
    MaxTokens int          `json:"max_tokens"`
}

type deepSeekResp struct {
    Choices []struct {
        Message ChatMessage `json:"message"`
    } `json:"choices"`
}

func callDeepSeek(prompt string) (string, error) {
    reqBody := deepSeekReq{
        Model: DEEPSEEK_MODEL,
        Messages: []ChatMessage{
            {Role: "system", Content: "你是一位温和、专业的健康恢复教练，用中文回答。"},
            {Role: "user", Content: prompt},
        },
        MaxTokens: 500,
    }
    jsonData, _ := json.Marshal(reqBody)

    req, _ := http.NewRequest("POST", DEEPSEEK_API_URL, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+DEEPSEEK_API_KEY)

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != 200 {
        return "", fmt.Errorf("API错误(%d): %s", resp.StatusCode, string(body))
    }

    var result deepSeekResp
    json.Unmarshal(body, &result)
    if len(result.Choices) > 0 {
        return result.Choices[0].Message.Content, nil
    }
    return "", fmt.Errorf("空响应")
}