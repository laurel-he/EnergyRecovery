package main

import (
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

// ---------- 认证 ----------
func register(c *gin.Context) {
    var req LoginReq
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
        return
    }
    if req.Username == "" || req.Password == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
        return
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
        return
    }

    _, err = db.Exec("INSERT INTO accounts (username, password_hash) VALUES (?, ?)", req.Username, string(hash))
    if err != nil {
        c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
        return
    }

    // 同时创建空的健康档案
    var accountID int
    db.QueryRow("SELECT id FROM accounts WHERE username=?", req.Username).Scan(&accountID)
    db.Exec("INSERT IGNORE INTO user_profile (account_id) VALUES (?)", accountID)

    // 生成token
    token, _ := generateToken(accountID, req.Username)
    c.JSON(http.StatusOK, gin.H{"token": token, "account_id": accountID, "username": req.Username})
}

func login(c *gin.Context) {
    var req LoginReq
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
        return
    }

    var accountID int
    var hash string
    err := db.QueryRow("SELECT id, password_hash FROM accounts WHERE username=?", req.Username).Scan(&accountID, &hash)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
        return
    }

    if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
        return
    }

    token, _ := generateToken(accountID, req.Username)
    c.JSON(http.StatusOK, gin.H{"token": token, "account_id": accountID, "username": req.Username})
}

func generateToken(accountID int, username string) (string, error) {
    claims := Claims{
        AccountID: accountID,
        Username:  username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(JWT_SECRET))
}

// 认证中间件
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenStr := c.GetHeader("Authorization")
        if len(tokenStr) < 7 || tokenStr[:7] != "Bearer " {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
            c.Abort()
            return
        }
        tokenStr = tokenStr[7:]

        claims := &Claims{}
        token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
            return []byte(JWT_SECRET), nil
        })
        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期"})
            c.Abort()
            return
        }
        c.Set("account_id", claims.AccountID)
        c.Set("username", claims.Username)
        c.Next()
    }
}

// ---------- 用户档案 ----------
func getProfile(c *gin.Context) {
    accountID := c.GetInt("account_id")
    var p UserProfile
    err := db.QueryRow(`SELECT COALESCE(height,0), COALESCE(weight,0), COALESCE(body_fat,0),
        COALESCE(age,0), COALESCE(symptoms,''), COALESCE(diagnosis_type,''), COALESCE(diagnosis_label,''),
        COALESCE(micro_death_index,0), COALESCE(micro_death_level,''), COALESCE(vitality_score,44),
        COALESCE(current_stage,1), COALESCE(stage_start_date,'')
        FROM user_profile WHERE account_id=?`, accountID).
        Scan(&p.Height, &p.Weight, &p.BodyFat, &p.Age, &p.Symptoms,
            &p.DiagnosisType, &p.DiagnosisLabel, &p.MicroDeathIndex, &p.MicroDeathLevel,
            &p.VitalityScore, &p.CurrentStage, &p.StageStartDate)
    if err != nil {
        // 创建默认档案
        db.Exec("INSERT IGNORE INTO user_profile (account_id) VALUES (?)", accountID)
        p = UserProfile{VitalityScore: 44, CurrentStage: 1}
    }
    c.JSON(http.StatusOK, p)
}

