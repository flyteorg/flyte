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

// RbacInterceptor is a gRPC interceptor that enforces RBAC or role based access control.
type RbacInterceptor struct {
	compiledPoliciesByRole       map[string]compiledAuthorizationPolicy
	compiledBypassMethodPatterns []*regexp.Regexp
	cfg                          config.Rbac
}

// compiledRule is a rule with compiled/validated regexes
type compiledRule struct {
	methodPattern *regexp.Regexp
	project       string
	domain        string
	name          string
}

func newCompiledRule(methodPattern *regexp.Regexp, project, domain, name string) compiledRule {
	return compiledRule{
		methodPattern: methodPattern,
		project:       project,
		domain:        domain,
		name:          name,
	}
}

// compiledAuthorizationPolicy is a policy with compiled/validated regexes
type compiledAuthorizationPolicy struct {
	role  string
	rules []compiledRule
}

func newAuthorizationPolicy(role string, rules []compiledRule) compiledAuthorizationPolicy {
	return compiledAuthorizationPolicy{
		role:  role,
		rules: rules,
	}
}

// NewRbacInterceptor initializes a new RbacInterceptor and does validation on the RBAC config
func NewRbacInterceptor(authCtx interfaces.AuthenticationContext) (*RbacInterceptor, error) {
	cfg := authCtx.Options().Rbac

	compiledPoliciesByRole := map[string]compiledAuthorizationPolicy{}

	// Validate policies
	for _, policy := range cfg.Policies {
		compiledPolicy, err := validatePolicy(policy)
		if err != nil {
			return nil, fmt.Errorf("validating authorization policy: %w", err)
		}

		_, exists := compiledPoliciesByRole[compiledPolicy.role]
		if exists {
			return nil, fmt.Errorf("found authorization policies with conflicting role %s", compiledPolicy.role)
		}

		compiledPoliciesByRole[compiledPolicy.role] = compiledPolicy
	}

	bypassMethodPatterns := []*regexp.Regexp{}

	for _, allowedMethod := range cfg.BypassMethodPatterns {
		// compile regexes and cache them
		compiled, err := regexp.Compile(allowedMethod)
		if err != nil {
			return nil, fmt.Errorf("compiling bypass method pattern %s: %w", allowedMethod, err)
		}

		bypassMethodPatterns = append(bypassMethodPatterns, compiled)
	}

	return &RbacInterceptor{
		compiledBypassMethodPatterns: bypassMethodPatterns,
		cfg:                          cfg,
		compiledPoliciesByRole:       compiledPoliciesByRole,
	}, nil
}

// UnaryInterceptor generates a grpc.UnaryServerInterceptor from RbacInterceptor
func (i *RbacInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		for _, allowedMethod := range i.compiledBypassMethodPatterns {
			if allowedMethod.MatchString(info.FullMethod) {
				logger.Debugf(ctx, "[%s] Authorization bypassed for method", info.FullMethod)
				return handler(ctx, req)
			}
		}

		identityContext := auth.IdentityContextFromContext(ctx)
		possibleRoles := i.resolvePossibleRoles(identityContext)

		if len(possibleRoles) == 0 {
			logger.Debugf(ctx, "[%s] No possible roles resolved. Unauthorized.", info.FullMethod)
			return nil, status.Errorf(codes.PermissionDenied, "")
		}

		logger.Debugf(ctx, "[%s] Found possible roles: %s", info.FullMethod, possibleRoles)

		// Try to match the possible roles against policies
		authorizedResourceScopes := i.calculateAuthorizedResourceScopes(ctx, possibleRoles, info)

		if len(authorizedResourceScopes) == 0 {
			logger.Debugf(ctx, "[%s] Found no matching authorization policy rules. Unauthorized.", info.FullMethod)
			return nil, status.Errorf(codes.PermissionDenied, "")
		}

		// Add authorized resource scopes to context so it can be used in the application later.
		isolationContext := isolation.NewIsolationContext(authorizedResourceScopes)
		ctx = isolationContext.WithContext(ctx)

		return handler(ctx, req)

	}
}

func (i *RbacInterceptor) resolvePossibleRoles(identityContext auth.IdentityContext) []string {

	roleSet := map[string]bool{}

	if i.cfg.TokenScopeRoleResolver.Enabled {

		for _, scopeRole := range identityContext.Scopes().List() {
			roleSet[scopeRole] = true
		}
	}

	if i.cfg.TokenClaimRoleResolver.Enabled {
		claimRoles := resolveRolesViaClaims(identityContext.Claims(), i.cfg.TokenClaimRoleResolver.TokenClaims)

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

		// Handle case of simple string value
		claimString, ok := claimIntf.(string)
		if ok {
			roleSet[claimString] = true
			continue
		}

		// Handle case of list of string values
		claimListElements, ok := claimIntf.([]interface{})
		if ok {
			for _, claimListElement := range claimListElements {
				claimStringElement, isString := claimListElement.(string)
				if isString {
					roleSet[claimStringElement] = true
				}
			}
		}
	}

	return maps.Keys(roleSet)
}

func (i *RbacInterceptor) calculateAuthorizedResourceScopes(ctx context.Context, roles []string, info *grpc.UnaryServerInfo) []isolation.ResourceScope {
	authorizedScopes := []isolation.ResourceScope{}

	matchingPolicies := map[string]compiledAuthorizationPolicy{}
	for _, role := range roles {
		policy, ok := i.compiledPoliciesByRole[role]
		if !ok {
			continue
		}

		matchingPolicies[role] = policy
	}

	logger.Debugf(ctx, "[%s] Found matching authorization policies: %+v", info.FullMethod, matchingPolicies)

	for role, policy := range matchingPolicies {
		matchingRules := discoverMatchingPolicyRulesForRequest(policy, info)

		if len(matchingRules) > 0 {
			logger.Debugf(ctx, "[%s] Found matching rules for role %s: %+v", info.FullMethod, role, matchingRules)
			// TODO: We should probably try and deduplicate any resource scopes but this is fine for now.
			for _, matchingRule := range matchingRules {
				authorizedScopes = append(authorizedScopes, isolation.ResourceScope{
					Project: matchingRule.project,
					Domain:  matchingRule.domain,
				})
			}
		} else {
			logger.Debugf(ctx, "[%s] Found no matching rules for role %s", info.FullMethod, role)
		}
	}

	return authorizedScopes
}

func discoverMatchingPolicyRulesForRequest(ap compiledAuthorizationPolicy, info *grpc.UnaryServerInfo) []compiledRule {
	matchingRules := []compiledRule{}
	for _, rule := range ap.rules {
		matches := rule.methodPattern.MatchString(info.FullMethod)

		if !matches {
			continue
		}

		matchingRules = append(matchingRules, rule)
	}

	return matchingRules
}

func validatePolicy(ap config.AuthorizationPolicy) (compiledAuthorizationPolicy, error) {
	compiledRules := []compiledRule{}

	for _, rule := range ap.Rules {
		if rule.Project == "" && rule.Domain != "" {
			return compiledAuthorizationPolicy{}, fmt.Errorf("authorization policy rule %s has invalid resource scope", rule.Name)
		}

		methodPattern, err := regexp.Compile(rule.MethodPattern)
		if err != nil {
			return compiledAuthorizationPolicy{}, fmt.Errorf("compiling rule %s pattern %s: %w", rule.Name, rule.MethodPattern, err)
		}

		compiledRules = append(compiledRules, newCompiledRule(methodPattern, rule.Project, rule.Domain, rule.Name))
	}

	return newAuthorizationPolicy(ap.Role, compiledRules), nil
}
