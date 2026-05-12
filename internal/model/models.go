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
	TestUID  string `json:"test_uuid"`
	TestName string `json:"test_name"`
}
