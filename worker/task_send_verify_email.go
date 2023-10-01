package worker

// PayloadSendVerifyEmail - contains all data of the task we want to store in redis, to be retrieved by worker from queue
type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}
