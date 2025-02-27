package interceptors

import (
	"context"
	"fmt"
	"regexp"

	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteadmin/auth"
	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/auth/isolation"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func GetAuthorizationInterceptor(authCtx interfaces.AuthenticationContext) (grpc.UnaryServerInterceptor, error) {

	noopFunc := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return nil, nil
	}

	opts := authCtx.Options().Rbac
	for _, policy := range opts.Policies {
		// FIXME: Move this to somewhere else?
		err := validatePolicy(policy)
		if err != nil {
			return noopFunc, fmt.Errorf("failed to validate authorization policy: %w", err)
		}
	}

	bypassMethodPatterns := []*regexp.Regexp{}

	for _, allowedMethod := range opts.BypassMethodPatterns {
		compiled, err := regexp.Compile(allowedMethod)
		if err != nil {
			return noopFunc, fmt.Errorf("compiling bypass method pattern %s: %w", allowedMethod, err)
		}

		bypassMethodPatterns = append(bypassMethodPatterns, compiled)
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		for _, allowedMethod := range bypassMethodPatterns {
			if allowedMethod.MatchString(info.FullMethod) {
				logger.Debugf(ctx, "[%s] Authorization bypassed for method", info.FullMethod)
				return handler(ctx, req)
			}
		}

		identityContext := auth.IdentityContextFromContext(ctx)
		roles := resolveRoles(opts, identityContext)

		if len(roles) == 0 {
			logger.Debugf(ctx, "[%s]No roles resolved. Unauthorized.", info.FullMethod)
			return nil, status.Errorf(codes.PermissionDenied, "")
		}

		logger.Debugf(ctx, "[%s]Found roles: %s", info.FullMethod, roles)

		authorizedResourceScopes, err := calculateAuthorizedResourceScopes(ctx, roles, opts.Policies, info)
		if err != nil {
			logger.Errorf(ctx, "[%s]Failed to calculate authorized scopes for user %s: %+v", info.FullMethod, identityContext.UserID(), err)
			return nil, status.Errorf(codes.Internal, "")
		}

		if len(authorizedResourceScopes) == 0 {
			logger.Debugf(ctx, "[%s]Found no matching authorization policy rules. Unauthorized.", info.FullMethod)
			return nil, status.Errorf(codes.PermissionDenied, "")
		}

		// Add authorized resource scopes to context
		isolationContext := isolation.NewIsolationContext(authorizedResourceScopes)

		isolationCtx := isolationContext.WithContext(ctx)
		return handler(isolationCtx, req)

	}, nil
}

func resolveRoles(rbac config.Rbac, identityContext auth.IdentityContext) []string {

	roleSet := map[string]bool{}

	if rbac.TokenScopeRoleResolver.Enabled {

		for _, scopeRole := range identityContext.Scopes().List() {
			roleSet[scopeRole] = true
		}
	}

	if rbac.TokenClaimRoleResolver.Enabled {
		claimRoles := resolveRolesViaClaims(identityContext.Claims(), rbac.TokenClaimRoleResolver.TokenClaims)

		for _, claimRole := range claimRoles {
			roleSet[claimRole] = true
		}
	}

	return maps.Keys(roleSet)
}

func resolveRolesViaClaims(claims map[string]interface{}, targetClaims []config.TokenClaim) []string {
	roleSet := map[string]bool{}

	for _, targetClaim := range targetClaims {
		claimIntf, ok := claims[targetClaim.Name]
		if !ok {
			continue
		}

		claimString, ok := claimIntf.(string)
		if ok {
			roleSet[claimString] = true
			continue
		}

		claimListElements, ok := claimIntf.([]interface{})
		if ok {
			for _, claimListElement := range claimListElements {
				claimStringElement, ok := claimListElement.(string)
				if ok {
					roleSet[claimStringElement] = true
				}
			}
		}
	}

	return maps.Keys(roleSet)
}

func calculateAuthorizedResourceScopes(ctx context.Context, roles []string, policies []config.AuthorizationPolicy, info *grpc.UnaryServerInfo) ([]isolation.ResourceScope, error) {
	authorizedScopes := []isolation.ResourceScope{}

	policiesByRole := map[string]config.AuthorizationPolicy{}
	for _, policy := range policies {
		policiesByRole[policy.Role] = policy
	}

	matchingPolicies := map[string]config.AuthorizationPolicy{}
	for _, role := range roles {
		policy, ok := policiesByRole[role]
		if !ok {
			continue
		}

		matchingPolicies[role] = policy
	}

	logger.Debugf(ctx, "[%s]Found matching authorization policies: %s", info.FullMethod, matchingPolicies)

	for role, policy := range matchingPolicies {
		matchingRules, err := authorizationPolicyMatchesRequest(policy, info)
		if err != nil {
			return authorizedScopes, fmt.Errorf("failed to match request: %w", err)
		}

		if len(matchingRules) > 0 {
			logger.Debugf(ctx, "[%s]Found matching rules for role %s: %s", info.FullMethod, role, matchingRules)
			for _, matchingRule := range matchingRules {
				authorizedScopes = append(authorizedScopes, isolation.ResourceScope{
					Project: matchingRule.Project,
					Domain:  matchingRule.Domain,
				})
			}
		} else {
			logger.Debugf(ctx, "[%s]Found no matching rules for role %s", info.FullMethod, role)
		}
	}

	return authorizedScopes, nil
}

func authorizationPolicyMatchesRequest(ap config.AuthorizationPolicy, info *grpc.UnaryServerInfo) ([]config.Rule, error) {
	matchingRules := []config.Rule{}
	for _, rule := range ap.Rules {
		matches, err := ruleMatchesRequest(rule, info)
		if err != nil {
			return []config.Rule{}, fmt.Errorf("matching rule against request: %w", err)
		}

		if !matches {
			continue
		}

		matchingRules = append(matchingRules, rule)
	}

	return matchingRules, nil
}

func ruleMatchesRequest(rule config.Rule, info *grpc.UnaryServerInfo) (bool, error) {
	pattern, err := regexp.Compile(rule.MethodPattern)
	if err != nil {
		return false, fmt.Errorf("compiling rule pattern %s: %w", rule.MethodPattern, err)
	}

	return pattern.MatchString(info.FullMethod), nil
}

func validatePolicy(ap config.AuthorizationPolicy) error {
	for _, rule := range ap.Rules {
		if rule.Project == "" && rule.Domain != "" {
			return fmt.Errorf("authorization policy rule %s has invalid resource scope", rule.Name)
		}
	}

	return nil
}
