package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dax"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSDAXSubnetGroup struct {
}

func init() {
	i := AWSDAXSubnetGroup{}
	listers = append(listers, i)
}

func (l AWSDAXSubnetGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DAXSubnetGroup}
}

func (l AWSDAXSubnetGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := dax.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeSubnetGroups(ctx.Context, &dax.DescribeSubnetGroupsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "Access Denied to API Version: DAX_V3") {
				// Regions that don't support DAX return access denied
				return nil, nil
			}
			return nil, fmt.Errorf("failed to list dax subnet groups: %w", err)
		}
		for _, v := range res.SubnetGroups {
			r := resource.New(ctx, resource.DAXSubnetGroup, v.SubnetGroupName, v.SubnetGroupName, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			for _, sn := range v.Subnets {
				r.AddRelation(resource.Ec2Subnet, sn.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
