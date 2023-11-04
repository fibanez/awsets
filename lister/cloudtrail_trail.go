package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/fibanez/awsets/arn"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSCloudtrailTrail struct {
}

func init() {
	i := AWSCloudtrailTrail{}
	listers = append(listers, i)
}

func (l AWSCloudtrailTrail) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudtrailTrail}
}

func (l AWSCloudtrailTrail) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudtrail.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	trails, err := svc.DescribeTrails(ctx.Context, &cloudtrail.DescribeTrailsInput{
		IncludeShadowTrails: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list cloudtrail trails: %w", err)
	}
	for _, trail := range trails.TrailList {
		r := resource.New(ctx, resource.CloudtrailTrail, trail.Name, trail.Name, trail)
		r.AddARNRelation(resource.KmsKey, trail.KmsKeyId)
		if trail.S3BucketName != nil {
			r.AddCrossRelation(ctx.AccountId, trail.HomeRegion, resource.S3Bucket, trail.S3BucketName, "")
		}
		if trail.SnsTopicARN != nil {
			snsArn := arn.ParseP(trail.SnsTopicARN)
			r.AddRelation(resource.SnsTopic, snsArn.ResourceId, snsArn.ResourceVersion)
		}
		//if trail.CloudWatchLogsLogGroupArn != nil { //TODO re-enable after arn parsing is fixed for log groups
		//	cwLogsArn := arn.ParseP(trail.CloudWatchLogsLogGroupArn)
		//	r.AddRelation(resource.LogGroup, cwLogsArn.ResourceId, "")
		//}
		if trail.CloudWatchLogsRoleArn != nil {
			cwLogsRoleArn := arn.ParseP(trail.CloudWatchLogsRoleArn)
			r.AddRelation(resource.LogGroup, cwLogsRoleArn.ResourceId, "")
		}

		trailArn := arn.ParseP(trail.TrailARN)
		if trailArn.Account == ctx.AccountId {
			if trail.HomeRegion != nil && *trail.HomeRegion == ctx.Region() {
				statusRes, err := svc.GetTrailStatus(ctx.Context, &cloudtrail.GetTrailStatusInput{
					Name: trail.Name,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get status for trail %s: %w", *trail.Name, err)
				}
				r.AddAttribute("Status", statusRes)
			}
		}
		rg.AddResource(r)
	}
	return rg, nil
}
