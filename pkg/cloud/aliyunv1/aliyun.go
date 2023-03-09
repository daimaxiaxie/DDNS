package aliyunv1

import (
	"DDNS/pkg/common"
	"DDNS/pkg/dnsutil"
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"math/rand"
	"os"
	"strings"
	"time"
)

type Aliyun struct {
	common.CloudInfo
	config Config
	client *alidns.Client
}

func init() {
	ali := &Aliyun{}
	ali.defaultSetting()
	common.Manager.Register(ali)
}

func (ali *Aliyun) Init(config []byte) error {

	if err := json.Unmarshal(config, &ali.config); err != nil {
		return fmt.Errorf("incorrect config for %s", ali.Name)
	}

	if err := ali.checkConfig(); err != nil {
		return err
	}

	var err error
	ali.client, err = alidns.NewClientWithAccessKey(ali.config.Region, ali.config.AccessKeyID, ali.config.AccessKeySecret)
	if err != nil {
		return err
	}

	if ali.config.ResolveMode == FirstOnly || ali.config.ResolveMode == RandomOnly {
		if len(ali.config.RecordID) < 2 {
			records, err := ali.GetAllRecord(true)
			if err != nil {
				return err
			}
			for _, record := range records {
				if record.RR == ali.config.RR && record.Type == ali.config.RecordType && strings.ToUpper(record.Status) == DNSEnable {
					ali.config.RecordID = record.RecordID
					break
				}
			}
		}
	}

	rand.Seed(time.Now().Unix())

	return nil
}

func (ali *Aliyun) Update() error {
	ip := dnsutil.GetIP(ali.config.NetInterface, ali.IPType())

	switch ali.config.ResolveMode {
	case FirstOnly:
		value := ali.safeGetIP(ip, 0)
		if now, err := ali.GetRecord(); err != nil || value == now {
			return err
		}
		if err := ali.UpdateRecord(value); err != nil {
			return err
		}
	case RandomOnly:
		value := ali.safeGetIP(ip, rand.Intn(len(ip)))
		if now, err := ali.GetRecord(); err != nil || value == now {
			return err
		}
		if err := ali.UpdateRecord(value); err != nil {
			return err
		}
	case MultiIP:
		if err := ali.DeleteRecord(); err != nil {
			return err
		}
		for _, i := range ip {
			if err := ali.AddRecord(i); err != nil {
				return err
			}
		}
		_ = ali.OpenSLB()
	}
	return nil
}

func (ali *Aliyun) Stop() {
	//TODO implement me
	panic("implement me")
}

func (ali *Aliyun) Info() common.CloudInfo {
	return ali.CloudInfo
}

func (ali *Aliyun) defaultSetting() {
	ali.Id, ali.Name, ali.Version = ID, Name, Version
	ali.config.Region = DefaultRegion
	ali.config.Line, ali.config.Lang, ali.config.UserClientIp, ali.config.TTL, ali.config.Priority = DefaultLine, DefaultLang, DefaultIP, DefaultTTL, DefaultPriority
}

func (ali *Aliyun) checkConfig() error {
	dnsutil.MustUnequal(ali.config.AccessKeyID, "")
	dnsutil.MustUnequal(ali.config.AccessKeySecret, "")
	dnsutil.MustUnequal(ali.config.Domain, "")
	dnsutil.MustUnequal(ali.config.RecordType, "")
	dnsutil.MustUnequal(ali.config.RR, "")

	if len(ali.config.Domain) < 2 {
		return fmt.Errorf("domain error")
	}
	if len(ali.config.RecordID) > 0 && ali.config.ResolveMode == MultiIP {
		ali.config.ResolveMode = DefaultMode
		_, _ = fmt.Fprintln(os.Stderr, "resolve mode is invalid")
	}
	if ali.config.ResolveMode != FirstOnly && ali.config.ResolveMode != RandomOnly && ali.config.ResolveMode != MultiIP {
		ali.config.ResolveMode = DefaultMode
		fmt.Println("use default resolve mode")
	}
	if (ali.IPType() == dnsutil.IPV4 || ali.IPType() == dnsutil.IPV6) && len(dnsutil.GetIP(ali.config.NetInterface, ali.IPType())) == 0 {
		panic("not found ip on the interface")
	}
	return nil
}

