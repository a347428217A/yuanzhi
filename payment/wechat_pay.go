package payment

import (
	"admin-api/config"
	"admin-api/models"
	"admin-api/utils"
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WechatPayClient 微信支付客户端
type WechatPayClient struct {
	AppID     string
	MchID     string
	APIKey    string
	NotifyURL string
}

var (
	wechatPayClient *WechatPayClient
)

func init() {
	// 初始化微信支付客户端
	wechatPayClient = &WechatPayClient{
		AppID:     config.Config.WechatPay.AppID,
		MchID:     config.Config.WechatPay.MchID,
		APIKey:    config.Config.WechatPay.APIKey,
		NotifyURL: config.Config.WechatPay.NotifyURL,
	}
}

// PrepayResponse 预支付响应
type PrepayResponse struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// WechatNotifyRequest 微信支付回调请求
type WechatNotifyRequest struct {
	ReturnCode    string `xml:"return_code"`
	ReturnMsg     string `xml:"return_msg"`
	AppID         string `xml:"appid"`
	MchID         string `xml:"mch_id"`
	NonceStr      string `xml:"nonce_str"`
	Sign          string `xml:"sign"`
	ResultCode    string `xml:"result_code"`
	OpenID        string `xml:"openid"`
	IsSubscribe   string `xml:"is_subscribe"`
	TradeType     string `xml:"trade_type"`
	BankType      string `xml:"bank_type"`
	TotalFee      int    `xml:"total_fee"`
	FeeType       string `xml:"fee_type"`
	CashFee       int    `xml:"cash_fee"`
	CashFeeType   string `xml:"cash_fee_type"`
	TransactionID string `xml:"transaction_id"`
	OutTradeNo    string `xml:"out_trade_no"`
	Attach        string `xml:"attach"`
	TimeEnd       string `xml:"time_end"`
}

// WechatNotifyResponse 微信支付回调响应
type WechatNotifyResponse struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
}

// CreateWechatPayOrder 创建微信支付订单
func CreateWechatPayOrder(options ...func(map[string]interface{})) (*PrepayResponse, error) {
	params := map[string]interface{}{
		"appid":            wechatPayClient.AppID,
		"mch_id":           wechatPayClient.MchID,
		"nonce_str":        generateNonceStr(32),
		"notify_url":       wechatPayClient.NotifyURL,
		"trade_type":       "JSAPI",
		"spbill_create_ip": "127.0.0.1",
	}

	// 1. 先应用所有选项
	for _, option := range options {
		option(params)
	}

	// 2. 类型转换确保total_fee是字符串
	if fee, ok := params["total_fee"].(int); ok {
		params["total_fee"] = strconv.Itoa(fee)
	}

	// 3. 必填项检查
	if params["out_trade_no"] == nil || params["total_fee"] == nil ||
		params["body"] == nil || params["openid"] == nil {
		return nil, errors.New("缺少必要参数")
	}

	// 4. 复制参数并排除sign字段
	signParams_ := make(map[string]interface{})
	for k, v := range params {
		if k != "sign" {
			signParams_[k] = v
		}
	}

	// 5. 生成签名（最后一步！）
	params["sign"] = generateSign(signParams_, wechatPayClient.APIKey)

	// 调试输出
	fmt.Println("最终签名参数:", signParams_)
	fmt.Println("生成签名:", params["sign"])

	// 转换为XML
	xmlData, err := mapToXML(params)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(xmlData))

	// 发送请求
	resp, err := sendWechatRequest("https://api.mch.weixin.qq.com/pay/unifiedorder", xmlData)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp)

	// 解析响应
	if resp["return_code"] != "SUCCESS" {
		return nil, errors.New("微信支付错误: " + resp["return_msg"])
	}

	if resp["result_code"] != "SUCCESS" {
		return nil, errors.New("微信支付业务错误: " + resp["err_code_des"])
	}

	// 构建预支付响应
	prepayResp := &PrepayResponse{
		AppID:     wechatPayClient.AppID,
		TimeStamp: fmt.Sprintf("%d", time.Now().Unix()),
		NonceStr:  generateNonceStr(32),
		Package:   "prepay_id=" + resp["prepay_id"],
		SignType:  "MD5",
	}

	// 生成支付签名
	signParams := map[string]interface{}{
		"appId":     prepayResp.AppID,
		"timeStamp": prepayResp.TimeStamp,
		"nonceStr":  prepayResp.NonceStr,
		"package":   prepayResp.Package,
		"signType":  prepayResp.SignType,
	}

	prepayResp.PaySign = generateSign(signParams, wechatPayClient.APIKey)

	return prepayResp, nil
}

