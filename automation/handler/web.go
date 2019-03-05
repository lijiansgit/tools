// web 服务
// 1. 监控宝回调POST: https://help.cloudwise.com/help/18/21/114
// 2. dnspod 回调：https://support.dnspod.cn/Kb/showarticle/tsid/24/

package handler

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/alecthomas/log4go"
	"github.com/go-martini/martini"
	"github.com/lijiansgit/go/libs/dnspod"
	"github.com/lijiansgit/tools/automation/config"
	"github.com/lijiansgit/tools/automation/models"
)

const (
	// ErrContentCode 内容错误
	ErrContentCode = 550
	// ErrContentCodeRes str
	ErrContentCodeRes = "content error"
	// ErrTokenCode token校验失败
	ErrTokenCode = 551
	// ErrTokenCodeRes str
	ErrTokenCodeRes = "token check error"
	// DNSPodRecordDown dnspod 记录故障状态
	DNSPodRecordDown = "Down"
	// DNSPodRecordOk dnspod 故障恢复
	DNSPodRecordOk = "Ok"
	// CloudwiseRecordDown cloudwise 报警状态
	CloudwiseRecordDown = "1"
	// CloudwiseRecordOk cloudwise 报警恢复
	CloudwiseRecordOk = "2"
	// DefaultRecordType 默认类型
	DefaultRecordType = "A"
)

// Web 结构体
type Web struct {
	Conf  *config.Config
	Ma    *martini.Martini
	Model *models.Model
}

// NewWeb 新Web
func NewWeb(conf *config.Config, model *models.Model) *Web {
	web := new(Web)
	web.Conf = conf
	web.Model = model
	web.Ma = martini.New()
	web.Ma.Use(web.logger())
	web.Ma.Action(web.router().Handle)
	return web
}

// Start 启动web服务
func (w *Web) Start() {
	mux := http.NewServeMux()
	mux.Handle("/", w.Ma)
	server := &http.Server{
		Addr:    w.Conf.Web.Addr,
		Handler: mux,
	}

	log.Info("HTTP Server Start: %s", w.Conf.Web.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Error("HTTP Server Start err(%v)", err)
		return
	}
}

func (w *Web) logger() martini.Handler {
	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		start := time.Now()

		addr := req.Header.Get("X-Rea1-IP")
		if addr == "" {
			addr = req.Header.Get("X-Forwarded-For")
		}
		if addr == "" {
			addr = req.RemoteAddr
		}
		rw := res.(martini.ResponseWriter)
		c.Next()

		log.Debug("REQ - %s - %s - %s - %v - %s - %v",
			req.Method, req.URL.Path, addr, rw.Status(),
			http.StatusText(rw.Status()), time.Since(start))
	}
}

func (w *Web) router() martini.Router {
	r := martini.NewRouter()
	r.Get(w.Conf.Web.StatusURL, w.status)
	r.Post(w.Conf.Web.CloudwiseURL, w.cloudwise)
	r.Post(w.Conf.Web.DNSPodURL, w.dnspod)
	return r
}

func (w *Web) status() string {
	return "ok"
}

func (w *Web) cloudwise(req *http.Request) (code int, res string) {
	code, res = 200, "ok"
	err := req.ParseMultipartForm(10240)
	if err != nil {
		log.Warn("cloudwise req.ParseMultipartForm warn(%v)", err)
		return ErrContentCode, ErrContentCodeRes
	}

	log.Debug("cloudwise req: %v", req)
	p := req.PostForm
	c := new(models.CloudwiseData)
	c.TaskID, err = strconv.Atoi(p.Get("task_id"))
	if err != nil {
		log.Warn("cloudwise strconv.Atoi warn(%v)", err)
		return ErrContentCode, ErrContentCodeRes
	}

	c.TaskName = p.Get("task_name")
	c.TaskType = p.Get("task_type")
	c.TaskSummary = p.Get("task_summary")
	c.MsgID = p.Get("msg_id")
	c.MsgStatus = p.Get("message_status")
	c.FaultTime = p.Get("fault_time")
	tokens := c.MsgID + strconv.Itoa(c.TaskID) + c.FaultTime + w.Conf.Cloudwise.Token
	token := fmt.Sprintf("%x", md5.Sum([]byte(tokens)))
	if token != p.Get("token") {
		log.Warn("cloudwise token check fail: %s", token)
		return ErrTokenCode, ErrTokenCodeRes
	}

	log.Info("cloudwise token check ok: %s", token)
	log.Info("received cloudwise callback: %s,%s", c.TaskSummary, c.MsgStatus)

	if _, err := w.Model.Orm.InsertOrUpdate(c, "task_id"); err != nil {
		log.Warn("cloudwise InsertOrUpdate warn(%v)", err)
	}

	if c.MsgStatus == CloudwiseRecordOk {
		go w.recordStartCloudwise(c)
	}

	// TODO 待监控宝方增加故障IP参数

	return code, res
}

