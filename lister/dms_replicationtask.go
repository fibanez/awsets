package lister

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSDMSReplicationTask struct {
}

func init() {
	i := AWSDMSReplicationTask{}
	listers = append(listers, i)
}

func (l AWSDMSReplicationTask) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.DMSReplicationTask}
}

func (l AWSDMSReplicationTask) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := databasemigrationservice.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeReplicationTasks(ctx.Context, &databasemigrationservice.DescribeReplicationTasksInput{
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
		for _, v := range res.ReplicationTasks {
			r := resource.New(ctx, resource.DMSReplicationTask, v.ReplicationTaskIdentifier, v.ReplicationTaskIdentifier, v)
			r.AddARNRelation(resource.DMSReplicationInstance, v.ReplicationInstanceArn)
			rg.AddResource(r)
		}
		return res.Marker, nil
	})
	return rg, err
}
