// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package authz

import (
	"context"
	"errors"
	"fmt"

	"github.com/facebookincubator/symphony/pkg/ent/predicate"
	"github.com/facebookincubator/symphony/pkg/ent/user"
	"github.com/facebookincubator/symphony/pkg/viewer"

	models2 "github.com/facebookincubator/symphony/pkg/authz/models"

	"github.com/facebookincubator/symphony/graph/graphql/models"
	"github.com/facebookincubator/symphony/pkg/ent"
	"github.com/facebookincubator/symphony/pkg/ent/privacy"
	"github.com/facebookincubator/symphony/pkg/ent/project"
	"github.com/facebookincubator/symphony/pkg/ent/projecttype"
)

func projectCudBasedCheck(ctx context.Context, cud *models.WorkforceCud, m *ent.ProjectMutation) (bool, error) {
	if m.Op().Is(ent.OpCreate) {
		typeID, exists := m.TypeID()
		if !exists {
			return false, errors.New("creating project with no type")
		}
		return checkWorkforce(cud.Create, nil, &typeID), nil
	}
	id, exists := m.ID()
	if !exists {
		return false, nil
	}
	projectTypeID, err := m.Client().ProjectType.Query().
		Where(projecttype.HasProjectsWith(project.ID(id))).
		OnlyID(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to fetch project type id: %w", err)
	}
	if m.Op().Is(ent.OpUpdateOne) {
		return checkWorkforce(cud.Update, nil, &projectTypeID), nil
	}
	return checkWorkforce(cud.Delete, nil, &projectTypeID), nil
}

func projectReadPredicate(ctx context.Context) predicate.Project {
	var predicates []predicate.Project
	rule := FromContext(ctx).WorkforcePolicy.Read
	switch rule.IsAllowed {
	case models2.PermissionValueYes:
		return nil
	case models2.PermissionValueByCondition:
		predicates = append(predicates,
			project.HasTypeWith(projecttype.IDIn(rule.ProjectTypeIds...)))
	}
	if v, exists := viewer.FromContext(ctx).(*viewer.UserViewer); exists {
		predicates = append(predicates,
			project.HasCreatorWith(user.ID(v.User().ID)),
		)
	}
	if woPredicate := workOrderReadPredicate(ctx); woPredicate != nil {
		predicates = append(predicates,
			project.HasWorkOrdersWith(woPredicate))
	}
	return project.Or(predicates...)
}

func isCreatorChanged(ctx context.Context, m *ent.ProjectMutation) (bool, error) {
	var currCreatorID *int
	creatorIDToSet, created := m.CreatorID()
	creatorCleared := m.CreatorCleared()
	if !created && !creatorCleared {
		return false, nil
	}
	projectID, exists := m.ID()
	if !exists {
		return created, nil
	}
	creatorID, err := m.Client().User.Query().
		Where(user.HasCreatedProjectsWith(project.ID(projectID))).
		OnlyID(ctx)
	if err == nil {
		currCreatorID = &creatorID
	}
	if err != nil && !ent.IsNotFound(err) {
		return false, privacy.Denyf("failed to fetch creator: %w", err)
	}
	switch {
	case currCreatorID == nil && created:
		return true, nil
	case currCreatorID != nil && created && *currCreatorID != creatorIDToSet:
		return true, nil
	case currCreatorID != nil && creatorCleared:
		return true, nil
	}
	return false, nil
}

// ProjectWritePolicyRule grants write permission to project based on policy.
func ProjectWritePolicyRule() privacy.MutationRule {
	return privacy.ProjectMutationRuleFunc(func(ctx context.Context, m *ent.ProjectMutation) error {
		cud := FromContext(ctx).WorkforcePolicy.Data
		allowed, err := projectCudBasedCheck(ctx, cud, m)
		if err != nil {
			return privacy.Denyf(err.Error())
		}
		creatorChanged, err := isCreatorChanged(ctx, m)
		if err != nil {
			return privacy.Denyf(err.Error())
		}
		if creatorChanged {
			v, isUser := viewer.FromContext(ctx).(*viewer.UserViewer)
			creatorID, exists := m.CreatorID()
			if !m.Op().Is(ent.OpCreate) || !isUser || !exists || v.User().ID != creatorID {
				allowed = allowed && (cud.TransferOwnership.IsAllowed == models2.PermissionValueYes)
			}
		}
		if allowed {
			return privacy.Allow
		}
		return privacy.Skip
	})
}

// ProjectReadPolicyRule grants read permission to project based on policy.
func ProjectReadPolicyRule() privacy.QueryRule {
	return privacy.ProjectQueryRuleFunc(func(ctx context.Context, q *ent.ProjectQuery) error {
		projectPredicate := projectReadPredicate(ctx)
		if projectPredicate != nil {
			q.Where(projectPredicate)
		}
		return privacy.Skip
	})
}

// ProjectTypeWritePolicyRule grants write permission to project type based on policy.
func ProjectTypeWritePolicyRule() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		return cudBasedRule(FromContext(ctx).WorkforcePolicy.Templates, m)
	})
}

// WorkOrderDefinitionWritePolicyRule grants write permission to work order definition based on policy.
func WorkOrderDefinitionWritePolicyRule() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		return allowOrSkip(FromContext(ctx).WorkforcePolicy.Templates.Update)
	})
}
