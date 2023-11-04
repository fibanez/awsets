package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/fibanez/awsets/arn"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSSagemakerModel struct {
}

func init() {
	i := AWSSagemakerModel{}
	listers = append(listers, i)
}

func (l AWSSagemakerModel) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.SagemakerModel}
}

func (l AWSSagemakerModel) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := sagemaker.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListModels(ctx.Context, &sagemaker.ListModelsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, model := range res.Models {
			v, err := svc.DescribeModel(ctx.Context, &sagemaker.DescribeModelInput{
				ModelName: model.ModelName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe model %s: %w", *model.ModelName, err)
			}

			modelArn := arn.ParseP(v.ModelArn)
			r := resource.New(ctx, resource.SagemakerModel, modelArn.ResourceId, v.ModelName, v)
			r.AddARNRelation(resource.IamRole, v.ExecutionRoleArn)
			if vpc := v.VpcConfig; vpc != nil {
				for _, sg := range vpc.SecurityGroupIds {
					r.AddRelation(resource.Ec2SecurityGroup, sg, "")
				}
				for _, sn := range vpc.Subnets {
					r.AddRelation(resource.Ec2Subnet, sn, "")
				}
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
