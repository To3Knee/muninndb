package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/scrypster/muninndb/internal/auth"
	"github.com/scrypster/muninndb/internal/engine/trigger"
	transportgrpc "github.com/scrypster/muninndb/internal/transport/grpc"
	pb "github.com/scrypster/muninndb/proto/gen/go/muninn/v1"
)

// mockEngine implements EngineAPI for testing. Every method returns a zero-value
// response and no error unless the test provides specific behaviour via the
// function fields.
type mockEngine struct {
	helloFn                func(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error)
	writeFn                func(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error)
	readFn                 func(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error)
	activateFn             func(ctx context.Context, req *pb.ActivateRequest) (*pb.ActivateResponse, error)
	linkFn                 func(ctx context.Context, req *pb.LinkRequest) (*pb.LinkResponse, error)
	forgetFn               func(ctx context.Context, req *pb.ForgetRequest) (*pb.ForgetResponse, error)
	statFn                 func(ctx context.Context, req *pb.StatRequest) (*pb.StatResponse, error)
	subscribeFn            func(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error)
	subscribeWithDeliverFn func(ctx context.Context, req *pb.SubscribeRequest, deliver trigger.DeliverFunc) (string, error)
	unsubscribeFn          func(ctx context.Context, subID string) error
}

func (m *mockEngine) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	if m.helloFn != nil {
		return m.helloFn(ctx, req)
	}
	return &pb.HelloResponse{ServerVersion: "test"}, nil
}

func (m *mockEngine) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	if m.writeFn != nil {
		return m.writeFn(ctx, req)
	}
	return &pb.WriteResponse{ID: "00000000000000000000000000"}, nil
}

func (m *mockEngine) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	if m.readFn != nil {
		return m.readFn(ctx, req)
	}
	return &pb.ReadResponse{}, nil
}

func (m *mockEngine) Activate(ctx context.Context, req *pb.ActivateRequest) (*pb.ActivateResponse, error) {
	if m.activateFn != nil {
		return m.activateFn(ctx, req)
	}
	return &pb.ActivateResponse{}, nil
}

func (m *mockEngine) Link(ctx context.Context, req *pb.LinkRequest) (*pb.LinkResponse, error) {
	if m.linkFn != nil {
		return m.linkFn(ctx, req)
	}
	return &pb.LinkResponse{OK: true}, nil
}

func (m *mockEngine) Forget(ctx context.Context, req *pb.ForgetRequest) (*pb.ForgetResponse, error) {
	if m.forgetFn != nil {
		return m.forgetFn(ctx, req)
	}
	return &pb.ForgetResponse{OK: true}, nil
}

func (m *mockEngine) Stat(ctx context.Context, req *pb.StatRequest) (*pb.StatResponse, error) {
	if m.statFn != nil {
		return m.statFn(ctx, req)
	}
	return &pb.StatResponse{}, nil
}

func (m *mockEngine) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	if m.subscribeFn != nil {
		return m.subscribeFn(ctx, req)
	}
	return &pb.SubscribeResponse{SubID: "sub-1", Status: "ok"}, nil
}

func (m *mockEngine) SubscribeWithDeliver(ctx context.Context, req *pb.SubscribeRequest, deliver trigger.DeliverFunc) (string, error) {
	if m.subscribeWithDeliverFn != nil {
		return m.subscribeWithDeliverFn(ctx, req, deliver)
	}
	return "mock-sub-id", nil
}

func (m *mockEngine) Unsubscribe(ctx context.Context, subID string) error {
	if m.unsubscribeFn != nil {
		return m.unsubscribeFn(ctx, subID)
	}
	return nil
}

