package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSSESReceiptFilter struct {
}

func init() {
	i := AWSSESReceiptFilter{}
	listers = append(listers, i)
}

func (l AWSSESReceiptFilter) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SesReceiptFilter,
	}
}

func (l AWSSESReceiptFilter) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ses.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	filters, err := svc.ListReceiptFilters(ctx.Context, &ses.ListReceiptFiltersInput{})
	if err != nil {
		if strings.Contains(err.Error(), "Unavailable Operation") {
			// If SES isn't available in a region, returns Unavailable Operation error
			return rg, nil
		}
		return rg, err
	}
	for _, v := range filters.Filters {
		r := resource.New(ctx, resource.SesReceiptFilter, v.Name, v.Name, v)
		rg.AddResource(r)
	}
	return rg, err
}