func (w *Web) dnspod(req *http.Request) (code int, res string) {
	code, res = 200, "ok"
	err := req.ParseForm()
	if err != nil {
		log.Warn("dnspod req.ParseForm warn(%v)", err)
		return ErrContentCode, ErrContentCodeRes
	}

	log.Debug("dnspod req: %v", req)
	p := req.PostForm
	d := new(models.DNSPodData)
	d.MonitorID, err = strconv.Atoi(p.Get("monitor_id"))
	if err != nil {
		log.Warn("dnspod strconv.Atoi warn(%v)", err)
		return ErrContentCode, ErrContentCodeRes
	}

	d.DomainID = p.Get("domain_id")
	d.DomainName = p.Get("domain")
	d.RecordID = p.Get("record_id")
	d.RecordName = p.Get("sub_domain")
	d.RecordLine = p.Get("record_line")
	d.RecordValue = p.Get("ip")
	d.Status = p.Get("status")
	d.StatusCode = p.Get("status_code")
	d.Reason = p.Get("reason")
	d.UpdateTime = p.Get("created_at")
	tokens := strconv.Itoa(d.MonitorID) + d.DomainID + d.RecordID + w.Conf.DNSPod.CallBackToken + d.UpdateTime
	token := fmt.Sprintf("%x", md5.Sum([]byte(tokens)))
	if token != p.Get("checksum") {
		log.Warn("dnspod token check fail: %s", token)
		return ErrTokenCode, ErrTokenCodeRes
	}

	log.Info("dnspod token check ok: %s", token)
	domains := d.RecordName + "." + d.DomainName
	log.Info("received dnspod callback %s,%s,%s",
		domains, d.RecordValue, d.Status)

	if _, err := w.Model.Orm.InsertOrUpdate(d, "monitor_id"); err != nil {
		log.Warn("dnspod InsertOrUpdate warn(%v)", err)
	}

	if d.Status == DNSPodRecordDown {
		go w.recordStatusSet(d, false)
	}

	return code, res
}

func (w *Web) recordStartCloudwise(c *models.CloudwiseData) {
	// tasksummary = "http://test.abc.com/test"
	list := strings.Split(c.TaskSummary, "://")
	if len(list) < 2 {
		return
	}

	domains := strings.Split(list[1], "/")[0]
	domainsList := strings.SplitN(domains, ".", 2)
	if len(domainsList) < 2 {
		return
	}

	dp := new(models.DomainOp)
	dp.RecordName = domainsList[0]
	dp.DomainName = domainsList[1]

	qs := w.Model.Orm.QueryTable(dp).Filter("record_name", dp.RecordName).Filter(
		"domain_name", dp.DomainName).Filter("op_type", 0)
	if !qs.Exist() {
		log.Warn("start record is no in table domain_op: %s", domains)
		return
	}

	if err := qs.OrderBy("-id").One(dp); err != nil {
		log.Error("qs.OrderBy().One() err(%v)", err)
		return
	}

	d := new(models.DNSPodData)
	d.DomainName = dp.DomainName
	d.RecordName = dp.RecordName
	d.RecordLine = dp.RecordLine
	d.RecordValue = dp.RecordValue

	w.recordStatusSet(d, true)
}

