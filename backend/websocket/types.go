package websocket

type CaptureRequest struct {
	ID       string `json:"id"`
	Action   string `json:"action"`
	URL      string `json:"url"`
	Selector string `json:"selector"`
	Timeout  int    `json:"timeout,omitempty"`
}

type CaptureResponse struct {
	ID        string       `json:"id"`
	PluginID  string       `json:"pluginId,omitempty"`
	Success   bool         `json:"success"`
	Data      *CaptureData `json:"data,omitempty"`
	Error     string       `json:"error,omitempty"`
	Timestamp string       `json:"timestamp"`
}

type CaptureData struct {
	ID        string `json:"id"`
	HTML      string `json:"html"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp,omitempty"`
}

type StatusMessage struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

