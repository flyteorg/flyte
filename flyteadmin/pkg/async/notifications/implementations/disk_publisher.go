package implementations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"

	pbcloudevents "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// FunctionCounter Create a global thread-safe counter that can be incremented with every publish
type FunctionCounter struct {
	mu    sync.Mutex
	count int
}

// Method on the struct that increments the count and performs the function's actions
func (fc *FunctionCounter) increment() int {
	fc.mu.Lock()
	defer func() {
		fc.count++
		fc.mu.Unlock()
	}()
	return fc.count
}

// DebugOnlyDiskEventPublisher implements pubsub.Publisher by writing to a local tmp dir
// To use, add:
//
//	case "disk":
//		publisher := implementations.NewDebugOnlyDiskEventPublisher()
//		return implementations.NewEventsPublisher(publisher, scope, []string{"all"})
//
// to factory.go and update your local configuration
type DebugOnlyDiskEventPublisher struct {
	Root string
	c    FunctionCounter
}

func (d *DebugOnlyDiskEventPublisher) generateFileName(key string) string {
	c := d.c.increment()
	return fmt.Sprintf("%s/evt_%03d_%s.pb", d.Root, c, key)
}

// Publish Implement the pubsub.Publisher interface
func (d *DebugOnlyDiskEventPublisher) Publish(ctx context.Context, key string, msg proto.Message) error {
	mb, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	fname := d.generateFileName(key)
	logger.Warningf(ctx, "DSKPUB: Publish [%s - %s] [%+v]", fname, key, msg.String())

	// #nosec
	return os.WriteFile(fname, mb, 0666)
}

func (d *DebugOnlyDiskEventPublisher) PublishRaw(ctx context.Context, key string, msg []byte) error {
	fname := d.generateFileName(key)
	logger.Warningf(ctx, "DSKPUB: PublishRaw [%s - %s] [%+v]", fname, key, msg)

	// #nosec
	return os.WriteFile(fname, msg, 0666)
}

func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GetTempFolder() string {
	tmpDir := os.Getenv("TMPDIR")
	// make a tmp folder
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		err := os.Mkdir(tmpDir, 0755)
		if err != nil {
			panic(err)
		}
	}
	r, err := generateRandomString(3)
	if err != nil {
		panic(err)
	}

	// Has a / already
	folder := fmt.Sprintf("%s%s", tmpDir, r)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		err := os.Mkdir(folder, 0755)
		if err != nil {
			panic(err)
		}
	}
	return folder
}

// NewDebugOnlyDiskEventPublisher Create a new DebugOnlyDiskEventPublisher
func NewDebugOnlyDiskEventPublisher() *DebugOnlyDiskEventPublisher {
	// make a random temp sub folder
	folder := GetTempFolder()
	logger.Warningf(context.Background(), "DSKPUB: Using disk publisher with root [%s] -----", folder)

	return &DebugOnlyDiskEventPublisher{
		Root: folder,
		c:    FunctionCounter{},
	}
}

// DebugOnlyDiskSenderForCloudEvents implements the Sender interface
// Combines the Sender and Publisher into one, not sure why there are two.
// To use, add:
//
//	case "disksender":
//	    sender = implementations.NewDebugOnlyDiskSenderForCloudEvents()
//
// to cloudevents/factory.go and update your local configuration. This captures all the transformations that the cloud
// event publisher does to the various events.
type DebugOnlyDiskSenderForCloudEvents struct {
	Root string
	c    FunctionCounter
}

func (s *DebugOnlyDiskSenderForCloudEvents) Send(ctx context.Context, notificationType string, event cloudevents.Event) error {
	fname := s.generateFileName(notificationType, event.Source())
	logger.Warningf(ctx, "DSKPUB: Send [%s - %s] [%+v]", fname, notificationType, event.Source())

	eventBytes, err := pbcloudevents.Protobuf.Marshal(&event)
	if err != nil {
		logger.Errorf(ctx, "Failed to marshal cloudevent with error: %v", err)
		panic(err)
	}
	// #nosec
	return os.WriteFile(fname, eventBytes, 0666)

	//return nil
}

func (s *DebugOnlyDiskSenderForCloudEvents) generateFileName(notificationType, key string) string {
	// The source comes in the form of fce772610fcfc4442a34/n0-0-n0-0-start-node
	// This strips out everything before the /
	nodeID := strings.Split(key, "/")[1]
	c := s.c.increment()
	return fmt.Sprintf("%s/cloud_%03d_%s_%s.pb", s.Root, c, notificationType, nodeID)
}

func NewDebugOnlyDiskSenderForCloudEvents() *DebugOnlyDiskSenderForCloudEvents {
	// make a random temp sub folder
	folder := GetTempFolder()
	logger.Warningf(context.Background(), "DSKPUB: Using disk sender with root [%s] -----", folder)

	return &DebugOnlyDiskSenderForCloudEvents{
		Root: folder,
		c:    FunctionCounter{},
	}
}
