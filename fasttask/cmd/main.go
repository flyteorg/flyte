package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/unionai/flyte/fasttask/cmd/pb"
	"google.golang.org/grpc"
)

type Task struct {
	TaskID    string            `json:"task_id"`
	Namespace string            `json:"namespace"`
	Command   []string          `json:"command"`
	EnvVars   map[string]string `json:"env_vars"`
}

type FastTaskServer struct {
	pb.UnimplementedFastTaskServer
	tasks      []Task
	taskQueue  chan Task
	mu         sync.RWMutex
	assignedTasks map[string]*Task
}

func NewFastTaskServer(tasksFile string) (*FastTaskServer, error) {
	server := &FastTaskServer{
		taskQueue:     make(chan Task, 100),
		assignedTasks: make(map[string]*Task),
	}
	
	if err := server.loadTasks(tasksFile); err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}
	
	go server.fillTaskQueue()
	
	return server, nil
}

func (s *FastTaskServer) loadTasks(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read tasks file: %w", err)
	}
	
	if err := json.Unmarshal(data, &s.tasks); err != nil {
		return fmt.Errorf("failed to parse tasks JSON: %w", err)
	}
	
	log.Printf("Loaded %d tasks from %s", len(s.tasks), filename)
	return nil
}

func (s *FastTaskServer) fillTaskQueue() {
	for _, task := range s.tasks {
		s.taskQueue <- task
	}
}

func (s *FastTaskServer) Heartbeat(stream pb.FastTask_HeartbeatServer) error {
	log.Println("Client connected for heartbeat stream")
	
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("Client disconnected")
			return nil
		}
		if err != nil {
			log.Printf("Error receiving heartbeat: %v", err)
			return err
		}
		
		log.Printf("Received heartbeat from worker %s, queue %s", req.WorkerId, req.QueueId)
		log.Printf("Worker capacity: %d/%d active, %d/%d backlog", 
			req.Capacity.ExecutionCount, req.Capacity.ExecutionLimit,
			req.Capacity.BacklogCount, req.Capacity.BacklogLimit)
		
		// Process task statuses
		for _, status := range req.TaskStatuses {
			log.Printf("Task %s in namespace %s is in phase %d: %s", 
				status.TaskId, status.Namespace, status.Phase, status.Reason)
		}
		
		// Send tasks based on available capacity
		availableCapacity := req.Capacity.ExecutionLimit - req.Capacity.ExecutionCount
		availableBacklog := req.Capacity.BacklogLimit - req.Capacity.BacklogCount
		
		totalAvailable := availableCapacity + availableBacklog
		
		for i := 0; i < int(totalAvailable); i++ {
			select {
			case task := <-s.taskQueue:
				response := &pb.HeartbeatResponse{
					TaskId:    task.TaskID,
					Namespace: task.Namespace,
					Cmd:       task.Command,
					Operation: pb.HeartbeatResponse_ASSIGN,
					EnvVars:   task.EnvVars,
					ExecId: &pb.ExecutionIdentifier{
						Org:     task.EnvVars["_U_ORG_NAME"],
						Project: task.EnvVars["FLYTE_INTERNAL_TASK_PROJECT"],
						Domain:  task.EnvVars["FLYTE_INTERNAL_TASK_DOMAIN"],
						Name:    task.TaskID,
					},
				}
				
				s.mu.Lock()
				s.assignedTasks[task.TaskID] = &task
				s.mu.Unlock()
				
				if err := stream.Send(response); err != nil {
					log.Printf("Error sending task assignment: %v", err)
					// Put task back in queue
					s.taskQueue <- task
					return err
				}
				
				log.Printf("Assigned task %s to worker %s", task.TaskID, req.WorkerId)
				
			default:
				// No more tasks available
				goto done
			}
		}
		
		done:
		// Send ACK if no tasks were assigned
		if totalAvailable == 0 {
			ackResponse := &pb.HeartbeatResponse{
				Operation: pb.HeartbeatResponse_ACK,
			}
			if err := stream.Send(ackResponse); err != nil {
				log.Printf("Error sending ACK: %v", err)
				return err
			}
		}
	}
}

func main() {
	var (
		port      = flag.Int("port", 8080, "gRPC server port")
		tasksFile = flag.String("tasks", "testdata/tasks.json", "Path to tasks JSON file")
	)
	flag.Parse()
	
	server, err := NewFastTaskServer(*tasksFile)
	if err != nil {
		log.Fatalf("Failed to create FastTask server: %v", err)
	}
	
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	
	grpcServer := grpc.NewServer()
	pb.RegisterFastTaskServer(grpcServer, server)
	
	log.Printf("FastTask server starting on port %d", *port)
	log.Printf("Loading tasks from: %s", *tasksFile)
	
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}