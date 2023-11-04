package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSCloudwatchLogsGroups struct {
}

func init() {
	i := AWSCloudwatchLogsGroups{}
	listers = append(listers, i)
}

func (l AWSCloudwatchLogsGroups) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.LogGroup,
		resource.LogSubscriptionFilter,
		resource.LogMetricFilter,
		//resource.LogStream,
	}
}

func (l AWSCloudwatchLogsGroups) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudwatchlogs.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()

	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.DescribeLogGroups(ctx.Context, &cloudwatchlogs.DescribeLogGroupsInput{
			Limit:     aws.Int32(50),
			NextToken: nt,
		})
		if err != nil {
			return nil, err
		}
		for _, v := range res.LogGroups {
			//groupArn := arn.ParseP(v.Arn)
			//r := resource.New(ctx, resourcetype.LogGroup, groupArn.ResourceId, aws.StringValue(v.LogGroupName), v) // TODO switch back to this after fixing ARN parsing
			r := resource.New(ctx, resource.LogGroup, v.LogGroupName, v.LogGroupName, v)
			r.AddARNRelation(resource.KmsKey, v.KmsKeyId)
			rg.AddResource(r)

			// Subscription Filters
			err := Paginator(func(nt2 *string) (*string, error) {
				filters, err := svc.DescribeSubscriptionFilters(ctx.Context, &cloudwatchlogs.DescribeSubscriptionFiltersInput{
					Limit:        aws.Int32(50),
					LogGroupName: v.LogGroupName,
					NextToken:    nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get subscription filters for log group %s: %w", *v.LogGroupName, err)
				}
				for _, subFilter := range filters.SubscriptionFilters {
					subResource := resource.New(ctx, resource.LogSubscriptionFilter, subFilter.FilterName, subFilter.FilterName, subFilter)
					subResource.AddRelation(resource.LogGroup, v.LogGroupName, "")
					subResource.AddARNRelation(resource.IamRole, subFilter.RoleArn)
					rg.AddResource(subResource)
				}
				return filters.NextToken, nil
			})
			if err != nil {
				return nil, err
			}

			// Metric Filters
			err = Paginator(func(nt2 *string) (*string, error) {
				metrics, err := svc.DescribeMetricFilters(ctx.Context, &cloudwatchlogs.DescribeMetricFiltersInput{
					Limit:        aws.Int32(50),
					LogGroupName: v.LogGroupName,
					NextToken:    nt2,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to get metric filters for log group %s: %w", *v.LogGroupName, err)
				}
				for _, metricFilter := range metrics.MetricFilters {
					mfResource := resource.New(ctx, resource.LogMetricFilter, metricFilter.FilterName, metricFilter.FilterName, metricFilter)
					mfResource.AddRelation(resource.LogGroup, v.LogGroupName, "")
					rg.AddResource(mfResource)
				}
				return metrics.NextToken, nil
			})
			if err != nil {
				return nil, err
			}
		}

		return res.NextToken, nil
	})
	return rg, err
}
