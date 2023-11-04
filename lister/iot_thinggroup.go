package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSIoTThingGroup struct {
}

func init() {
	i := AWSIoTThingGroup{}
	listers = append(listers, i)
}

func (l AWSIoTThingGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.IoTThingGroup}
}

func (l AWSIoTThingGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := iot.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListThingGroups(ctx.Context, &iot.ListThingGroupsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list iot thing groups: %w", err)
		}
		for _, group := range res.ThingGroups {
			r := resource.New(ctx, resource.IoTThingGroup, group.GroupName, group.GroupName, group)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