func updateProfile(c *gin.Context) {
    accountID := c.GetInt("account_id")
    var p UserProfile
    if err := c.ShouldBindJSON(&p); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    _, err := db.Exec(`INSERT INTO user_profile (account_id, height, weight, body_fat, age, symptoms,
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
        accountID, p.Height, p.Weight, p.BodyFat, p.Age, p.Symptoms,
        p.DiagnosisType, p.DiagnosisLabel, p.MicroDeathIndex, p.MicroDeathLevel,
        p.VitalityScore, p.CurrentStage, p.StageStartDate)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// ---------- 任务 ----------
func getTodayTasks(c *gin.Context) {
    accountID := c.GetInt("account_id")
    date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

    rows, _ := db.Query("SELECT task_id, completed FROM task_completions WHERE account_id=? AND task_date=?", accountID, date)
    defer rows.Close()
    taskMap := map[string]bool{}
    for rows.Next() {
        var id string
        var comp bool
        rows.Scan(&id, &comp)
        taskMap[id] = comp
    }

    tasks := []map[string]interface{}{
        {"id": "water", "name": "喝够8杯水", "icon": "💧", "hint": "约1.6L，少量多次", "completed": taskMap["water"]},
        {"id": "protein", "name": "早餐吃一份蛋白质", "icon": "🥚", "hint": "鸡蛋/牛奶/豆制品", "completed": taskMap["protein"]},
        {"id": "sleep", "name": "23:30前躺下睡觉", "icon": "🌙", "hint": "放下手机，闭眼即可", "completed": taskMap["sleep"]},
        {"id": "walk", "name": "散步10分钟", "icon": "🚶", "hint": "不要求速度，出门就行", "completed": taskMap["walk"]},
    }
    c.JSON(http.StatusOK, tasks)
}

func toggleTask(c *gin.Context) {
    accountID := c.GetInt("account_id")
    var req TaskCompletion
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if req.TaskDate == "" || req.TaskID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数"})
        return
    }
    _, err := db.Exec(`INSERT INTO task_completions (account_id, task_date, task_id, completed)
        VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE completed=VALUES(completed)`,
        accountID, req.TaskDate, req.TaskID, req.Completed)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---------- 打卡 ----------
func manualCheckin(c *gin.Context) {
    accountID := c.GetInt("account_id")
    var ch Checkin
    if err := c.ShouldBindJSON(&ch); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if ch.CheckinDate == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "日期不能为空"})
        return
    }
    _, err := db.Exec("INSERT IGNORE INTO checkins (account_id, checkin_date) VALUES (?,?)", accountID, ch.CheckinDate)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "checked in"})
}

func getCheckinHistory(c *gin.Context) {
    accountID := c.GetInt("account_id")
    rows, _ := db.Query("SELECT checkin_date FROM checkins WHERE account_id=? AND checkin_date >= DATE_SUB(CURDATE(), INTERVAL 28 DAY)", accountID)
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
    accountID := c.GetInt("account_id")
    rows, _ := db.Query("SELECT stage_num, days_completed FROM stage_progress WHERE account_id=?", accountID)
    defer rows.Close()
    progress := map[int]int{}
    for rows.Next() {
        var num, days int
        rows.Scan(&num, &days)
        progress[num] = days
    }
    c.JSON(http.StatusOK, gin.H{"stage_progress": progress})
}

func updateStageProgress(c *gin.Context) {
    accountID := c.GetInt("account_id")
    var sp StageProgress
    if err := c.ShouldBindJSON(&sp); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    _, err := db.Exec(`INSERT INTO stage_progress (account_id, stage_num, days_completed) VALUES (?,?,?)
        ON DUPLICATE KEY UPDATE days_completed=VALUES(days_completed)`,
        accountID, sp.StageNum, sp.DaysCompleted)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

// ---------- 访问统计 ----------
func recordVisit(c *gin.Context) {
    ip := c.ClientIP()
    userAgent := c.GetHeader("User-Agent")
    accountID := c.GetInt("account_id") // 可能为0（未登录）
    var uid interface{}
    if accountID > 0 {
        uid = accountID
    }
    db.Exec("INSERT INTO visits (ip, account_id, user_agent) VALUES (?,?,?)", ip, uid, userAgent)
    c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}

func getStats(c *gin.Context) {
    var total, today int64
    db.QueryRow("SELECT COUNT(*) FROM visits").Scan(&total)
    db.QueryRow("SELECT COUNT(*) FROM visits WHERE DATE(visit_time)=CURDATE()").Scan(&today)
    rows, _ := db.Query("SELECT id, ip, COALESCE(account_id,0), visit_time, user_agent FROM visits ORDER BY id DESC LIMIT 20")
    defer rows.Close()
    recent := []Visit{}
    for rows.Next() {
        var v Visit
        var t time.Time
        rows.Scan(&v.ID, &v.IP, &v.AccountID, &t, &v.UserAgent)
        v.VisitTime = t.Format("2006-01-02 15:04:05")
        recent = append(recent, v)
    }
    c.JSON(http.StatusOK, Stats{TotalVisits: total, TodayVisits: today, RecentVisits: recent})
}

// ---------- AI 诊断 ----------
func aiDiagnose(c *gin.Context) {
    var req struct {
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
1. 用一句话总结用户当前的主要问题。
2. 给出三个最容易做到的恢复行动（不涉及减肥）。
3. 语气温暖、鼓励。
4. 直接输出结果。`, req.Height, req.Weight, req.BodyFat, req.Age, req.Symptoms)

    aiResp, err := callDeepSeek(prompt)
    if err != nil {
        // 降级方案：返回固定温暖提示
        fallback := "🤗 你好，我是你的能量恢复伙伴。虽然今天无法进行智能分析，但请记住：你的身体不需要急于求成，先恢复能量比什么都重要。试着从今天喝够水、早睡10分钟开始，一点一点来。"
        c.JSON(http.StatusOK, gin.H{"analysis": fallback})
        return
    }
    c.JSON(http.StatusOK, gin.H{"analysis": aiResp})
}