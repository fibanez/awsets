package lister

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/fibanez/awsets/arn"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

type AWSImageBuilderImageRecipe struct {
}

func init() {
	i := AWSImageBuilderImageRecipe{}
	listers = append(listers, i)
}

func (l AWSImageBuilderImageRecipe) Types() []resource.ResourceType {
	return []resource.ResourceType{
		resource.ImageBuilderImageRecipe,
	}
}

func (l AWSImageBuilderImageRecipe) List(ctx context.AWSetsCtx) (*resource.Group, error) {

	svc := imagebuilder.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListImageRecipes(ctx.Context, &imagebuilder.ListImageRecipesInput{
			MaxResults: 100,
			NextToken:  nt,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list imagebuilder image recipes: %w", err)
		}
		for _, recipeSummary := range res.ImageRecipeSummaryList {
			v, err := svc.GetImageRecipe(ctx.Context, &imagebuilder.GetImageRecipeInput{
				ImageRecipeArn: recipeSummary.Arn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get imagebuilder image recipe %s: %w", *recipeSummary.Name, err)
			}
			recipeArn := arn.ParseP(v.ImageRecipe.Arn)
			r := resource.New(ctx, resource.ImageBuilderImageRecipe, recipeArn.ResourceId, v.ImageRecipe.Name, v.ImageRecipe)
			if v.ImageRecipe.BlockDeviceMappings != nil {
				for _, bdm := range v.ImageRecipe.BlockDeviceMappings {
					if bdm.Ebs != nil {
						r.AddARNRelation(resource.KmsKey, bdm.Ebs.KmsKeyId)
						r.AddRelation(resource.Ec2Snapshot, bdm.Ebs.SnapshotId, "")
					}
				}
			}
			for _, c := range v.ImageRecipe.Components {
				r.AddARNRelation(resource.ImageBuilderComponent, c.ComponentArn)
			}
			rg.AddResource(r)
		}
		return res.NextToken, nil
	})
	return rg, err
}
