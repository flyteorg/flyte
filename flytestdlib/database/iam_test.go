package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTokens stands in for the AWS token builder and records every request.
type fakeTokens struct {
	mu        sync.Mutex
	token     string
	err       error
	endpoints []string
	users     []string
}

func (f *fakeTokens) set(token string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.token = token
}

func (f *fakeTokens) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.endpoints)
}

func (f *fakeTokens) build(_ context.Context, endpoint, user string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.endpoints = append(f.endpoints, endpoint)
	f.users = append(f.users, user)
	return f.token, f.err
}

// stubIAMTokenBuilder replaces the AWS-backed token builder for the duration of the test.
func stubIAMTokenBuilder(t *testing.T, fake *fakeTokens) {
	t.Helper()
	orig := newIAMTokenBuilder
	newIAMTokenBuilder = func(context.Context) (iamTokenBuilder, error) {
		return fake.build, nil
	}
	t.Cleanup(func() { newIAMTokenBuilder = orig })
}

func TestIAMBeforeConnect(t *testing.T) {
	t.Run("sets a token for the connection's endpoint and user", func(t *testing.T) {
		fake := &fakeTokens{token: "tok"}
		cc, err := pgx.ParseConfig("host=replica.example.com port=6543 user=flyte dbname=flyte")
		require.NoError(t, err)

		require.NoError(t, iamBeforeConnect(fake.build)(context.Background(), cc))
		assert.Equal(t, "tok", cc.Password)
		assert.Equal(t, []string{"replica.example.com:6543"}, fake.endpoints)
		assert.Equal(t, []string{"flyte"}, fake.users)
	})

	t.Run("returns the builder's error", func(t *testing.T) {
		fake := &fakeTokens{err: errors.New("no credentials")}
		cc, err := pgx.ParseConfig("host=db.example.com port=5432 user=flyte dbname=flyte")
		require.NoError(t, err)

		err = iamBeforeConnect(fake.build)(context.Background(), cc)
		assert.ErrorContains(t, err, "build RDS IAM auth token for db.example.com:5432: no credentials")
		assert.Empty(t, cc.Password)
	})
}

func TestConnectorOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("password auth adds no hook", func(t *testing.T) {
		opts, err := connectorOptions(ctx, PostgresConfig{Password: "secret"})
		assert.NoError(t, err)
		assert.Empty(t, opts)
	})

	t.Run("invalid config is rejected before any connection", func(t *testing.T) {
		_, err := connectorOptions(ctx, PostgresConfig{AuthType: "iamm"})
		assert.ErrorContains(t, err, "unknown database postgres authType")
	})

	t.Run("iam auth adds the token hook", func(t *testing.T) {
		stubIAMTokenBuilder(t, &fakeTokens{})
		opts, err := connectorOptions(ctx, PostgresConfig{AuthType: AuthTypeIAM, ExtraOptions: "sslmode=require"})
		assert.NoError(t, err)
		assert.Len(t, opts, 1)
	})

	t.Run("iam auth surfaces builder setup errors", func(t *testing.T) {
		orig := newIAMTokenBuilder
		t.Cleanup(func() { newIAMTokenBuilder = orig })
		newIAMTokenBuilder = func(context.Context) (iamTokenBuilder, error) {
			return nil, errors.New("no region")
		}
		_, err := connectorOptions(ctx, PostgresConfig{AuthType: AuthTypeIAM, ExtraOptions: "sslmode=require"})
		assert.ErrorContains(t, err, "no region")
	})
}

func TestIAMPassfileRefresh(t *testing.T) {
	fake := &fakeTokens{token: `host:5432/?Action=connect&X-Amz-Signature=a\b`}
	p := &iamPassfile{
		path:       filepath.Join(t.TempDir(), "pgpass"),
		endpoint:   "db.example.com:5432",
		user:       "flyte",
		buildToken: fake.build,
	}
	require.NoError(t, p.refresh(context.Background()))

	content, err := os.ReadFile(p.path)
	require.NoError(t, err)
	assert.Equal(t, `*:*:*:*:host\:5432/?Action=connect&X-Amz-Signature=a\\b`+"\n", string(content))
	info, err := os.Stat(p.path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(), "lib/pq ignores pgpass files with group or world access")
	assert.Equal(t, []string{"db.example.com:5432"}, fake.endpoints)

	fake.err = errors.New("expired credentials")
	assert.ErrorContains(t, p.refresh(context.Background()), "expired credentials")
	unchanged, err := os.ReadFile(p.path)
	require.NoError(t, err)
	assert.Equal(t, content, unchanged, "a failed refresh must leave the last token in place")
}

func TestQuoteDsnValue(t *testing.T) {
	assert.Equal(t, `'/tmp/flyte-pgpass-1/pgpass'`, quoteDsnValue("/tmp/flyte-pgpass-1/pgpass"))
	assert.Equal(t, `'/tmp/it\'s a\\dir'`, quoteDsnValue(`/tmp/it's a\dir`))
}

const iamTestPGPort = 15439