// newTestAuthStore opens an in-memory pebble database and returns an auth.Store.
func newTestAuthStore(t *testing.T) *auth.Store {
	t.Helper()
	db, err := pebble.Open("", &pebble.Options{FS: vfs.NewMem()})
	if err != nil {
		t.Fatalf("open test auth db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return auth.NewStore(db)
}

// freePort returns an available TCP port on localhost.
func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not find free port: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

// TestServerStartStop creates a Server, calls Serve in a goroutine, verifies it
// is accepting TCP connections, then cancels the context and verifies clean shutdown.
func TestServerStartStop(t *testing.T) {
	addr := freePort(t)
	engine := &mockEngine{}
	srv := transportgrpc.NewServer(addr, engine, newTestAuthStore(t))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ctx)
	}()

	// Verify the server is accepting TCP connections within a reasonable window.
	var conn net.Conn
	var dialErr error
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, dialErr = net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if dialErr == nil {
			conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if dialErr != nil {
		t.Fatalf("could not connect to server at %s: %v", addr, dialErr)
	}

	// Cancel context to trigger graceful shutdown.
	cancel()

	select {
	case err := <-serveErr:
		if err != nil {
			t.Errorf("Serve returned unexpected error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("server did not shut down within 3 seconds after context cancellation")
	}
}

// TestGracefulShutdown starts a server, calls Shutdown with a reasonable timeout,
// and verifies it returns nil.
func TestGracefulShutdown(t *testing.T) {
	addr := freePort(t)
	engine := &mockEngine{}
	srv := transportgrpc.NewServer(addr, engine, newTestAuthStore(t))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- srv.Serve(ctx)
	}()

	// Wait for server to be up.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Call Shutdown with a 2-second timeout.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		t.Errorf("Shutdown returned unexpected error: %v", err)
	}
}

// TestEngineAPIInterface verifies at compile time that mockEngine satisfies the
// EngineAPI interface. This test contains no runtime assertions — if it compiles,
// the interface contract is met.
func TestEngineAPIInterface(t *testing.T) {
	// Verify that NewServer accepts our mock without a cast.
	engine := &mockEngine{}
	_ = transportgrpc.NewServer(":0", engine, newTestAuthStore(t))
}

// TestSubscribeWithDeliverInterface verifies at compile-time and runtime that:
//   - mockEngine satisfies the full EngineAPI interface including SubscribeWithDeliver
//   - The deliver func passed to SubscribeWithDeliver correctly channels pushes
//
// Note: the pb.* types in this project are hand-written Go structs with protobuf
// struct tags but without proto.Message implementation (ProtoReflect), so wire-level
// gRPC streaming cannot be tested end-to-end in unit tests. The server-side logic for
// the Subscribe streaming handler is exercised indirectly via the engine adapter and
// trigger system integration tests.
func TestSubscribeWithDeliverInterface(t *testing.T) {
	// Compile-time check: mockEngine satisfies transportgrpc.EngineAPI.
	var _ transportgrpc.EngineAPI = &mockEngine{}

	// Runtime: verify the SubscribeWithDeliver mock correctly invokes the deliver func.
	received := make(chan *trigger.ActivationPush, 4)
	var capturedDeliver trigger.DeliverFunc

	eng := &mockEngine{
		subscribeWithDeliverFn: func(ctx context.Context, req *pb.SubscribeRequest, deliver trigger.DeliverFunc) (string, error) {
			capturedDeliver = deliver
			return "test-sub-id", nil
		},
	}

	// Call SubscribeWithDeliver — in production this is called by grpc.Server.Subscribe.
	ctx := context.Background()
	req := &pb.SubscribeRequest{Vault: "default", PushOnWrite: true}
	deliver := func(ctx context.Context, push *trigger.ActivationPush) error {
		received <- push
		return nil
	}

	subID, err := eng.SubscribeWithDeliver(ctx, req, deliver)
	if err != nil {
		t.Fatalf("SubscribeWithDeliver: %v", err)
	}
	if subID != "test-sub-id" {
		t.Errorf("subID = %q, want test-sub-id", subID)
	}
	if capturedDeliver == nil {
		t.Fatal("deliver func was not captured")
	}

	// Simulate the trigger system calling the deliver func.
	push := &trigger.ActivationPush{
		SubscriptionID: subID,
		Trigger:        trigger.TriggerNewWrite,
		PushNumber:     1,
		At:             time.Now(),
	}
	if err := capturedDeliver(ctx, push); err != nil {
		t.Fatalf("capturedDeliver: %v", err)
	}

	select {
	case got := <-received:
		if got.SubscriptionID != subID {
			t.Errorf("push.SubscriptionID = %q, want %q", got.SubscriptionID, subID)
		}
		if string(got.Trigger) != string(trigger.TriggerNewWrite) {
			t.Errorf("push.Trigger = %q, want %q", got.Trigger, trigger.TriggerNewWrite)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for push")
	}
}
