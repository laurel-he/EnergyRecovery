package main

type User struct {
    DeviceID         string  `json:"device_id"`
    Height           float64 `json:"height"`
    Weight           float64 `json:"weight"`
    BodyFat          float64 `json:"body_fat"`
    Age              int     `json:"age"`
    Symptoms         string  `json:"symptoms"`
    DiagnosisType    string  `json:"diagnosis_type"`
    DiagnosisLabel   string  `json:"diagnosis_label"`
    MicroDeathIndex  int     `json:"micro_death_index"`
    MicroDeathLevel  string  `json:"micro_death_level"`
    VitalityScore    int     `json:"vitality_score"`
    CurrentStage     int     `json:"current_stage"`
    StageStartDate   string  `json:"stage_start_date"`
}

type TaskCompletion struct {
    DeviceID  string `json:"device_id"`
    TaskDate  string `json:"task_date"`
    TaskID    string `json:"task_id"`
    Completed bool   `json:"completed"`
}

type Checkin struct {
    DeviceID    string `json:"device_id"`
    CheckinDate string `json:"checkin_date"`
}

type StageProgress struct {
    DeviceID      string `json:"device_id"`
    StageNum      int    `json:"stage_num"`
    DaysCompleted int    `json:"days_completed"`
}

type Visit struct {
    ID        int    `json:"id"`
    IP        string `json:"ip"`
    DeviceID  string `json:"device_id"`
    VisitTime string `json:"visit_time"`
    UserAgent string `json:"user_agent"`
}

type Stats struct {
    TotalVisits  int64   `json:"total_visits"`
    TodayVisits  int64   `json:"today_visits"`
    RecentVisits []Visit `json:"recent_visits"`
}