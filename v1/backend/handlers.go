package main

import (
    "fmt"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
)

// ---------- 用户 ----------
func getUser(c *gin.Context) {
    deviceID := c.Query("device_id")
    if deviceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "device_id required"})
        return
    }
    var u User
    err := db.QueryRow(`SELECT device_id, COALESCE(height,0), COALESCE(weight,0), COALESCE(body_fat,0),
        COALESCE(age,0), COALESCE(symptoms,''), COALESCE(diagnosis_type,''), COALESCE(diagnosis_label,''),
        COALESCE(micro_death_index,0), COALESCE(micro_death_level,''), COALESCE(vitality_score,44),
        COALESCE(current_stage,1), COALESCE(stage_start_date,'')
        FROM users WHERE device_id=?`, deviceID).
        Scan(&u.DeviceID, &u.Height, &u.Weight, &u.BodyFat, &u.Age, &u.Symptoms,
            &u.DiagnosisType, &u.DiagnosisLabel, &u.MicroDeathIndex, &u.MicroDeathLevel,
            &u.VitalityScore, &u.CurrentStage, &u.StageStartDate)
    if err != nil {
        // 创建默认用户
        _, err := db.Exec(`INSERT INTO users (device_id) VALUES (?)`, deviceID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        u = User{DeviceID: deviceID, VitalityScore: 44, CurrentStage: 1}
    }
    c.JSON(http.StatusOK, u)
}

func updateUser(c *gin.Context) {
    var u User
    if err := c.ShouldBindJSON(&u); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if u.DeviceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "device_id required"})
        return
    }
    _, err := db.Exec(`INSERT INTO users (device_id, height, weight, body_fat, age, symptoms,
        diagnosis_type, diagnosis_label, micro_death_index, micro_death_level,
        vitality_score, current_stage, stage_start_date)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)
        ON DUPLICATE KEY UPDATE
        height=VALUES(height), weight=VALUES(weight), body_fat=VALUES(body_fat),
        age=VALUES(age), symptoms=VALUES(symptoms),
        diagnosis_type=VALUES(diagnosis_type), diagnosis_label=VALUES(diagnosis_label),
        micro_death_index=VALUES(micro_death_index), micro_death_level=VALUES(micro_death_level),
        vitality_score=VALUES(vitality_score), current_stage=VALUES(current_stage),
        stage_start_date=VALUES(stage_start_date)`,
        u.DeviceID, u.Height, u.Weight, u.BodyFat, u.Age, u.Symptoms,
        u.DiagnosisType, u.DiagnosisLabel, u.MicroDeathIndex, u.MicroDeathLevel,
        u.VitalityScore, u.CurrentStage, u.StageStartDate)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// ---------- 任务 ----------
func getTodayTasks(c *gin.Context) {
    deviceID := c.Query("device_id")
    date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
    if deviceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "device_id required"})
        return
    }
    rows, err := db.Query(`SELECT task_id, completed FROM task_completions WHERE device_id=? AND task_date=?`, deviceID, date)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()
    taskMap := make(map[string]bool)
    for rows.Next() {
        var id string
        var comp bool
        rows.Scan(&id, &comp)
        taskMap[id] = comp
    }
    // 基础任务（后面可根据阶段动态调整）
    tasks := []map[string]interface{}{
        {"id": "water", "name": "喝够8杯水", "icon": "💧", "hint": "约1.6L，少量多次", "completed": taskMap["water"]},
        {"id": "protein", "name": "早餐吃一份蛋白质", "icon": "🥚", "hint": "鸡蛋/牛奶/豆制品", "completed": taskMap["protein"]},
        {"id": "sleep", "name": "23:30前躺下睡觉", "icon": "🌙", "hint": "放下手机，闭眼即可", "completed": taskMap["sleep"]},
        {"id": "walk", "name": "散步10分钟", "icon": "🚶", "hint": "不要求速度，出门就行", "completed": taskMap["walk"]},
    }
    c.JSON(http.StatusOK, tasks)
}

func toggleTask(c *gin.Context) {
    var req TaskCompletion
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if req.DeviceID == "" || req.TaskDate == "" || req.TaskID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
        return
    }
    _, err := db.Exec(`INSERT INTO task_completions (device_id, task_date, task_id, completed)
        VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE completed=VALUES(completed)`,
        req.DeviceID, req.TaskDate, req.TaskID, req.Completed)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---------- 打卡 ----------
