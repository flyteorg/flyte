package plugin

import (
	"sync"
	"time"

	"github.com/unionai/flyte/fasttask/plugin/interfaces"
	"github.com/unionai/flyte/fasttask/plugin/pb"
)

type environmentImpl struct {
	createdAt               int64
	envID                   interfaces.ExecutionEnvID
	failureMessage          string
	fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec
	lastScaledDownAt        int64
	lock                    sync.RWMutex
	state                   interfaces.State
	workers                 *sync.Map // map[string]*worker
}

var _ interfaces.Environment = (*environmentImpl)(nil)

func (e *environmentImpl) GetWorker(workerID string) interfaces.Worker {
	workerValue, exists := e.workers.Load(workerID)
	if !exists {
		return nil
	}

	worker := workerValue.(interfaces.Worker)
	return worker
}

func (e *environmentImpl) GetOrCreateWorker(workerID string) interfaces.Worker {
	workerValue, loaded := e.workers.Load(workerID)
	if !loaded {
		worker := &workerImpl{
			id:             workerID,
			lastAccessedAt: time.Now().Unix(),
			lock:           sync.RWMutex{},
			responseChan:   make(chan *pb.HeartbeatResponse, GetConfig().HeartbeatBufferSize),
			state:          interfaces.INITIALIZING,
		}

		workerValue, _ = e.workers.LoadOrStore(workerID, worker)
	}

	worker := workerValue.(interfaces.Worker)
	return worker
}

func (e *environmentImpl) CreatedAt() int64 {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.createdAt
}

func (e *environmentImpl) EnvID() interfaces.ExecutionEnvID {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.envID
}

func (e *environmentImpl) FailureMessage() string {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.failureMessage
}

func (e *environmentImpl) SetFailureMessage(failureMessage string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.failureMessage = failureMessage
}

func (e *environmentImpl) LastScaledDownAt() int64 {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.lastScaledDownAt
}

func (e *environmentImpl) SetLastScaledDownAt(lastScaledDownAt int64) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.lastScaledDownAt = lastScaledDownAt
}

func (e *environmentImpl) State() interfaces.State {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.state
}

func (e *environmentImpl) SetState(state interfaces.State) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.state = state
}

func (e *environmentImpl) Recover(envID interfaces.ExecutionEnvID, fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.envID = envID
	e.fastTaskEnvironmentSpec = fastTaskEnvironmentSpec
}

func (e *environmentImpl) FastTaskEnvironmentSpec() *pb.FastTaskEnvironmentSpec {
	return e.fastTaskEnvironmentSpec
}

func (e *environmentImpl) RangeWorkers(fn func(workerID string, worker interfaces.Worker) bool) {
	e.workers.Range(func(key, value interface{}) bool {
		return fn(key.(string), value.(interfaces.Worker))
	})
}

func (e *environmentImpl) DeleteWorker(workerID string) {
	e.workers.Delete(workerID)
}

type workerImpl struct {
	capacity       *pb.Capacity
	id             string
	lastAccessedAt int64
	lock           sync.RWMutex
	responseChan   chan *pb.HeartbeatResponse
	state          interfaces.State
}

var _ interfaces.Worker = (*workerImpl)(nil)

func (w *workerImpl) Capacity() *pb.Capacity {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.capacity
}

func (w *workerImpl) SetCapacity(capacity *pb.Capacity) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.capacity = capacity
}

func (w *workerImpl) LastAccessedAt() int64 {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.lastAccessedAt
}

func (w *workerImpl) SetLastAccessedAt(lastAccessedAt int64) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.lastAccessedAt = lastAccessedAt
}

func (w *workerImpl) State() interfaces.State {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.state
}

func (w *workerImpl) SetState(state interfaces.State) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.state = state
}

func (w *workerImpl) ID() string {
	return w.id
}

func (w *workerImpl) EnqueueHeartbeatResponse(response *pb.HeartbeatResponse) {
	w.responseChan <- response
}

func (w *workerImpl) Responses() <-chan *pb.HeartbeatResponse {
	return w.responseChan
}

type inMemoryStore struct {
	environments *sync.Map // map[string]*Environment
}

var _ interfaces.EnvironmentStore = (*inMemoryStore)(nil)

func (i *inMemoryStore) Delete(executionEnvID string) {
	i.environments.Delete(executionEnvID)
}

func (i *inMemoryStore) Get(executionEnvID string) interfaces.Environment {
	envValue, exists := i.environments.Load(executionEnvID)
	if !exists {
		return nil
	}

	env, _ := envValue.(interfaces.Environment)
	return env
}

func (i *inMemoryStore) GetOrCreate(executionEnvID string, env interfaces.Environment) interfaces.Environment {
	actual, _ := i.environments.LoadOrStore(executionEnvID, env)
	return actual.(interfaces.Environment)
}

func (i *inMemoryStore) List() []interfaces.Environment {
	var envs []interfaces.Environment
	i.environments.Range(func(_, value interface{}) bool {
		env, _ := value.(interfaces.Environment)
		envs = append(envs, env)
		return true
	})
	return envs
}

func newEnvironmentStore() interfaces.EnvironmentStore {
	return &inMemoryStore{
		environments: &sync.Map{},
	}
}
