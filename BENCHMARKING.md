# Xomoi-Core: Performance & Benchmarking

Xomoi-Core is designed as a Sovereign Edge Node—capable of massive ingestion on low-power, constrained hardware without sprawling memory or triggering Go Garbage Collector (GC) latency spikes.

This document outlines the internal benchmarking methodology and the certified limits we discovered when testing the architecture on legacy hardware.

## Methodology

Our custom `benchmark.go` tool tests the absolute limits of the broker across three modal vectors:
1. **Ingestion (Stress Test):** Floods the broker with concurrent publishers.
2. **End-to-End Latency:** Measures the precise millisecond round-trip delay of a single packet from publisher, through the broker's processing engine, and out to a subscriber.
3. **QoS Verification:** Tests both QoS 0 (Fire and Forget) and QoS 1 (Guaranteed Delivery with `PUBACK`).

The tests below were conducted on a legacy **Intel Pentium** edge node with 4 cores, communicating over a Tailscale WireGuard VPN.

---

## 1. Zero-Allocation Pipeline & GC Performance

The core philosophy of Xomoi is bypassing the Go Garbage Collector using `sync.Pool` and zero-allocation libraries (`buger/jsonparser`). 

During a 60-second stress test parsing 350,000+ Zstandard-compressed JSON packets, the broker achieved:
* **Max RAM Allocated:** 182.43 MB
* **Total GC Triggers:** 19 (Over 350,000 packets)
* **Max GC Pause Latency:** 7.74 ms
* **Max Goroutine Sprawl:** 1019

The GC metrics conclusively prove that the architecture is immune to the memory fragmentation and OOM crashes typical in Node.js or Python IoT stacks.

---

## 2. End-to-End Latency

**Command:** `go run scripts/benchmark.go -ip <IP> -mode latency -qos 1 -time 15`

When testing a single packet's journey (Publisher -> VPN -> Broker -> JSON Parse -> HotState -> Subscriber -> VPN), the broker introduced virtually zero latency.
* **Average Round-Trip Latency:** `1.47 ms` (Localhost) / `7.46 ms` (Cross-Internet VPN)

This proves the ingestion pipeline and the `mochi-mqtt` router operate at the physical speed limit of the network.

---

## 3. The Disk I/O Bottleneck & Ring Buffer Tuning

The theoretical maximum throughput of the CPU is significantly higher than the physical write speed of an SSD/HDD. Xomoi uses an in-memory **Ring Buffer** to absorb packet tsunamis and flush them to SQLite in bulk transactions.

**Benchmark Results (QoS 1, 500 Workers, 120s):**
* **Batch Size 1,000:** `3,649 msgs/sec`
* **Batch Size 10,000:** `5,586 msgs/sec`
* **Batch Size 50,000:** `5,052 msgs/sec`

### The "Goldilocks Zone"
By exposing `XOMOI_RING_BATCH_SIZE` as an environment variable, we found that **10,000** is the optimal transaction batch size. 
If the batch is too small (1,000), the OS spends too much time waiting for disk `fsync()` confirmations.
If the batch is too large (50,000), the massive SQL query string chokes the Go-to-CGO boundary before it ever reaches SQLite.

### Backpressure Survival
When hammered with 1,000 concurrent workers, the Ring Buffer successfully saturated its 100,000 packet limit and mathematically engaged TCP backpressure. Instead of attempting to hold the packets in RAM and crashing the runtime, the node throttled the network down to precisely what the physical hard drive could ingest.

---

## Summary

Xomoi-Core is certified to sustain **5,500+ guaranteed QoS 1 messages per second** on legacy Pentium hardware without memory degradation, proving it is fully hardened for enterprise-scale edge deployments.
