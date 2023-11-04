package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/fibanez/awsets/arn"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSImageBuilderComponent struct {
}

func init() {
	i := AWSImageBuilderComponent{}
	listers = append(listers, i)
}

func (l AWSImageBuilderComponent) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ImageBuilderComponent,
	}
}

func (l AWSImageBuilderComponent) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := imagebuilder.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListComponents(ctx.Context, &imagebuilder.ListComponentsInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list imagebuilder components: %w", err)
		}
		for _, cv := range res.ComponentVersionList {
			v, err := svc.GetComponent(ctx.Context, &imagebuilder.GetComponentInput{
				ComponentBuildVersionArn: cv.Arn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get imagebuilder component %s: %w", *cv.Name, err)
			}
			cArn := arn.ParseP(v.Component.Arn)
			r := resource.NewVersion(ctx, resource.ImageBuilderComponent, cArn.ResourceId, cArn.ResourceVersion, v.Component.Name, v.Component)
			r.AddARNRelation(resource.KmsKey, v.Component.KmsKeyId)
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
