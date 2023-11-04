package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/emr"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSEMRSecurityConfiguration struct {
}

func init() {
	i := AWSEMRSecurityConfiguration{}
	listers = append(listers, i)
}

func (l AWSEMRSecurityConfiguration) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.EmrSecurityConfiguration}
}

func (l AWSEMRSecurityConfiguration) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := emr.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListSecurityConfigurations(ctx.Context, &emr.ListSecurityConfigurationsInput{})
		if err != nil {
			return nil, err
		}
		for _, id := range res.SecurityConfigurations {
			v, err := svc.DescribeSecurityConfiguration(ctx.Context, &emr.DescribeSecurityConfigurationInput{
				Name: id.Name,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe security config %s: %w", *id.Name, err)
			}
			r := resource.New(ctx, resource.EmrSecurityConfiguration, v.Name, v.Name, v)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
