package monitor_model

type UserAgent struct {
	Model
	UserAgent   string `json:"userAgent,omitempty"`
	Description string `json:"description,omitempty"`
	AppCodeName string `json:"appCodeName,omitempty"`
	AppName     string `json:"appName,omitempty"`
	AppVersion  string `json:"appVersion,omitempty"`
	Platform    string `json:"platform,omitempty"`
	Vendor      string `json:"vendor,omitempty"`
	VendorSub   string `json:"vendorSub,omitempty"`
}
