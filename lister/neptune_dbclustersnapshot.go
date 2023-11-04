package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSNeptuneDbClusterSnapshot struct {
}

func init() {
	i := AWSNeptuneDbClusterSnapshot{}
	listers = append(listers, i)
}

func (l AWSNeptuneDbClusterSnapshot) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.NeptuneDbClusterSnapshot}
}

func (l AWSNeptuneDbClusterSnapshot) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := neptune.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	paginator := neptune.NewDescribeDBClusterSnapshotsPaginator(svc, &neptune.DescribeDBClusterSnapshotsInput{
		MaxRecords: aws.Int32(100),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to list neptune cluster snapshots: %w", err)
		}
		for _, v := range page.DBClusterSnapshots {
			if !strings.Contains(*v.Engine, "neptune") {
				continue
			}
			r := resource.New(ctx, resource.NeptuneDbClusterSnapshot, v.DBClusterSnapshotIdentifier, v.DBClusterSnapshotIdentifier, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			r.AddRelation(resource.NeptuneDbCluster, v.DBClusterIdentifier, "")
			r.AddRelation(resource.Ec2Vpc, v.VpcId, "")

			rg.AddResource(r)
		}
	}
	return rg, nil
}
