package impl

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/cacheservice/pkg/errors"
	"github.com/flyteorg/flyte/cacheservice/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/cacheservice/pkg/runtime/configs"
	"github.com/flyteorg/flyte/cacheservice/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

var (
	_ interfaces.CacheDataStoreClient = &memClient{}
)

type memClient struct {
	cacheMap *map[string]*models.CachedOutput
}

// Get returns the cached output for the given key. It returns an error if the key does not exist.
func (c *memClient) Get(ctx context.Context, key string) (*models.CachedOutput, error) {
	cache := *c.cacheMap
	if value, exists := cache[key]; exists {
		return value, nil
	}

	return nil, errors.NewNotFoundError("output", key)
}

// Put will always set the value for the given key. It will overwrite the existing value if it exists.
func (c *memClient) Put(ctx context.Context, key string, cachedOutput *models.CachedOutput) error {
	cache := *c.cacheMap
	cache[key] = cachedOutput
	return nil
}

// Delete is an idempotent operation. It will not return an error if the key does not exist.
func (c *memClient) Delete(ctx context.Context, key string) error {
	cache := *c.cacheMap
	if _, exists := cache[key]; exists {
		delete(cache, key)
		return nil
	}

	return errors.NewNotFoundError("output", key)
}

var (
	_ interfaces.CacheDataStoreClient = &dynamoClient{}
)

type dynamoClient struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

// Get returns the cached output for the given key. It returns an error if the key does not exist.
func (c *dynamoClient) Get(ctx context.Context, key string) (*models.CachedOutput, error) {
	dynamoKey, err := attributevalue.Marshal(key)
	if err != nil {
		logger.Debugf(ctx, "Failed to marshal key to dynamo key with error %v", err)
		return nil, err
	}

	response, err := c.DynamoDbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{"id": dynamoKey}, TableName: aws.String(c.TableName),
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to get item from dynamo with error %v", err)
		return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to get item from dynamo with error %v", err)
	}
	if response.Item == nil {
		logger.Debugf(ctx, "Item with key %s not found in dynamo", key)
		return nil, errors.NewNotFoundError("output", key)
	}

	cachedOutput := models.CachedOutput{}
	err = attributevalue.UnmarshalMap(response.Item, &cachedOutput)
	if err != nil {
		logger.Debugf(ctx, "Failed to unmarshal dynamo item with error %v", err)
		return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to unmarshal dynamo item with error %v", err)
	}

	return &cachedOutput, nil
}

// Put will always set the value for the given key. It will overwrite the existing value if it exists.
func (c *dynamoClient) Put(ctx context.Context, key string, cachedOutput *models.CachedOutput) error {
	item, err := attributevalue.MarshalMap(cachedOutput)
	if err != nil {
		logger.Debugf(ctx, "Failed to marshal output to dynamo object with error %v", err)
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed to marshal output to dynamo object with error %v", err)
	}

	_, err = c.DynamoDbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(c.TableName), Item: item,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed updating dynamo table with error %v", err)
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed updating dynamo table with error %v", err)
	}
	return nil
}

// Delete is an idempotent operation. It will not return an error if the key does not exist.
func (c *dynamoClient) Delete(ctx context.Context, key string) error {
	dynamoKey, err := attributevalue.Marshal(key)
	if err != nil {
		logger.Debugf(ctx, "Failed to marshal key to dynamo key with error %v", err)
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed to marshal key to dynamo key with error %v", err)
	}
	_, err = c.DynamoDbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(c.TableName), Key: map[string]types.AttributeValue{"id": dynamoKey},
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to delete item with key %v from dynamo with error %v", key, err)
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed to delete item with key %v from dynamo with error %v", key, err)
	}
	return nil
}

var (
	_ interfaces.CacheDataStoreClient = &redisClient{}
)

type redisClient struct {
	RedisClient *redis.Client
}

// Get returns the cached output for the given key. It returns an error if the key does not exist.
func (r *redisClient) Get(ctx context.Context, key string) (*models.CachedOutput, error) {
	val, err := r.RedisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		logger.Debugf(ctx, "Item with key %s not found in redis", key)
		return nil, errors.NewNotFoundError("output", key)
	} else if err != nil {
		logger.Errorf(ctx, "Failed to get item from redis with error %v", err)
		return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to get item from redis with error %v", err)
	}

	var cachedOutput models.CachedOutput
	err = json.Unmarshal([]byte(val), &cachedOutput)
	if err != nil {
		logger.Errorf(ctx, "Failed to unmarshal output with error %v", err)
		return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to unmarshal output with error %v", err)
	}

	return &cachedOutput, nil
}

// Put will always set the value for the given key. It will overwrite the existing value if it exists.
func (r *redisClient) Put(ctx context.Context, key string, cachedOutput *models.CachedOutput) error {
	val, err := json.Marshal(cachedOutput)
	if err != nil {
		logger.Errorf(ctx, "Failed to marshal output with error %v", err)
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed to marshal output with error %v", err)
	}

	err = r.RedisClient.Set(ctx, key, val, 0).Err()
	if err != nil {
		logger.Errorf(ctx, "Failed updating redis for key %v with error %v", key, err)
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed updating redis for key %v with error %v", key, err)
	}

	return nil
}

// Delete is an idempotent operation. It will not return an error if the key does not exist.
func (r *redisClient) Delete(ctx context.Context, key string) error {
	_, err := r.RedisClient.Del(ctx, key).Result()
	if err != nil {
		logger.Errorf(ctx, "Failed deleting redis key %v with error %v", key, err)
		return errors.NewCacheServiceErrorf(codes.Internal, "Failed deleting redis key %v with error %v", key, err)
	}

	return nil
}

func NewCacheDataStore(ctx context.Context, serviceConfig configs.CacheServiceConfig) interfaces.CacheDataStoreClient {
	clientType := serviceConfig.DataStoreType

	switch clientType {
	case configs.Mem:
		cacheMap := make(map[string]*models.CachedOutput)
		return &memClient{cacheMap: &cacheMap}
	case configs.DynamoDB:
		cfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
			awsConfig.WithRegion(serviceConfig.AwsRegion),
		)
		if err != nil {
			panic("unable to load AWS config to connect to Dynamo" + err.Error())
		}
		return &dynamoClient{
			DynamoDbClient: dynamodb.NewFromConfig(cfg),
			TableName:      "cache",
		}
	case configs.Redis:
		rdb := redis.NewClient(
			&redis.Options{
				Addr:     serviceConfig.RedisAddress,
				Username: serviceConfig.RedisUsername,
				Password: serviceConfig.RedisPassword,
			})

		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			panic("failed to connect to redis " + err.Error())
		}

		return &redisClient{
			RedisClient: rdb,
		}
	}

	panic("unsupported cache data store type")
}