// HandlePaymentResult 处理支付结果
func HandlePaymentResult(req WechatNotifyRequest) error {
	// 验证支付结果
	if req.ReturnCode != "SUCCESS" || req.ResultCode != "SUCCESS" {
		return fmt.Errorf("支付失败: %s", req.ReturnMsg)
	}

	// 查询本地支付记录
	payment, err := models.GetPaymentByOutTradeNo(req.OutTradeNo)
	if err != nil {
		return fmt.Errorf("未找到支付记录: %s", req.OutTradeNo)
	}

	// 检查金额是否一致
	if req.TotalFee != payment.Amount {
		return fmt.Errorf("金额不一致: 本地%d, 微信%d", payment.Amount, req.TotalFee)
	}

	// 获取当前时间（用于 PaidAt）
	now := time.Now()

	// 更新支付记录
	payment.Status = models.PaymentStatusSucceeded
	payment.PaidAt = &now // 使用指针赋值
	payment.TransactionID = req.TransactionID
	payment.RawNotify = utils.ToJSONString(req) // 使用新实现的工具函数

	if err := models.UpdatePayment(payment); err != nil {
		return fmt.Errorf("更新支付状态失败: %v", err)
	}

	// 更新关联预约状态
	if payment.AppointmentID != 0 {
		appointment, err := models.GetAppointmentByID(payment.AppointmentID)
		if err != nil {
			// 记录错误但不中断整个流程
			log.Printf("获取预约失败: %v", err)
		} else {
			// 更新预约状态和支付关联
			appointment.Status = models.AppointmentStatusPaid
			appointment.PaymentID = payment.ID // 关联支付记录

			if err := models.UpdateAppointment(appointment); err != nil {
				log.Printf("更新预约状态失败: %v", err)
			}
		}
	}

	return nil
}

// CreateWechatRefund 创建微信退款
func CreateWechatRefund(outTradeNo, outRefundNo string, totalFee, refundFee int, reason string) error {
	params := map[string]interface{}{
		"appid":         wechatPayClient.AppID,
		"mch_id":        wechatPayClient.MchID,
		"nonce_str":     generateNonceStr(32),
		"out_trade_no":  outTradeNo,
		"out_refund_no": outRefundNo,
		"total_fee":     totalFee,
		"refund_fee":    refundFee,
		"refund_desc":   reason,
	}

	// 生成签名
	params["sign"] = generateSign(params, wechatPayClient.APIKey)

	// 转换为XML
	xmlData, err := mapToXML(params)
	if err != nil {
		return err
	}

	// 发送退款请求
	resp, err := sendWechatRequest("https://api.mch.weixin.qq.com/secapi/pay/refund", xmlData, true)
	if err != nil {
		return err
	}

	// 解析响应
	if resp["return_code"] != "SUCCESS" {
		return errors.New("微信退款错误: " + resp["return_msg"])
	}

	if resp["result_code"] != "SUCCESS" {
		return errors.New("微信退款业务错误: " + resp["err_code_des"])
	}

	return nil
}

