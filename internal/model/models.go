package models

type ApiAddTestReq struct {
	TestName         string  `json:"test_name"`
	StartTime        string  `json:"start_time"`
	DurationMin      int     `json:"duration"`
	TPS              float32 `json:"tps"`
	AdditionalParams string  `json:"additional_params"`
}

type ApiAddTestRes struct {
	TestUID string `json:"test_uuid"`
}

type ApiGetTestsRes []TestRes

type TestRes struct {
	TestUID   string `json:"test_uuid"`
	TestName  string `json:"test_name"`
	StartTime string `json:"start_time"`
	IsStarted bool   `json:"is_started"`
}

type ApiGetTestRes struct {
	TestName         string  `json:"test_name"`
	StartTime        string  `json:"start_time"`
	EndTime          string  `json:"end_time"`
	TPS              float32 `json:"tps"`
	AdditionalParams string  `json:"additional_params"`
	IsStarted        bool    `json:"is_started"`
}

type ApiUpdateTestReq struct {
	TestName         string  `json:"test_name" validate:"required"`
	StartTime        string  `json:"start_time" validate:"required"`
	EndTime          string  `json:"end_time" validate:"required"`
	TPS              float32 `json:"tps" validate:"required"`
	AdditionalParams string  `json:"additional_params" validate:"required"`
}

type TestDeleteMessage struct {
	TestUID string
	UserID  int
}

type ReqStartTest struct {
	TestName         string  `json:"test_name"`
	TPS              float32 `json:"tps"`
	AdditionalParams string  `json:"additional_params"`
}
