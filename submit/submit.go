package submit

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/svvu/gomws/mws"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var header []string
var keys map[string]int

func init() {
	header = []string{
		"feed_product_type", //产品类型  Food
		"item_sku",          //卖家库存单位 10162513
		"brand_name",        //品牌名称 Nescafe
		"item_name",         //项目名称（又名标题）
		"manufacturer",      //制造商
		"external_product_id",
		"external_product_id_type",
		//"part_number", //制造商零件号 ABC_1234
		"recommended_browse_nodes",
		//"vintage", //酿造年份
		//"alcohol_content",                 //酒精含量
		//"alcohol_content_unit_of_measure", //酒精含量计量单位
		"standard_price", //标准价格
		"quantity",       //数量
		"main_image_url",
		"product_description", //描述
		//主图像URL
		//"item_weight",                 //物品重量
		//"item_weight_unit_of_measure", //物品重量单位
		//"batteries_required",          //这个产品是电池还是使用电池？ FALSE
		//"are_batteries_included",      //包括电池 FALSE
		//"battery_cell_composition",    //电池组成
		//"battery_type1 - battery_type3",
		//"number_of_batteries1 - number_of_batteries3",
		//"battery_weight",
		//"battery_weight_unit_of_measure",
		//"number_of_lithium_metal_cells",
		//"number_of_lithium_ion_cells",
		//"lithium_battery_packaging",
		//"lithium_battery_energy_content",
		//"lithium_battery_energy_content_unit_of_measure",
		//"lithium_battery_weight_unit_of_measure",
		//"supplier_declared_dg_hz_regulation1 - supplier_declared_dg_hz_regulation5",
		//"hazmat_united_nations_regulatory_id",
		//"safety_data_sheet_url",
		//"item_volume",
		//"item_volume_unit_of_measure",
		//"ghs_classification_class1 - ghs_classification_class3",
	}
	keys = make(map[string]int)
	for k, v := range header {
		keys[v] = k
	}
}

type Submit struct {
	*mws.Client
	config mws.Config
}

func NewClient(config mws.Config) (*Submit, error) {
	upload := new(Submit)
	base, err := mws.NewClient(config, upload.Version(), upload.Name())
	if err != nil {
		return nil, err
	}
	upload.Client = base
	upload.config = config
	return upload, nil
}

// Version return the current version of api
func (o Submit) Version() string {
	return ""
}

// Name return the name of the api
func (o Submit) Name() string {
	return ""
}

// GetServiceStatus Returns the operational status of the Orders API section.
// http://docs.developer.amazonservices.com/en_US/orders/2013-09-01/MWS_GetServiceStatus.html
func (o Submit) GetServiceStatus() (*mws.Response, error) {
	params := mws.Parameters{
		"Action": "GetServiceStatus",
	}

	return o.SendRequest(params)
}

func (o Submit) UploadProduct() (rsp *mws.Response, err error) {
	x := NewProductXml()
	x.Header.MerchantIdentifier = o.SellerId
	x.Message.MessageID = fmt.Sprintf("%v,", rand.New(rand.NewSource(time.Now().UnixNano())).Int())
	x.Message.OperationType = "Update"
	x.Message.Product.SKU = "1231231"
	x.Message.Product.DescriptionData.Title = "title"
	x.Message.Product.DescriptionData.Brand = "testbrand"
	x.Message.Product.DescriptionData.Description = "testdesc"
	x.Message.Product.DescriptionData.RecommendedBrowseNode = "1981004031"
	b, err := xml.MarshalIndent(x, "", "    ")
	if err != nil {
		fmt.Println(err)
		return
	}
	b = []byte(strings.TrimSpace(xml.Header + string(b)))
	rsp, err = o.SubmitFeed("_POST_PRODUCT_DATA_", b)
	return
}

func (o Submit) UpdatePrice() {

}

func (o Submit) UpdateImg() {

}

func (o Submit) UpdateQuantity() {

}