func manualCheckin(c *gin.Context) {
    var ch Checkin
    if err := c.ShouldBindJSON(&ch); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if ch.DeviceID == "" || ch.CheckinDate == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
        return
    }
    _, err := db.Exec(`INSERT IGNORE INTO checkins (device_id, checkin_date) VALUES (?,?)`, ch.DeviceID, ch.CheckinDate)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "checked in"})
}

func getCheckinHistory(c *gin.Context) {
    deviceID := c.Query("device_id")
    if deviceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "device_id required"})
        return
    }
    rows, err := db.Query(`SELECT checkin_date FROM checkins WHERE device_id=? AND checkin_date >= DATE_SUB(CURDATE(), INTERVAL 28 DAY)`, deviceID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()
    dates := []string{}
    for rows.Next() {
        var d string
        rows.Scan(&d)
        dates = append(dates, d)
    }
    c.JSON(http.StatusOK, gin.H{"dates": dates})
}

// ---------- 阶段进度 ----------
func getStageProgress(c *gin.Context) {
    deviceID := c.Query("device_id")
    if deviceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "device_id required"})
        return
    }
    rows, err := db.Query(`SELECT stage_num, days_completed FROM stage_progress WHERE device_id=?`, deviceID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()
    progress := make(map[int]int)
    for rows.Next() {
        var num, days int
        rows.Scan(&num, &days)
        progress[num] = days
    }
    c.JSON(http.StatusOK, gin.H{"stage_progress": progress})
}

func updateStageProgress(c *gin.Context) {
    var sp StageProgress
    if err := c.ShouldBindJSON(&sp); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    _, err := db.Exec(`INSERT INTO stage_progress (device_id, stage_num, days_completed) VALUES (?,?,?)
        ON DUPLICATE KEY UPDATE days_completed=VALUES(days_completed)`,
        sp.DeviceID, sp.StageNum, sp.DaysCompleted)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// ---------- 访问统计 ----------
func recordVisit(c *gin.Context) {
    deviceID := c.PostForm("device_id")
    ip := c.GetHeader("X-Real-IP")
    if ip == "" {
        ip = c.GetHeader("X-Forwarded-For")
    }
    if ip == "" {
        ip = c.ClientIP()
    }
    userAgent := c.GetHeader("User-Agent")
    db.Exec(`INSERT INTO visits (ip, device_id, user_agent) VALUES (?,?,?)`, ip, deviceID, userAgent)
    c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

func getStats(c *gin.Context) {
    var totalVisits, todayVisits int64
    db.QueryRow("SELECT COUNT(*) FROM visits").Scan(&totalVisits)
    db.QueryRow("SELECT COUNT(*) FROM visits WHERE DATE(visit_time) = CURDATE()").Scan(&todayVisits)

    rows, err := db.Query(`SELECT id, ip, COALESCE(device_id,''), visit_time, user_agent FROM visits ORDER BY id DESC LIMIT 20`)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()
    recent := []Visit{}
    for rows.Next() {
        var v Visit
        var t time.Time
        rows.Scan(&v.ID, &v.IP, &v.DeviceID, &t, &v.UserAgent)
        v.VisitTime = t.Format("2006-01-02 15:04:05")
        recent = append(recent, v)
    }
    c.JSON(http.StatusOK, Stats{TotalVisits: totalVisits, TodayVisits: todayVisits, RecentVisits: recent})
}

// ---------- AI 诊断（云端 DeepSeek） ----------
func aiDiagnose(c *gin.Context) {
    var req struct {
        DeviceID string  `json:"device_id"`
        Height   float64 `json:"height"`
        Weight   float64 `json:"weight"`
        BodyFat  float64 `json:"body_fat"`
        Age      int     `json:"age"`
        Symptoms string  `json:"symptoms"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
        return
    }
    prompt := fmt.Sprintf(`你是一位富有同理心的健康恢复教练。请根据以下用户数据，给出简短的诊断和第一阶段恢复建议（用中文，不超过300字）。

数据：
- 身高：%.0f cm
- 体重：%.1f kg
- 体脂率：%.1f%%
- 年龄：%d 岁
- 自述症状：%s

要求：
1. 用一句话总结用户当前的主要问题（如“疲劳型脂肪堆积”）。
2. 给出三个最容易做到的恢复行动（不涉及减肥，只关注恢复能量）。
3. 语气温暖、鼓励，不制造焦虑。
4. 直接输出结果，不要额外说明。`, req.Height, req.Weight, req.BodyFat, req.Age, req.Symptoms)

    aiResp, err := callDeepSeek(prompt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "AI服务异常: " + err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"analysis": aiResp})
}