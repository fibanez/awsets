package lister

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSEc2VpcPeering struct {
}

func init() {
	i := AWSEc2VpcPeering{}
	listers = append(listers, i)
}

func (l AWSEc2VpcPeering) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Ec2VpcPeering}
}

func (l AWSEc2VpcPeering) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := ec2.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeVpcPeeringConnections(ctx.Context, &ec2.DescribeVpcPeeringConnectionsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.VpcPeeringConnections {
			r := resource.New(ctx, resource.Ec2VpcPeering, v.VpcPeeringConnectionId, v.VpcPeeringConnectionId, v)
			if v.AccepterVpcInfo != nil {
				r.AddCrossRelation(ctx.AccountId, v.AccepterVpcInfo.Region, resource.Ec2Vpc, v.AccepterVpcInfo.VpcId, "")
			}
			if v.RequesterVpcInfo != nil {
				r.AddCrossRelation(ctx.AccountId, v.RequesterVpcInfo.Region, resource.Ec2Vpc, v.RequesterVpcInfo.VpcId, "")
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
