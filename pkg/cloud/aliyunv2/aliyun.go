package aliyunv2

import (
	"DDNS/pkg/common"
	"DDNS/pkg/dnsutil"
	"encoding/json"
	"fmt"
	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type Aliyun struct {
	common.CloudInfo
	config Config
	client *alidns20150109.Client
}

func init() {
	ali := &Aliyun{
		CloudInfo: common.CloudInfo{
			Id:      ID,
			Name:    Name,
			Version: Version,
		},
	}
	common.Manager.Register(ali)
}

func (ali *Aliyun) Init(config []byte) error {
	ali.config.Region = DefaultRegion
	ali.config.Line, ali.config.Lang, ali.config.UserClientIp, ali.config.TTL, ali.config.Priority = DefaultLine, DefaultLang, DefaultIP, DefaultTTL, DefaultPriority

	if err := json.Unmarshal(config, &ali.config); err != nil {
		return fmt.Errorf("incorrect config for %s", ali.Name)
	}

	apiConfig := &openapi.Config{
		AccessKeyId:     tea.String(ali.config.AccessKeyID),
		AccessKeySecret: tea.String(ali.config.AccessKeySecret),
		RegionId:        tea.String(ali.config.Region),
	}
	var err error
	ali.client, err = alidns20150109.NewClient(apiConfig)
	if err != nil {
		return err
	}
	return nil
}

func (ali *Aliyun) Update() error {
	//TODO implement me
	panic("implement me")
}

func (ali *Aliyun) Stop() {
	//TODO implement me
	panic("implement me")
}

func (ali *Aliyun) Info() common.CloudInfo {
	return ali.CloudInfo
}

func (ali *Aliyun) GetALLRecord() (*alidns20150109.DescribeDomainRecordsResponseBodyDomainRecords, error) {
	request := &alidns20150109.DescribeDomainRecordsRequest{
		RRKeyWord:  tea.String(ali.config.RR),
		PageSize:   tea.Int64(DefaultPageSize),
		DomainName: tea.String(ali.config.Domain),
	}

	runtime := &util.RuntimeOptions{}
	response, err := ali.client.DescribeDomainRecordsWithOptions(request, runtime)
	if err != nil {
		return nil, err
	}

	if len(response.Body.DomainRecords.Record) == 0 {
		return nil, fmt.Errorf("get all record response error : %s", response.Body)
	}
	return response.Body.DomainRecords, nil
}

func (ali *Aliyun) UpdateRecord(value string) error {
	request := &alidns20150109.UpdateDomainRecordRequest{}
	request.RecordId = tea.String(ali.config.RecordID)
	request.RR = tea.String(ali.config.RR)
	request.Type = tea.String(ali.config.RecordType)

	if dnsutil.Equal(ali.config.Value, "") {
		request.Value = tea.String(value)
	} else {
		request.Value = tea.String(ali.config.Value)
	}
	if dnsutil.Unequal(ali.config.Line, DefaultLine) {
		request.Line = tea.String(ali.config.Line)
	}
	if dnsutil.Unequal(ali.config.Lang, DefaultLang) {
		request.Lang = tea.String(ali.config.Lang)
	}
	if dnsutil.Unequal(ali.config.TTL, DefaultTTL) {
		request.TTL = tea.Int64(ali.config.TTL)
	}
	if dnsutil.Unequal(ali.config.Priority, DefaultPriority) {
		request.Priority = tea.Int64(ali.config.Priority)
	}

	runtime := &util.RuntimeOptions{}

	response, err := ali.client.UpdateDomainRecordWithOptions(request, runtime)
	if err != nil {
		return err
	}
	fmt.Println(response.Body)
	return nil
}