// TestIAMAuthEmbeddedPostgres checks against a real server that every IAM connection path authenticates with
// the token, and that a new connection fetches a new token. The fake token is the role's password, which
// contains the ':' and '\' that pgpass escaping must handle.
func TestIAMAuthEmbeddedPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip("starts an embedded postgres")
	}
	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(iamTestPGPort).
		Database("postgres").
		Username("postgres").
		Password("postgres").
		RuntimePath(fmt.Sprintf("/tmp/embedded-postgres-%d", iamTestPGPort)))
	require.NoError(t, pg.Start())
	t.Cleanup(func() { _ = pg.Stop() })

	admin, err := sqlx.Open("pgx", fmt.Sprintf(
		"host=localhost port=%d user=postgres password=postgres dbname=postgres sslmode=disable", iamTestPGPort))
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	const firstToken = `tok:en\one`
	_, err = admin.Exec(fmt.Sprintf(`CREATE ROLE iam_user LOGIN CREATEDB PASSWORD %s`, pq.QuoteLiteral(firstToken)))
	require.NoError(t, err)

	iamConfig := PostgresConfig{
		Host:     "localhost",
		Port:     iamTestPGPort,
		DbName:   "iam_bootstrap",
		User:     "iam_user",
		AuthType: AuthTypeIAM,
		// The embedded server has no TLS; prefer still exercises the IAM path, which rejects sslmode=disable.
		ExtraOptions: "sslmode=prefer",
	}

	t.Run("pool creates the database and fetches a token per connection", func(t *testing.T) {
		fake := &fakeTokens{token: firstToken}
		stubIAMTokenBuilder(t, fake)

		db, err := CreatePostgresDbIfNotExists(context.Background(), iamConfig)
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		// Missing-database attempt, bootstrap connection to "postgres", then the new database.
		assert.GreaterOrEqual(t, fake.calls(), 3)
		for _, endpoint := range fake.endpoints {
			assert.Equal(t, fmt.Sprintf("localhost:%d", iamTestPGPort), endpoint)
		}

		db.SetMaxIdleConns(0) // every query opens a new physical connection
		before := fake.calls()
		for i := 0; i < 2; i++ {
			var name string
			require.NoError(t, db.Get(&name, "SELECT current_database()"))
			assert.Equal(t, "iam_bootstrap", name)
		}
		assert.Equal(t, before+2, fake.calls())
	})

	t.Run("pool fails when the token is rejected", func(t *testing.T) {
		stubIAMTokenBuilder(t, &fakeTokens{token: "wrong"})
		_, err := CreatePostgresDbIfNotExists(context.Background(), iamConfig)
		assert.ErrorContains(t, err, "password authentication failed")
	})

	t.Run("listener reconnects with a refreshed token", func(t *testing.T) {
		fake := &fakeTokens{token: firstToken}
		stubIAMTokenBuilder(t, fake)
		listenerConfig := iamConfig
		listenerConfig.DbName = "postgres" // NOTIFY is per database; admin notifies on postgres
		listenerConfig.ExtraOptions += " application_name=iam_listener_test"

		reconnected := make(chan struct{}, 1)
		listener, err := NewPostgresListener(context.Background(), listenerConfig, 10*time.Millisecond, 100*time.Millisecond,
			func(ev pq.ListenerEventType, _ error) {
				if ev == pq.ListenerEventReconnected {
					select {
					case reconnected <- struct{}{}:
					default:
					}
				}
			})
		require.NoError(t, err)
		t.Cleanup(func() { _ = listener.Close() })
		require.NoError(t, listener.Listen("iam_test"))

		// Rotate the password so the reconnect only succeeds with a token written after the disconnect.
		const secondToken = `tok:en\two`
		fake.set(secondToken)
		_, err = admin.Exec(fmt.Sprintf(`ALTER ROLE iam_user PASSWORD %s`, pq.QuoteLiteral(secondToken)))
		require.NoError(t, err)
		_, err = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE application_name = 'iam_listener_test'`)
		require.NoError(t, err)

		select {
		case <-reconnected:
		case <-time.After(10 * time.Second):
			t.Fatal("listener did not reconnect with the refreshed token")
		}
		// pq sends nil on Notify after a reconnect; drain it before checking delivery.
		require.Eventually(t, func() bool {
			_, err := admin.Exec(`SELECT pg_notify('iam_test', 'hello')`)
			require.NoError(t, err)
			select {
			case n := <-listener.Notify:
				return n != nil && n.Extra == "hello"
			case <-time.After(200 * time.Millisecond):
				return false
			}
		}, 10*time.Second, 50*time.Millisecond)
	})
}

// The real builder signs locally from the SDK credential chain and region; static env credentials keep it offline.
func TestNewIAMTokenBuilder(t *testing.T) {
	isolateAWSEnv := func(t *testing.T) {
		t.Helper()
		missing := filepath.Join(t.TempDir(), "missing")
		t.Setenv("AWS_CONFIG_FILE", missing)
		t.Setenv("AWS_SHARED_CREDENTIALS_FILE", missing)
		t.Setenv("AWS_PROFILE", "")
		t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
		t.Setenv("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
		t.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
		t.Setenv("AWS_REGION", "")
		t.Setenv("AWS_DEFAULT_REGION", "")
	}
	buildToken := func(t *testing.T) string {
		t.Helper()
		build, err := newIAMTokenBuilder(context.Background())
		require.NoError(t, err)
		token, err := build(context.Background(), "db.example.com:5432", "flyte")
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(token, "db.example.com:5432?Action=connect&"), token)
		assert.Contains(t, token, "DBUser=flyte")
		return token
	}

	t.Run("region from AWS_REGION", func(t *testing.T) {
		isolateAWSEnv(t)
		t.Setenv("AWS_REGION", "ap-southeast-2")
		assert.Contains(t, buildToken(t), "%2Fap-southeast-2%2Frds-db%2Faws4_request")
	})

	t.Run("region from AWS_DEFAULT_REGION", func(t *testing.T) {
		isolateAWSEnv(t)
		t.Setenv("AWS_DEFAULT_REGION", "us-west-2")
		assert.Contains(t, buildToken(t), "%2Fus-west-2%2Frds-db%2Faws4_request")
	})

	t.Run("no region anywhere is an error", func(t *testing.T) {
		isolateAWSEnv(t)
		_, err := newIAMTokenBuilder(context.Background())
		assert.ErrorContains(t, err, "RDS IAM auth needs an AWS region")
	})
}
