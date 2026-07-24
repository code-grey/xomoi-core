# Mochi-MQTT Open Source PR Strategy

**Target:** Upstream `mochi-mqtt/server` repository.
**Goal:** Submit a surgical Pull Request to eliminate O(N) map allocations during subscriber routing, drastically reducing Garbage Collection (GC) pauses during massive fan-outs.

## 1. The Core Issue
During our extreme Phase B Fanout benchmarks (2,000 subscribers, 50 publishers), the Go Garbage Collector choked with 11ms "Stop The World" pauses. We traced the memory leak to `topics.go` and `server.go`. Mochi-MQTT's routing engine was calling `GetAll()` on subscriptions for every published message. `GetAll()` allocates a brand new Go map and copies the entire subscriber list into it. With 50 publishers sending 10 msgs/sec to 2,000 subscribers, this triggered **1,000,000 map allocations per second**.

## 2. The Surgical Fix (To be PR'd)
We will introduce a zero-allocation `SelectInto` interface.
Instead of:
```go
// O(N) Allocation
func (s *Subscriptions) GetAll() map[string]packets.Subscription
```
We propose:
```go
// O(1) Pointer Injection
func (s *Subscriptions) SelectInto(out *[]*packets.Subscription)
```
The Radix Tree will accept a pointer to a pre-allocated slice and inject subscriber pointers directly into it, bypassing `make(map)` entirely. 

## 3. What NOT to PR
We will **not** PR our Reactor Pattern (the User-Space connection channel and Worker Pool). That architecture is highly tuned for low-power Edge Nodes to bypass OS TCP backlogs (`SOMAXCONN`). Core maintainers will likely argue that standard users should just tune their Linux `sysctl` settings rather than complicating the TCP Accept loop. We keep the Reactor Pattern exclusively in the Xomoi fork.

## 4. The Execution Plan
1. **Fork & Clone:** Fork the official `mochi-mqtt/server` repo to a fresh directory.
2. **Apply Changes:** Carefully extract the `SelectInto` modifications from our `xomoi-core/mochi-mqtt-fork` and apply them to the clean fork.
3. **Internal Test Suite:** Run `go test ./...` to ensure we didn't break any of Mochi's internal routing logic.
4. **Allocation Benchmarks:** Run `go test -bench=. -benchmem` to mathematically prove that our changes drop `allocs/op` to zero on the hot path.
5. **Draft the PR:** Write a compelling PR description citing our 73% reduction in GC pauses (from 11ms to 3ms) and the massive fanout throughput increase.
