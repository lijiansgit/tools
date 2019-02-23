package models

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/lijiansgit/tools/automation/config"
	// mysql
	_ "github.com/go-sql-driver/mysql"
)

// DNSPodData post 数据
type DNSPodData struct {
	ID          int    `orm:"column(id)"`
	MonitorID   int    `orm:"column(monitor_id);unique"`
	DomainID    string `orm:"column(domain_id)"`
	DomainName  string `orm:"column(domain_name)"`
	RecordID    string `orm:"column(record_id)"`
	RecordName  string
	RecordLine  string
	RecordValue string
	Status      string
	StatusCode  string
	Reason      string
	UpdateTime  string
}

// TableName 自定义设置表明
func (d *DNSPodData) TableName() string {
	return "dnspod_data"
}

// CloudwiseData 监控宝post数据
type CloudwiseData struct {
	ID          int    `orm:"column(id)"`
	TaskID      int    `orm:"column(task_id);unique"`
	TaskName    string `orm:"column(task_name)"`
	TaskType    string
	TaskSummary string
	MsgID       string `orm:"column(msg_id)"`
	MsgStatus   string
	FaultTime   string
}

// 域名操作记录
type DomainOp struct {
	ID          int    `orm:"column(id)"`
	OpType      bool   `orm:"column(op_type)"`
	DomainName  string `orm:"column(domain_name)"`
	RecordName  string
	RecordType  string
	RecordLine  string
	RecordValue string
	UpdateTime  int
}

// Model struct
type Model struct {
	Conf *config.Config
	Orm  orm.Ormer
}

// NewModel init
func NewModel(conf *config.Config) (m *Model, err error) {
	// orm.Log = orm.NewLog(log.LogWriter)
	orm.Debug = true

	// set default database
	ds := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8",
		conf.MySQL.User, conf.MySQL.Passwd, conf.MySQL.Addr, conf.MySQL.DB)
	err = orm.RegisterDataBase("default", "mysql", ds)
	if err != nil {
		return m, err
	}

	// register model
	orm.RegisterModel(new(DNSPodData))
	orm.RegisterModel(new(CloudwiseData))
	orm.RegisterModel(new(DomainOp))

	// create table
	err = orm.RunSyncdb("default", false, true)
	if err != nil {
		return m, err
	}

	m = &Model{
		Conf: conf,
		Orm:  orm.NewOrm(),
	}

	return m, nil
}