func (o Submit) GetFeedSubmissionList(FeedSubmissionIds ...string) (result map[string]interface{}, err error) {
	if len(FeedSubmissionIds) == 0 {
		return nil, errors.New("请输入FeedSubmissionIds")
	}
	params := mws.Parameters{
		"Action": "GetFeedSubmissionList",
	}
	for k, FeedSubmissionId := range FeedSubmissionIds {
		key := fmt.Sprintf("FeedSubmissionIdList.Id.%v", k+1)
		params[key] = FeedSubmissionId
	}
	rsp, err := o.SendRequest(params)
	if err != nil {
		return
	}
	if rsp.Error != nil {
		err = rsp.Error
		return
	}
	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}

	result = make(map[string]interface{})
	result["status"] = false
	result["error_content"] = ""
	result["NextToken"] = ""
	result["FeedSubmissionId"] = ""
	result["FeedProcessingStatus"] = ""
	res, err := mws.NewResultParser(b)
	//解析出错，则直接将body的内容返回
	if err != nil {
		err = nil
		result["status"] = false
		result["error_content"] = string(b)
		return
	}
	result["status"] = true
	result["FeedSubmissionId"], _ = res.FindByKeys("GetFeedSubmissionListResult", "FeedSubmissionInfo", "FeedSubmissionId")[0].ToString()
	result["FeedProcessingStatus"], _ = res.FindByKeys("GetFeedSubmissionListResult", "FeedSubmissionInfo", "FeedProcessingStatus")[0].ToString()
	HasNext, _ := res.FindByKeys("GetFeedSubmissionListResult", "HasNext")[0].ToBool()
	if HasNext {
		result["NextToken"], _ = res.FindByKeys("GetFeedSubmissionListResult", "NextToken")[0].ToString()
	}

	return
}

func (o Submit) GetFeedSubmissionListByNextToken(NextToken string) (result map[string]interface{}, err error) {
	if NextToken == "" {
		err = errors.New("NextToken 不能为空")
		return
	}
	params := mws.Parameters{
		"Action":    "GetFeedSubmissionListByNextToken",
		"NextToken": NextToken,
	}

	rsp, err := o.SendRequest(params)
	if err != nil {
		return
	}
	if rsp.Error != nil {
		err = rsp.Error
		return
	}
	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}

	result = make(map[string]interface{})
	result["status"] = false
	result["error_content"] = ""
	result["NextToken"] = ""
	result["FeedSubmissionId"] = ""
	result["FeedProcessingStatus"] = ""
	res, err := mws.NewResultParser(b)
	//解析出错，则直接将body的内容返回
	if err != nil {
		err = nil
		result["status"] = false
		result["error_content"] = string(b)
		return
	}
	result["status"] = true
	result["FeedSubmissionId"], _ = res.FindByKeys("GetFeedSubmissionListByNextTokenResult", "FeedSubmissionInfo", "FeedSubmissionId")[0].ToString()
	result["FeedProcessingStatus"], _ = res.FindByKeys("GetFeedSubmissionListByNextTokenResult", "FeedSubmissionInfo", "FeedProcessingStatus")[0].ToString()
	HasNext, _ := res.FindByKeys("GetFeedSubmissionListByNextTokenResult", "HasNext")[0].ToBool()
	if HasNext {
		result["NextToken"], _ = res.FindByKeys("GetFeedSubmissionListResult", "NextToken")[0].ToString()
	}
	return
}

func (o Submit) GetFeedSubmissionResult(FeedSubmissionId string) (result map[string]interface{}, err error) {
	params := mws.Parameters{
		"Action":           "GetFeedSubmissionResult",
		"FeedSubmissionId": FeedSubmissionId,
	}

	rsp, err := o.SendRequest(params)
	if err != nil {
		fmt.Println(3434)
		return
	}
	if rsp.Error != nil {
		err = rsp.Error
		return
	}
	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}

	result = make(map[string]interface{})
	result["status"] = false
	result["error_content"] = ""
	result["MessageType"] = ""
	result["ProcessingReport"] = ""
	res, err := mws.NewResultParser(b)
	//解析出错，则直接将body的内容返回
	if err != nil {
		err = nil
		result["status"] = false
		result["error_content"] = string(b)
		return
	}
	result["status"] = true
	result["MessageType"], _ = res.FindByKeys("GetFeedSubmissionListResult", "FeedSubmissionInfo", "FeedSubmissionId")[0].ToString()
	result["ProcessingReport"], _ = res.FindByKeys("GetFeedSubmissionListResult", "FeedSubmissionInfo", "FeedProcessingStatus")[0].ToString()
	return
}

