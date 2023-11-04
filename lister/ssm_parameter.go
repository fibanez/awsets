package lister

import (
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSSsmParameter struct {
}

func init() {
	i := AWSSsmParameter{}
	listers = append(listers, i)
}

func (l AWSSsmParameter) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SsmParameter}
}

func (l AWSSsmParameter) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ssm.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeParameters(ctx.Context, &ssm.DescribeParametersInput{
			MaxResults: 50,
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, parameter := range res.Parameters {
			r := resource.New(ctx, resource.SsmParameter, parameter.Name, parameter.Name, parameter)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
