package lister

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/fibanez/awsets/context"
	"github.com/fibanez/awsets/resource"
)

var listWafv2RuleGroupOnce sync.Once

type AWSWafv2RuleGroup struct {
}

func init() {
	i := AWSWafv2RuleGroup{}
	listers = append(listers, i)
}

func (l AWSWafv2RuleGroup) Types() []resource.ResourceType {
	return []resource.ResourceType{resource.Wafv2RuleGroup}
}

func (l AWSWafv2RuleGroup) List(ctx context.AWSetsCtx) (*resource.Group, error) {
	rg := resource.NewGroup()

	rg, err := wafv2RuleGroupQuery(ctx, types.ScopeRegional)
	if err != nil {
		return nil, fmt.Errorf("failed to list rule groups: %w", err)
	}

	// Do global
	var outerErr error
	listWafv2RuleGroupOnce.Do(func() {
		ctxUsEast := ctx.Copy("us-east-1")
		rgNew, err := wafv2RuleGroupQuery(*ctxUsEast, types.ScopeCloudfront)
		if err != nil {
			outerErr = fmt.Errorf("failed to list global rule groups: %w", err)
		}
		rg.Merge(rgNew)
	})

	return rg, outerErr
}

func wafv2RuleGroupQuery(ctx context.AWSetsCtx, scope types.Scope) (*resource.Group, error) {
	svc := wafv2.NewFromConfig(ctx.AWSCfg)
	rg := resource.NewGroup()
	err := Paginator(func(nt *string) (*string, error) {
		res, err := svc.ListRuleGroups(ctx.Context, &wafv2.ListRuleGroupsInput{
			Limit:      aws.Int32(100),
			NextMarker: nt,
			Scope:      scope,
		})
		if err != nil {
			return nil, err
		}
		for _, rgId := range res.RuleGroups {
			rulegroup, err := svc.GetRuleGroup(ctx.Context, &wafv2.GetRuleGroupInput{
				Id:    rgId.Id,
				Name:  rgId.Name,
				Scope: scope,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get rule group %s for scope %v: %w", *rgId.Id, scope, err)
			}
			if v := rulegroup.RuleGroup; v != nil {
				var r resource.Resource
				if scope == types.ScopeCloudfront {
					r = resource.NewGlobal(ctx, resource.Wafv2RuleGroup, v.Id, v.Name, v)
				} else {
					r = resource.New(ctx, resource.Wafv2RuleGroup, v.Id, v.Name, v)
				}
				rg.AddResource(r)
			}
		}
		return res.NextMarker, nil
	})
	return rg, err
}