var DataParseFail error = errors.New("数据解析出错")

func (o Submit) SubmitTpl(list []interface{}) (result map[string]interface{}, err error) {
	buf := NewBuffer()
	obj := NewCSV(buf)
	obj.SetDelimiter("\t")
	obj.setInputEncoding("UTF8")
	//obj.InsertOne([]string{"TemplateType=Offer", "Version=2014.0703"})
	obj.InsertOne([]string{"TemplateType=fptcustom", "Version=2019.0721"})

	obj.InsertOne(header)
	obj.InsertOne(header)
	header_len := len(header)
	for _, row := range list {
		rowData, ok := row.(map[string]interface{})
		if !ok {
			err = DataParseFail
			return
		}
		l := make([]string, header_len)
		for index, col := range rowData {
			key, ok := keys[index]
			if ok {
				l[key] = fmt.Sprintf("%v", col)
			}
		}
		obj.InsertOne(l)

	}

	obj.Flush()
	rsp, err := o.SubmitFeed("_POST_FLAT_FILE_LISTINGS_DATA_", buf.Get())
	if err != nil {
		return
	}
	if rsp.Error != nil {
		err = rsp.Error
		return
	}
	b, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return
	}

	result = make(map[string]interface{})
	result["status"] = false
	result["error_content"] = ""
	result["FeedSubmissionId"] = ""
	result["FeedType"] = ""
	result["FeedProcessingStatus"] = ""
	res, err := mws.NewResultParser(b)
	//解析出错，则直接将body的内容返回
	if err != nil {
		err = nil
		result["status"] = false
		result["error_content"] = string(b)
		return
	}
	result["status"] = true
	result["FeedSubmissionId"], _ = res.FindByKeys("SubmitFeedResult", "FeedSubmissionInfo", "FeedSubmissionId")[0].ToString()
	result["FeedType"], _ = res.FindByKeys("SubmitFeedResult", "FeedSubmissionInfo", "FeedType")[0].ToString()
	result["FeedProcessingStatus"], _ = res.FindByKeys("SubmitFeedResult", "FeedSubmissionInfo", "FeedProcessingStatus")[0].ToString()
	return
}

func (o Submit) SubmitFeed(feedType string, body []byte) (rsp *mws.Response, err error) {

	//构建请求query信息
	params := mws.Parameters{
		"Action":   "SubmitFeed",
		"FeedType": feedType,
	}

	//发送xml到mws
	rsp, err = o.SendXMl(body, params)
	return
}

func (o Submit) SendRequest(structuredParams mws.Parameters) (*mws.Response, error) {
	request, err := o.buildRequest(structuredParams)
	if err != nil {
		return nil, err
	}

	resp, err := o.Client.Do(request)
	if err != nil {
		return nil, err
	}

	return mws.NewResponse(resp), nil
}

func (o Submit) buildRequest(structuredParams mws.Parameters) (*http.Request, error) {
	params, err := structuredParams.Normalize()
	if err != nil {
		return nil, err
	}

	encodedParams := o.signQuery(params).Encode()
	req, err := http.NewRequest(
		"POST",
		"https://"+o.Host,
		bytes.NewBufferString(encodedParams),
	)

	if err != nil {
		return nil, err
	}

	// Add content headers.
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(encodedParams)))

	return req, nil
}

func (o Submit) SendXMl(body []byte, structuredParams mws.Parameters) (*mws.Response, error) {
	//构建mws xml请求
	request, err := o.buildXMLRequest(body, structuredParams)
	if err != nil {
		return nil, err
	}
	//执行请求
	resp, err := o.Client.Do(request)
	if err != nil {
		return nil, err
	}

	return mws.NewResponse(resp), nil
}

