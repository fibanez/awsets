package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSConfigDeliveryChannel struct {
}

func init() {
	i := AWSConfigDeliveryChannel{}
	listers = append(listers, i)
}

func (l AWSConfigDeliveryChannel) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.ConfigDeliveryChannel}
}

func (l AWSConfigDeliveryChannel) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := configservice.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()

	channels, err := svc.DescribeDeliveryChannels(ctx.Context, &configservice.DescribeDeliveryChannelsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list config delivery channels: %w", err)
	}
	for _, v := range channels.DeliveryChannels {
		r := resource.New(ctx, resource.ConfigDeliveryChannel, v.Name, v.Name, v)
		r.AddRelation(resource.S3Bucket, v.S3BucketName, "")
		rg.AddResource(r)
	}

	return rg, nil
}