// recordStatusSet 暂停或者开启记录
func (w *Web) recordStatusSet(d *models.DNSPodData, enabled bool) {
	if !w.configCheck(d) {
		return
	}

	dns := dnspod.NewDomain(w.Conf.DNSPod.Token, d.DomainName)
	dns.SetRecord(d.RecordName, DefaultRecordType, d.RecordLine, d.RecordValue)
	domains := d.RecordName + "." + d.DomainName
	retry := 10
	for i := 1; i <= retry; i++ {
		log.Info("RecordStatusSet retry %d/%d sleep 1m: %s,%s,%s,%v",
			i, retry, domains, d.RecordLine, d.RecordValue, enabled)
		time.Sleep(time.Minute)
		if enabled {
			if !w.cloudwiseStatusCheck(d) {
				log.Warn("cloudwise is fault, can't start record: %s", domains)
				continue
			}
		} else {
			if w.dnspodStatusCheck(d) || w.cloudwiseStatusCheck(d) {
				log.Warn("dnspod or cloudwise is ok, can't stop record: %s", domains)
				continue
			}

			enabledNum, err := w.enabledRecordLineCheck(d)
			if err != nil {
				log.Error("dns.RecordQuery err(%v)", err)
				continue
			}

			if enabledNum < 2 {
				log.Warn("dnspod enabled record line less two: %s, %v", domains, enabled)
				continue
			}
		}

		record, err := dns.RecordQuery()
		if err != nil {
			log.Error("dns.RecordQuery err(%v)", err)
			continue
		}

		if (enabled && record.Enabled == "1") || (!enabled && record.Enabled == "0") {
			log.Warn("record %s status is already: %v, exit retry", domains, enabled)
			return
		}

		// 操作记录写入数据库，提前写入，防止操作成功但写入数据库出错
		dp := new(models.DomainOp)
		dp.OpType = enabled
		dp.DomainName = d.DomainName
		dp.RecordName = d.RecordName
		dp.RecordType = DefaultRecordType
		dp.RecordLine = d.RecordLine
		dp.RecordValue = d.RecordValue
		dp.UpdateTime = int(time.Now().Unix())
		if _, err = w.Model.Orm.Insert(dp); err != nil {
			log.Error("Orm.Insert err(%v)", err)
			continue
		}

		log.Info("domain op insert ok: %s,%v", domains, enabled)

		if err = dns.RecordStatusSet(enabled); err != nil {
			log.Error("dns.RecordStatusSet err(%v)", err)
			continue
		}

		log.Info("RecordStatusSet ok: %s,%s,%v, exit retry", domains, d.RecordValue, enabled)

		// 设置域名注释
		remark := time.Now().Format(time.RFC3339)
		if err = dns.RecordRemarkSet(remark); err != nil {
			log.Warn("RecordRemarkSet warn(%v)", err)
		} else {
			log.Info("RecordRemarkSet ok: %s,%s", domains, remark)
		}

		return
	}
}

// configCheck 检查记录是否存在config中
func (w *Web) configCheck(d *models.DNSPodData) bool {
	if !ListExist(w.Conf.Domain.A.Record[d.DomainName], d.RecordName) {
		log.Warn("%s.%s is no in config", d.RecordName, d.DomainName)
		return false
	}

	recordValueList := w.Conf.Domain.A.Line[d.RecordLine]
	if len(recordValueList) < 2 {
		log.Warn("%s line less two in config", d.RecordLine)
		return false
	}

	if !ListExist(recordValueList, d.RecordValue) {
		log.Warn("%s: %s is no in config", d.RecordLine, d.RecordValue)
		return false
	}

	return true
}

// enabledRecordLineCheck 开启的线路数量
func (w *Web) enabledRecordLineCheck(d *models.DNSPodData) (num int, err error) {
	dns := dnspod.NewDomain(w.Conf.DNSPod.Token, d.DomainName)
	recordValueList := w.Conf.Domain.A.Line[d.RecordLine]
	for _, v := range recordValueList {
		dns.SetRecord(d.RecordName, DefaultRecordType, d.RecordLine, v)
		record, err := dns.RecordQuery()
		if err != nil {
			return num, err
		}

		if record.Enabled == "1" {
			num++
		}
	}

	return num, err
}

// dnspodStatusCheck 从mysql获取状态，故障返回false
func (w *Web) dnspodStatusCheck(d *models.DNSPodData) bool {
	qs := w.Model.Orm.QueryTable(d).Filter("monitor_id", d.MonitorID)
	if !qs.Exist() {
		return true
	}

	if qs.Filter("status", DNSPodRecordOk).Exist() {
		return true
	}

	return false
}

// cloudwiseStatusCheck 从mysql获取状态，故障返回false
func (w *Web) cloudwiseStatusCheck(d *models.DNSPodData) bool {
	domains := d.RecordName + "." + d.DomainName
	c := new(models.CloudwiseData)
	qs := w.Model.Orm.QueryTable(c).Filter("task_summary__icontains", domains)
	if !qs.Exist() {
		return true
	}

	if qs.Filter("msg_status", CloudwiseRecordOk).Exist() {
		return true
	}

	return false
}

// ListExist 判断k是否存在list
func ListExist(list []string, k string) bool {
	for _, v := range list {
		if v == k {
			return true
		}
	}

	return false
}
