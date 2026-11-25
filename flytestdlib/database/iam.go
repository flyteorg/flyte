package database

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	rdsauth "github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/jackc/pgx/v5"
)

// iamTokenBuilder returns an AWS RDS IAM auth token for user at endpoint ("host:port").
// A token is valid for 15 minutes but only for authentication; open connections outlive it.
type iamTokenBuilder func(ctx context.Context, endpoint, user string) (string, error)

// newIAMTokenBuilder loads the AWS SDK default credential chain once, so every token reuses its cached
// credentials. An empty region falls back to the region the SDK resolves (AWS_REGION, profile, ...).
// Replaced in tests.
var newIAMTokenBuilder = func(ctx context.Context, region string) (iamTokenBuilder, error) {
	var opts []func(*awsConfig.LoadOptions) error
	if region != "" {
		opts = append(opts, awsConfig.WithRegion(region))
	}
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config for RDS IAM auth: %w", err)
	}
	if awsCfg.Region == "" {
		return nil, fmt.Errorf("RDS IAM auth needs an AWS region: set database.postgres.region or AWS_REGION")
	}
	return func(ctx context.Context, endpoint, user string) (string, error) {
		return rdsauth.BuildAuthToken(ctx, endpoint, awsCfg.Region, user, awsCfg.Credentials)
	}, nil
}

// iamBeforeConnect sets a fresh IAM auth token as the password of each new pgx connection.
func iamBeforeConnect(buildToken iamTokenBuilder) func(context.Context, *pgx.ConnConfig) error {
	return func(ctx context.Context, cc *pgx.ConnConfig) error {
		endpoint := net.JoinHostPort(cc.Host, strconv.Itoa(int(cc.Port)))
		token, err := buildToken(ctx, endpoint, cc.User)
		if err != nil {
			return fmt.Errorf("build RDS IAM auth token for %s: %w", endpoint, err)
		}
		cc.Password = token
		return nil
	}
}

// iamPassfile holds a single wildcard pgpass entry with the current IAM auth token.
type iamPassfile struct {
	path       string
	endpoint   string
	user       string
	buildToken iamTokenBuilder
}

func (p *iamPassfile) refresh(ctx context.Context) error {
	token, err := p.buildToken(ctx, p.endpoint, p.user)
	if err != nil {
		return fmt.Errorf("build RDS IAM auth token for %s: %w", p.endpoint, err)
	}
	// Write and rename so a concurrent dial never reads a partial file. CreateTemp creates the file
	// with mode 0600, which lib/pq requires before it will read a pgpass file.
	tmp, err := os.CreateTemp(filepath.Dir(p.path), ".pgpass-")
	if err != nil {
		return fmt.Errorf("write pgpass file: %w", err)
	}
	defer os.Remove(tmp.Name()) //nolint:errcheck // already renamed on success
	if _, err := tmp.WriteString("*:*:*:*:" + escapePgpassField(token) + "\n"); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write pgpass file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write pgpass file: %w", err)
	}
	if err := os.Rename(tmp.Name(), p.path); err != nil {
		return fmt.Errorf("write pgpass file: %w", err)
	}
	return nil
}

// escapePgpassField escapes the separators of a pgpass field. RDS tokens contain ':'.
func escapePgpassField(s string) string {
	return strings.NewReplacer(`\`, `\\`, `:`, `\:`).Replace(s)
}

// quoteDsnValue single-quotes a key/value DSN value.
func quoteDsnValue(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s) + "'"
}