//构建请求信息
func (o Submit) buildXMLRequest(body []byte, structuredParams mws.Parameters) (*http.Request, error) {
	params, err := structuredParams.Normalize()
	if err != nil {
		return nil, err
	}
	params = o.signQuery(params)
	//signature := o.generateSignature(params)
	//params.Set("Signature", signature)
	//fmt.Println("Signature:", signature)
	encodedParams := params.Encode()
	req, err := http.NewRequest(
		"POST",
		o.EndPoint()+"?"+encodedParams,
		bytes.NewBuffer(body),
	)
	fmt.Println(encodedParams)
	if err != nil {
		return nil, err
	}

	// Add content headers.
	//req.Header.Add("Content-Type", fmt.Sprintf("%v", "text/xml; charset=iso-8859-1"))
	req.Header.Add("Content-Type", fmt.Sprintf("%v", "text/xml"))
	req.Header.Add("Accept", "application/xml")
	fmt.Println(o.Host)
	req.Header.Add("Host", o.Client.Host)
	//body = []byte("AWSAccessKeyId=AKIAIDH5ZZXVEUKJMEVA&Action=SubmitFeed&FeedType=_POST_FLAT_FILE_LISTINGS_DATA_&MWSAuthToken=amzn.mws.a4b4182e-02ba-5bc7-6b26-1ab2dd11f244&SellerId=A22SW78TJZH24R&Signature=Ys15EN1rx5iiHW7EzSpsSnJVU7nWgRx9nn%2F50QtDeqM%3D&SignatureMethod=HmacSHA256&SignatureVersion=2&Timestamp=2019-07-25T04%3A01%3A26Z&Version=2009-01-01")
	//body = []byte{}
	content_md5_str, err := content_md5(body)
	if err != nil {
		return nil, err
	}
	fmt.Println("md5-content", content_md5_str)
	req.Header.Add("Content-MD5", content_md5_str)

	return req, nil
}

func content_md5(body []byte) (content_md5 string, err error) {
	fmt.Println(string(body))
	h := md5.New()
	h.Write(body)
	content_md5 = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return
}

// signQuery generate the signature and add the signature to the http parameters.
func (o Submit) signQuery(params mws.Values) mws.Values {
	//"SellerId":         o.SellerId,
	//	"MWSAuthToken":     o.AuthToken,
	//	"SignatureMethod":  o.SignatureMethod(),
	//	"SignatureVersion": o.SignatureVersion(),
	//	"AWSAccessKeyId":   o.config.AccessKey,
	//	"Version":          "2009-01-01",
	//	"Timestamp":        time.Now().UTC().Format(time.RFC3339),
	// Add client info to the query params.
	params.Set("SellerId", o.SellerId)
	if o.AuthToken != "" {
		params.Set("MWSAuthToken", o.AuthToken)
	}
	//fmt.Println("o.config.AccessKey ", o.config.AccessKey)
	params.Set("SignatureMethod", o.SignatureMethod())
	params.Set("SignatureVersion", o.SignatureVersion())
	params.Set("AWSAccessKeyId", o.config.AccessKey)
	params.Set("Version", "2009-01-01")
	params.Set("Timestamp", time.Now().UTC().Format(time.RFC3339))

	signature := o.generateSignature(params)
	params.Set("Signature", signature)
	return params
}

// signature generate the signature by the parameters and the secretKey using HmacSHA256.
func (o Submit) generateSignature(params mws.Values) string {
	stringToSign := o.generateStringToSignV2(params)
	signature2 := mws.SignV2(stringToSign, o.config.SecretKey)
	return signature2
}

// generateStringToSignV2 Generate the string to sign for the query.
func (o Submit) generateStringToSignV2(params mws.Values) string {
	var stringToSign bytes.Buffer
	stringToSign.WriteString("POST\n")
	stringToSign.WriteString(o.Host)
	//stringToSign.WriteString("mws.amazonservices.com")
	stringToSign.WriteString("\n")
	path := o.Path()
	if path == "" {
		path = "/"
	}
	stringToSign.WriteString(path)
	stringToSign.WriteString("\n")
	stringToSign.WriteString(params.Encode())
	//stringToSign.WriteString("AWSAccessKeyId=AKIAJ3WT3MQK7V5I2BIQ&Action=SubmitFeed&ContentMD5Value=tYu6rV3g1viWVov06iTraA%3D%3D&FeedType=&MWSAuthToken=amzn.mws.a4b4182e-02ba-5bc7-6b26-1ab2dd11f244&Merchant=A1XKW8C3I72FG6&PurgeAndReplace=false&SignatureMethod=HmacSHA256&SignatureVersion=2&Timestamp=2019-07-29T07%3A42%3A23Z&Version=2009-01-01")
	//fmt.Println(params.Encode())
	return stringToSign.String()
}
