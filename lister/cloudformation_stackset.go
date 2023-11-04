package lister

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSCloudFormationStackSet struct {
}

func init() {
	i := AWSCloudFormationStackSet{}
	listers = append(listers, i)
}

func (l AWSCloudFormationStackSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.CloudFormationStackSet}
}

func (l AWSCloudFormationStackSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := cloudformation.NewFromConfig(ctx.AWSCfg)

	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListStackSets(ctx.Context, &cloudformation.ListStackSetsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "StackSets is not supported in this region") {
				// If StackSets are not supported in a region, returns validation exception
				return nil, nil
			}
			return nil, err
		}
		for _, summary := range res.Summaries {
			v, err := svc.DescribeStackSet(ctx.Context, &cloudformation.DescribeStackSetInput{
				StackSetName: summary.StackSetName,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to describe stack set %s: %w", *summary.StackSetName, err)
			}
			r := resource.New(ctx, resource.CloudFormationStackSet, v.StackSet.StackSetId, v.StackSet.StackSetName, v.StackSet)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