// VerifyWechatSign 验证微信签名
func VerifyWechatSign(req WechatNotifyRequest, apiKey string) bool {
	params := map[string]interface{}{
		"return_code":    req.ReturnCode,
		"return_msg":     req.ReturnMsg,
		"appid":          req.AppID,
		"mch_id":         req.MchID,
		"nonce_str":      req.NonceStr,
		"result_code":    req.ResultCode,
		"openid":         req.OpenID,
		"is_subscribe":   req.IsSubscribe,
		"trade_type":     req.TradeType,
		"bank_type":      req.BankType,
		"total_fee":      req.TotalFee,
		"fee_type":       req.FeeType,
		"cash_fee":       req.CashFee,
		"cash_fee_type":  req.CashFeeType,
		"transaction_id": req.TransactionID,
		"out_trade_no":   req.OutTradeNo,
		"attach":         req.Attach,
		"time_end":       req.TimeEnd,
	}

	// 移除空值
	for k, v := range params {
		if v == "" || v == 0 {
			delete(params, k)
		}
	}

	calculatedSign := generateSign(params, apiKey)
	return calculatedSign == req.Sign
}

// ========== 辅助函数 ==========

// 生成随机字符串
func generateNonceStr(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	rand.Seed(time.Now().UnixNano())
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// 生成签名
func generateSign(params map[string]interface{}, apiKey string) string {
	// 1. 只过滤sign字段，保留空值
	filtered := make(map[string]string)
	for k, v := range params {
		if k == "sign" {
			continue
		}
		// 转换为字符串（空值保留空字符串）
		strVal := fmt.Sprintf("%v", v)
		filtered[k] = strVal
	}

	// 2. 按键名ASCII排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 3. 拼接键值对（使用原始值）
	var buf bytes.Buffer
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(filtered[k]) // 原始值，不编码！
	}

	// 4. 拼接API密钥
	buf.WriteString("&key=")
	buf.WriteString(apiKey)

	// 5. MD5计算
	hasher := md5.New()
	hasher.Write(buf.Bytes())
	return strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
}

// map转XML
func mapToXML(params map[string]interface{}) (string, error) {
	buf := bytes.NewBufferString("<xml>")
	for k, v := range params {
		buf.WriteString(fmt.Sprintf("<%s>", k))
		buf.WriteString(fmt.Sprintf("%v", v))
		buf.WriteString(fmt.Sprintf("</%s>", k))
	}
	buf.WriteString("</xml>")
	return buf.String(), nil
}

// 发送微信请求
func sendWechatRequest(url, xmlData string, useCert ...bool) (map[string]string, error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// 如果需要证书
	if len(useCert) > 0 && useCert[0] {
		cert, err := tls.LoadX509KeyPair(
			config.Config.WechatPay.CertPath,
			config.Config.WechatPay.KeyPath,
		)
		if err != nil {
			return nil, fmt.Errorf("加载证书失败: %v", err)
		}

		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		}
	}

	// 发送请求
	resp, err := client.Post(url, "application/xml", bytes.NewBufferString(xmlData))
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析XML
	result := make(map[string]string)
	if err := xml.Unmarshal(body, (*mapStringString)(&result)); err != nil {
		return nil, fmt.Errorf("解析XML失败: %v", err)
	}

	return result, nil
}

// mapStringString 用于XML解析
type mapStringString map[string]string

func (m *mapStringString) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*m = map[string]string{}
	for {
		var e struct {
			XMLName xml.Name
			Content string `xml:",chardata"`
		}
		err := d.Decode(&e)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		(*m)[e.XMLName.Local] = e.Content
	}
	return nil
}

// ========== 选项模式 ==========

// OutTradeNo 设置商户订单号
func OutTradeNo(outTradeNo string) func(map[string]interface{}) {
	return func(params map[string]interface{}) {
		params["out_trade_no"] = outTradeNo
	}
}

// Amount 设置支付金额(分)
func Amount(amount int) func(map[string]interface{}) {
	return func(params map[string]interface{}) {
		params["total_fee"] = amount
	}
}

// Description 设置商品描述
func Description(desc string) func(map[string]interface{}) {
	return func(params map[string]interface{}) {
		params["body"] = desc
	}
}

// OpenID 设置用户OpenID
func OpenID(openid string) func(map[string]interface{}) {
	return func(params map[string]interface{}) {
		params["openid"] = openid
	}
}