func (ali *Aliyun) GetAllRecord(rr bool) ([]RecordDetail, error) {
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.DomainName = ali.config.Domain

	if rr {
		request.RRKeyWord = ali.config.RR
	}
	request.PageSize = requests.NewInteger(DefaultPageSize)

	responses, err := ali.client.DescribeDomainRecords(request)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	data := responses.GetHttpContentBytes()
	var records DescribeRecordResponse = DescribeRecordResponse{}
	err = json.Unmarshal(data, &records)
	if err != nil {
		return nil, err
	} else if len(records.DomainRecords.Record) == 0 {
		return nil, fmt.Errorf("get all record response error : %s", data)
	}

	return records.DomainRecords.Record, nil
}

func (ali *Aliyun) GetRecord() (string, error) {
	request := alidns.CreateDescribeDomainRecordInfoRequest()
	request.Scheme = DefaultScheme
	request.RecordId = ali.config.RecordID
	response, err := ali.client.DescribeDomainRecordInfo(request)
	if err != nil {
		return "", err
	}
	return response.Value, nil
}

func (ali *Aliyun) AddRecord(value string) error {
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = DefaultScheme

	request.DomainName = ali.config.Domain
	request.RR = ali.config.RR
	request.Type = ali.config.RecordType

	request.Value = value

	if _, err := ali.client.AddDomainRecord(request); err != nil {
		return err
	}
	fmt.Println(time.Now().Local(), " add record : ", value)
	return nil
}

func (ali *Aliyun) UpdateRecord(value string) error {
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = DefaultScheme

	request.Type = ali.config.RecordType
	request.RR = ali.config.RR
	request.RecordId = ali.config.RecordID

	if dnsutil.Equal(ali.config.Value, "") {
		request.Value = value
	} else {
		request.Value = ali.config.Value
	}
	if dnsutil.Unequal(ali.config.Line, DefaultLine) {
		request.Line = ali.config.Line
	}
	if dnsutil.Unequal(ali.config.Lang, DefaultLang) {
		request.Lang = ali.config.Lang
	}
	if dnsutil.Unequal(ali.config.TTL, DefaultTTL) {
		request.TTL = requests.NewInteger(ali.config.TTL)
	}
	if dnsutil.Unequal(ali.config.Priority, DefaultPriority) {
		request.Priority = requests.NewInteger(ali.config.Priority)
	}

	response, err := ali.client.UpdateDomainRecord(request)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(response.GetHttpContentString())
	fmt.Println(time.Now().Local(), " update record : ", request.Value, response.GetHttpContentString())
	return nil
}

func (ali *Aliyun) DeleteRecord() error {
	request := alidns.CreateDeleteSubDomainRecordsRequest()
	request.Scheme = DefaultScheme

	request.Domain = ali.config.Domain
	request.RR = ali.config.RR
	response, err := ali.client.DeleteSubDomainRecords(request)
	if err != nil {
		return err
	}
	fmt.Printf("Delete record : %v", response.GetHttpContentString())
	return nil
}

func (ali *Aliyun) OpenSLB() error {
	request := alidns.CreateSetDNSSLBStatusRequest()
	request.Scheme = DefaultScheme

	request.SubDomain = ali.config.RR + "." + ali.client.Domain
	response, err := ali.client.SetDNSSLBStatus(request)
	if err != nil {
		return err
	}
	fmt.Printf("Open SLB : %s", response.GetHttpContentString())
	return nil
}

func (ali *Aliyun) IPType() int {
	str := strings.ToUpper(ali.config.RecordType)
	switch str {
	case "A":
		return dnsutil.IPV4
	case "CNAME", "MX", "NS", "TXT", "CAA":
		return dnsutil.Other
	case "AAAA":
		return dnsutil.IPV6
	default:
		return dnsutil.ALLIP
	}
	return dnsutil.ALLIP
}

func (ali *Aliyun) safeGetIP(ip []string, index int) string {
	if len(ip) <= index {
		return ""
	}
	return ip[index]
}

func compare(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, str1 := range a {
		if str1 != b[i] {
			return false
		}
	}
	return true
}
