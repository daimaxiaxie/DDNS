# DDNS
DDNS
# Support
|   Cloud   | Support  |
|:---------:|:--------:|
| Aliyun V1 | &#x2705; |
| Aliyun V2 | &#x274C; |
|  Tencent  | &#x274C; |
|  Huawei   | &#x274C; |

# Usage
### windows
exe and config in same director. Double click to run. Or
`ddns.exe start --configPath=config.json`
### Linux
`./ddns start --configPath=./config.json`

`./ddns stop`

# Config
#### example dev.example.com
```
{
  "cloud": "aliyun",
  "version": 1,
  "duration": 60,
  "extra": {
    "accessKeyID": "key id",
    "accessKeySecret": "key secret",
    "recordID": "81000000000000000",
    "recordType": "AAAA",
    "rr": "dev",
    "value": "",
    "domain": "example.com",
    "netInterface": "WLAN",
    "resolveMode": "RandomOnly"
  }
}
```
duration : minute<br/>
recordID : allow null