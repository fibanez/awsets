package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/fibanez/awsets/arn"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSSagemakerEndpointConfig struct {
}

func init() {
	i := AWSSagemakerEndpointConfig{}
	listers = append(listers, i)
}

func (l AWSSagemakerEndpointConfig) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.SagemakerEndpointConfig,
	}
}

func (l AWSSagemakerEndpointConfig) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := sagemaker.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListEndpointConfigs(ctx.Context, &sagemaker.ListEndpointConfigsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, epc := range res.EndpointConfigs {
			v, err := svc.DescribeEndpointConfig(ctx.Context, &sagemaker.DescribeEndpointConfigInput{
				EndpointConfigName: epc.EndpointConfigName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe endpoint config %s: %w", *epc.EndpointConfigName, err)
			}
			epcArn := arn.ParseP(v.EndpointConfigArn)
			r := resource.New(ctx, resource.SagemakerEndpointConfig, epcArn.ResourceId, v.EndpointConfigName, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			if v.DataCaptureConfig != nil {
				r.AddARNRelation(resource.KmsKey, v.DataCaptureConfig.KmsKeyId)
			}
			for _, pv := range v.ProductionVariants {
				r.AddRelation(resource.SagemakerModel, pv.ModelName, "")
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
