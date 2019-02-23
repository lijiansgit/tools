package handler

import (
	"context"
	"encoding/json"

	log "github.com/alecthomas/log4go"
	"github.com/lijiansgit/tools/automation/config"
	"github.com/lijiansgit/tools/automation/models"

	domain "github.com/lijiansgit/tools/automation/proto/domain"
)

// Domain struct
type Domain struct {
	Conf  *config.Config
	Model *models.Model
}

// NewDomain new struct
func NewDomain(c *config.Config, m *models.Model) *Domain {
	return &Domain{
		Conf:  c,
		Model: m,
	}
}

// Call 域名查询
func (d *Domain) Call(ctx context.Context, req *domain.Request, rsp *domain.Response) error {
	log.Debug("Received Domain.Call request")
	rsp.Msg = "test"
	if req.Operation != "domain_op" {
		return nil
	}

	res, err := d.domainOpQeury()
	if err != nil {
		return err
	}

	rsp.Msg = res
	return nil
}

func (d *Domain) domainOpQeury() (res string, err error) {
	var dp []*models.DomainOp
	if _, err = d.Model.Orm.QueryTable("domain_op").All(&dp); err != nil {
		return res, err
	}

	b, err := json.Marshal(dp)
	if err != nil {
		return res, err
	}

	res = string(b)
	log.Info("res: %s", res)
	return res, nil
}
