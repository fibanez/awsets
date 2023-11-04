package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSWafRegionalSizeConstraintSet struct {
}

func init() {
	i := AWSWafRegionalSizeConstraintSet{}
	listers = append(listers, i)
}

func (l AWSWafRegionalSizeConstraintSet) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.WafRegionalSizeConstraint}
}

func (l AWSWafRegionalSizeConstraintSet) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	svc := wafregional.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListSizeConstraintSets(ctx.Context, &wafregional.ListSizeConstraintSetsInput{
			Limit:      100,
			NextMarker: nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list regional size constraint sets: %w", err)
		}
		for _, id := range res.SizeConstraintSets {
			matchSet, err := svc.GetSizeConstraintSet(ctx.Context, &wafregional.GetSizeConstraintSetInput{
				SizeConstraintSetId: id.SizeConstraintSetId,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get size constraint set %s: %w", *id.SizeConstraintSetId, err)
			}
			if v := matchSet.SizeConstraintSet; v != nil {
				r := resource.New(ctx, resource.WafRegionalSizeConstraint, v.SizeConstraintSetId, v.Name, v)
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
