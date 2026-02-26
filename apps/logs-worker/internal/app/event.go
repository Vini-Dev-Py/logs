package app

type Event struct {
	EventID      string         `json:"eventId"`
	CompanyID    string         `json:"companyId"`
	TraceID      string         `json:"traceId"`
	NodeID       string         `json:"nodeId"`
	ParentNodeID *string        `json:"parentNodeId"`
	ServiceName  string         `json:"serviceName"`
	Operation    map[string]any `json:"operation"`
	HTTP         map[string]any `json:"http"`
	DB           map[string]any `json:"db"`
	Metadata     map[string]any `json:"metadata"`
}
