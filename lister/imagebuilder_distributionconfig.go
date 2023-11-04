package lister

import (
	"fmt"

	"github.com/fibanez/awsets/arn"

	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"

	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSImageBuilderDistributionConfig struct {
}

func init() {
	i := AWSImageBuilderDistributionConfig{}
	listers = append(listers, i)
}

func (l AWSImageBuilderDistributionConfig) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ImageBuilderDistributionConfiguration,
	}
}

func (l AWSImageBuilderDistributionConfig) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := imagebuilder.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListDistributionConfigurations(ctx.Context, &imagebuilder.ListDistributionConfigurationsInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list imagebuilder distribution configs: %w", err)
		}
		for _, config := range res.DistributionConfigurationSummaryList {
			configRes, err := svc.GetDistributionConfiguration(ctx.Context, &imagebuilder.GetDistributionConfigurationInput{
				DistributionConfigurationArn: config.Arn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get imagebuilder distribution config %s: %w", *config.Name, err)
			}
			v := configRes.DistributionConfiguration
			configArn := arn.ParseP(v.Arn)
			r := resource.New(ctx, resource.ImageBuilderDistributionConfiguration, configArn.ResourceId, v.Name, v)
			for _, dist := range v.Distributions {
				if dist.AmiDistributionConfiguration != nil {
					r.AddARNRelation(resource.KmsKey, dist.AmiDistributionConfiguration.KmsKeyId)
				}
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
