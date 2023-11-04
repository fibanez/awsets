package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSDMSReplicationSubnetGroup struct {
}

func init() {
	i := AWSDMSReplicationSubnetGroup{}
	listers = append(listers, i)
}

func (l AWSDMSReplicationSubnetGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DMSReplicationSubnetGroup}
}

func (l AWSDMSReplicationSubnetGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := databasemigrationservice.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeReplicationSubnetGroups(ctx.Context, &databasemigrationservice.DescribeReplicationSubnetGroupsInput{
			MaxRecords: aws.Int32(100),
			Marker:     nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "exceeded maximum number of attempts") {
				// If DMS is not supported in a region, it triggers this error
				return nil, nil
			}
			return nil, err
		}
		for _, v := range res.ReplicationSubnetGroups {
			r := resource.New(ctx, resource.DMSReplicationSubnetGroup, v.ReplicationSubnetGroupIdentifier, v.ReplicationSubnetGroupIdentifier, v)
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")
			for _, sn := range v.Subnets {
				r.AddRelation(resource.Ec2Subnet, sn.SubnetIdentifier, "")
			}
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
