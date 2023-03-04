package aliyunv2

const ID int8 = 1
const Version int8 = 2
const Name string = "aliyun"

const DefaultPageSize = 100

const DefaultRegion = "public"

const DefaultScheme = "https"
const DefaultLine = "default"
const DefaultLang = "en"
const DefaultIP = "127.0.0.1"
const DefaultTTL = 600
const DefaultPriority = 1

const FirstOnly = "FirstOnly"
const RandomOnly = "RandomOnly"
const MultiIP = "ALL"
const DefaultMode = FirstOnly
const DNSEnable = "Enable"

type Config struct {
	Region          string `json:"region,omitempty"`
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`

	ResolveMode string `json:"resolveMode,omitempty"` //FirstOnly, RandomOnly, ALL

	RecordID     string `json:"recordID"`
	RecordType   string `json:"recordType"`
	RR           string `json:"rr"`
	Value        string `json:"value,omitempty"`
	NetInterface string `json:"netInterface"`
	Line         string `json:"line,omitempty"`
	Lang         string `json:"lang,omitempty"`
	UserClientIp string `json:"userClientIp,omitempty"`
	TTL          int64  `json:"ttl,omitempty"`
	Priority     int64  `json:"priority,omitempty"`

	Domain string `json:"domain,omitempty"`
}
