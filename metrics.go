// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

const dimensionless = "1"

var mWiredTigerCache = stats.Int64("wiredtiger.cache", "The WiredTiger engine cache stats", "By")
var mMemStats = stats.Int64("memory", "The memory statistics", "By")
var mBits = stats.Int64("bits", "The number of bits used", dimensionless)
var mNetworkStats = stats.Int64("network_stats", "Network statistics", "By")
var mRequests = stats.Int64("requests", "Number of requests", dimensionless)
var mConnectionStats = stats.Int64("connections", "Stats about incoming connections", dimensionless)
var mFlushStats = stats.Int64("flush_stats", "Flush statistics", "By")
var mFlushLatencyMs = stats.Float64("flush_average_latency", "Average flush latencies", "ms")
var mJournaledMB = stats.Int64("dur_mb", "Records the various statistics related to journaling", "MBy")
var mJournaledLatencyMs = stats.Float64("dur_latency_ms", "Records the various journaling latencies", "ms")
var mJournaledCommits = stats.Int64("dur_commits", "Records the number of journaled commits", dimensionless)
var mOpcounters = stats.Int64("op_counters", "Records number of operations related to commands and basic CRUD", dimensionless)
var mOpcountersRepl = stats.Int64("op_counters_repl", "Records number of operations related to commands and basic CRUD in the REPL", dimensionless)
var mDBRecordStats = stats.Int64("db_record_stats", "Records number of DB operations", dimensionless)

var keyType, _ = tag.NewKey("type")
var keyValue, _ = tag.NewKey("value")
var keyPID, _ = tag.NewKey("pid")
var keyHost, _ = tag.NewKey("host")
var keyVersion, _ = tag.NewKey("version")
var keyProcess, _ = tag.NewKey("process")

var compulsoryKeys = []tag.Key{keyVersion, keyHost, keyPID, keyProcess}

func withCompulsoryKeys(otherKeys ...tag.Key) []tag.Key {
	return append(compulsoryKeys, otherKeys...)
}

var defaultBytesDistribution = view.Distribution(0.0, 1024.0, 2048.0, 4096.0, 16384.0, 65536.0, 262144.0, 1048576.0, 4194304.0, 16777216.0, 67108864.0, 268435456.0, 1073741824.0, 2147483648.0, 4294967296.0, 8589934592.0, 17179869184.0, 34359738368.0, 68719476736.0)
var defaultMegabytesDistribution = view.Distribution(0.0, 5.0, 10.0, 25.0, 100.0, 200.0, 500.0, 1024, 1536.0, 2048, 4096, 8192, 16384, 32768, 65536, 131072, 262144, 524288, 1048576, 2097152, 4194304, 8388608)
var connectionCountDistribution = view.Distribution(0.0, 5.0, 10.0, 15.0, 20.0, 40.0, 60.0, 80.0, 100.0, 200.0, 400.0, 800.0, 1000.0, 1400.0, 1800.0, 2200.0, 2600.0, 3000.0, 3500.0, 4000.0, 4500.0, 5000.0, 5500.0, 6000.0, 6500.0, 7000.0, 7500.0, 8000.0, 8500.0, 9000.0, 9500.0, 10000.0, 14000.0, 18000.0, 20000.0, 25000.0, 30000.0, 35000.0, 40000.0, 45000.0, 50000)
var defaultMillisecondsDistribution = view.Distribution(
	// [0ms, 0.001ms, 0.005ms, 0.01ms, 0.05ms, 0.1ms, 0.5ms, 1ms, 1.5ms, 2ms, 2.5ms, 5ms, 10ms, 25ms, 50ms, 100ms, 200ms, 400ms, 600ms, 800ms, 1s, 1.5s, 2.5s, 5s, 10s, 20s, 40s, 100s, 200s, 500s, 750s, 1000s]
	0.0, 0.000001, 0.000005, 0.00001, 0.00005, 0.0001, 0.0005, 0.001, 0.0015, 0.002, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.2, 0.4, 0.6, 0.8, 1.0, 1.5, 2.5, 5.0, 10.0, 20.0, 40.0, 100.0, 200.0, 500.0, 750.0, 1000.0)

var allViews = []*view.View{
	{
		Name:        "mongodb/server/wiredtiger.cache",
		Measure:     mWiredTigerCache,
		Description: "WiredTiger engine cache stats",
		Aggregation: defaultBytesDistribution,
		TagKeys:     withCompulsoryKeys(keyType, keyValue),
	},
	{
		Name:        "mongodb/server/mem_stats",
		Measure:     mMemStats,
		Description: "Memory statistics",
		Aggregation: defaultBytesDistribution,
		TagKeys:     withCompulsoryKeys(keyType, keyValue),
	},
	{
		Name:        "mongodb/server/mem_stats.bits",
		Measure:     mBits,
		Description: "Memory statistics bits",
		Aggregation: view.Count(),
		TagKeys:     withCompulsoryKeys(keyType, keyValue),
	},
	{
		Name:        "mongodb/server/network",
		Measure:     mNetworkStats,
		Description: "Network statistics",
		Aggregation: defaultBytesDistribution,
		TagKeys:     withCompulsoryKeys(keyType, keyValue),
	},
	{
		Name:        "mongodb/server/requests",
		Measure:     mRequests,
		Description: "The number of requests",
		Aggregation: view.Count(),
		TagKeys:     withCompulsoryKeys(keyType, keyValue),
	},
	{
		Name:        "mongodb/server/connection",
		Measure:     mConnectionStats,
		Description: "Information about incoming connections",
		Aggregation: connectionCountDistribution,
		TagKeys:     withCompulsoryKeys(keyType),
	},
	{
		Name:        "mongodb/server/flush_stats",
		Measure:     mFlushStats,
		Description: "Information about memory flushes",
		Aggregation: defaultBytesDistribution,
		TagKeys:     withCompulsoryKeys(keyType),
	},
	{
		Name:        "mongodb/server/flush_latency",
		Measure:     mFlushLatencyMs,
		Description: "Latency information about memory flushes",
		Aggregation: defaultMillisecondsDistribution,
		TagKeys:     withCompulsoryKeys(keyType),
	},
	{
		Name:        "mongodb/server/dur_mb",
		Measure:     mJournaledMB,
		Description: "Various memory stats related to journaling",
		Aggregation: defaultMegabytesDistribution,
		TagKeys:     withCompulsoryKeys(keyType),
	},
	{
		Name:        "mongodb/server/dur_latency_ms",
		Measure:     mJournaledLatencyMs,
		Description: "Various latencies related to journaling",
		Aggregation: defaultMillisecondsDistribution,
		TagKeys:     withCompulsoryKeys(keyType),
	},
	{
		Name:        "mongodb/server/dur_commits",
		Measure:     mJournaledCommits,
		Description: "Various commits related to journaling",
		Aggregation: view.Count(),
		TagKeys:     withCompulsoryKeys(keyType),
	},
	{
		Name:        "mongodb/server/op_counters",
		Measure:     mOpcounters,
		Description: "Records number of operations related to commands and basic CRUD",
		Aggregation: view.Count(),
		TagKeys:     withCompulsoryKeys(keyType),
	},
	{
		Name:        "mongodb/server/op_counters_repl",
		Measure:     mOpcountersRepl,
		Description: "Records number of operations related to commands and basic CRUD using the REPL",
		Aggregation: view.Count(),
		TagKeys:     withCompulsoryKeys(keyType),
	},
}

func insertCompulsoryTagKeys(ctx context.Context, sv *ServerStatus) (context.Context, error) {
	// Firstly record some compulsory but disambiguating information such as:
	//      * PID
	//      * Host
	//      * Version
	//      * Process
	return tag.New(ctx,
		tag.Upsert(keyPID, fmt.Sprintf("%d", sv.Pid)),
		tag.Upsert(keyHost, sv.Host),
		tag.Upsert(keyVersion, sv.Version),
		tag.Upsert(keyProcess, sv.Process))
}

func (sv *ServerStatus) recordStats(ctx context.Context) context.Context {
	ctx, _ = insertCompulsoryTagKeys(ctx, sv)
	methods := []func(context.Context, *ServerStatus) context.Context{
		recordMemStats,
		recordNetworkStats,
		recordConnectionStats,
		recordWiredTigerCacheStats,
		recordFlushStats,
		recordJournalingStats,
		recordOpcounters,
		recordOpcountersRepl,
		recordDBRecordStats,
	}

	for _, method := range methods {
		ctx = method(ctx, sv)
	}
	return ctx
}

func recordWiredTigerCacheStats(ctx context.Context, sv *ServerStatus) context.Context {
	wt := sv.WiredTiger
	// Per scope, we'll redefine ctx to ensure that we can disambiguate between the various
	// key types e.g "transaction_checkpoints" vs "cache"
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "transaction_checkpoints"))
		stats.Record(ctx, mWiredTigerCache.M(wt.Transaction.TransCheckpoints))
	}

	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "cache"))
		{
			ctx, _ := tag.New(ctx, tag.Upsert(keyValue, "tracked_dirty_bytes"))
			stats.Record(ctx, mWiredTigerCache.M(wt.Cache.TrackedDirtyBytes))
		}
		{
			ctx, _ := tag.New(ctx, tag.Upsert(keyValue, "max_bytes_configured"))
			stats.Record(ctx, mWiredTigerCache.M(wt.Cache.MaxBytesConfigured))
		}
		{
			ctx, _ := tag.New(ctx, tag.Upsert(keyValue, "current_cached_bytes"))
			stats.Record(ctx, mWiredTigerCache.M(wt.Cache.CurrentCachedBytes))
		}
	}
	return ctx
}

func recordMemStats(ctx context.Context, sv *ServerStatus) context.Context {
	mem := sv.Mem
	if mem == nil {
		return ctx
	}
	ctx, _ = tag.New(ctx, tag.Upsert(keyType, "mem_stats"))
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyValue, "bits"))
		stats.Record(ctx, mBits.M(mem.Bits))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyValue, "resident"))
		stats.Record(ctx, mMemStats.M(mem.Resident))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyValue, "virtual"))
		stats.Record(ctx, mMemStats.M(mem.Virtual))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyValue, "mapped"))
		stats.Record(ctx, mMemStats.M(mem.Mapped))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyValue, "mapped_with_journal"))
		stats.Record(ctx, mMemStats.M(mem.MappedWithJournal))
	}

	return ctx
}

func recordNetworkStats(ctx context.Context, sv *ServerStatus) context.Context {
	nt := sv.Network
	if nt == nil {
		return ctx
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "bytes_in"))
		stats.Record(ctx, mNetworkStats.M(nt.BytesIn))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "bytes_out"))
		stats.Record(ctx, mNetworkStats.M(nt.BytesOut))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "num_requests"))
		stats.Record(ctx, mRequests.M(nt.NumRequests))
	}

	return ctx
}

func recordConnectionStats(ctx context.Context, sv *ServerStatus) context.Context {
	conns := sv.Connections
	if conns == nil {
		return ctx
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "available"))
		stats.Record(ctx, mConnectionStats.M(conns.Available))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "current"))
		stats.Record(ctx, mConnectionStats.M(conns.Current))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "total_created"))
		stats.Record(ctx, mConnectionStats.M(conns.TotalCreated))
	}

	return ctx
}

func recordFlushStats(ctx context.Context, sv *ServerStatus) context.Context {
	bgf := sv.BackgroundFlushing
	if bgf == nil {
		return ctx
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "total_flushes"))
		stats.Record(ctx, mFlushStats.M(bgf.Flushes))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "total_latency"))
		stats.Record(ctx, mFlushLatencyMs.M(float64(bgf.TotalMs)))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "average_latency"))
		stats.Record(ctx, mFlushLatencyMs.M(float64(bgf.AverageMs)))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "last_latency"))
		stats.Record(ctx, mFlushLatencyMs.M(float64(bgf.LastMs)))
	}

	return ctx
}

func recordJournalingStats(ctx context.Context, sv *ServerStatus) context.Context {
	dur := sv.Dur
	if dur == nil {
		return ctx
	}

	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "journaled"))
		stats.Record(ctx, mJournaledMB.M(dur.JournaledMB))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "write_to_data_files"))
		stats.Record(ctx, mJournaledMB.M(dur.WriteToDataFilesMB))
	}

	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "commits"))
		stats.Record(ctx, mJournaledCommits.M(dur.Commits))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "commits_in_write_lock"))
		stats.Record(ctx, mJournaledCommits.M(dur.CommitsInWriteLock))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "early_commits"))
		stats.Record(ctx, mJournaledCommits.M(dur.EarlyCommits))
	}

	// Journaling DurTiming related stats
	timing := dur.TimeMs
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "dt"))
		stats.Record(ctx, mJournaledLatencyMs.M(float64(timing.Dt)))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "prep_log_buffer"))
		stats.Record(ctx, mJournaledLatencyMs.M(float64(timing.PrepLogBuffer)))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "write_to_journal"))
		stats.Record(ctx, mJournaledLatencyMs.M(float64(timing.WriteToJournal)))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "write_to_data_files"))
		stats.Record(ctx, mJournaledLatencyMs.M(float64(timing.WriteToDataFiles)))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "remap_private_view"))
		stats.Record(ctx, mJournaledLatencyMs.M(float64(timing.RemapPrivateView)))
	}

	return ctx
}

func recordOpcounters(ctx context.Context, sv *ServerStatus) context.Context {
	ops := sv.Opcounters
	if ops == nil {
		return ctx
	}

	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "insert"))
		stats.Record(ctx, mOpcounters.M(ops.Insert))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "query"))
		stats.Record(ctx, mOpcounters.M(ops.Query))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "update"))
		stats.Record(ctx, mOpcounters.M(ops.Update))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "delete"))
		stats.Record(ctx, mOpcounters.M(ops.Delete))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "get_more"))
		stats.Record(ctx, mOpcounters.M(ops.GetMore))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "command"))
		stats.Record(ctx, mOpcounters.M(ops.Command))
	}

	return ctx
}

func recordOpcountersRepl(ctx context.Context, sv *ServerStatus) context.Context {
	opsr := sv.OpcountersRepl
	if opsr == nil {
		return ctx
	}

	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "insert"))
		stats.Record(ctx, mOpcountersRepl.M(opsr.Insert))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "query"))
		stats.Record(ctx, mOpcountersRepl.M(opsr.Query))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "update"))
		stats.Record(ctx, mOpcountersRepl.M(opsr.Update))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "delete"))
		stats.Record(ctx, mOpcountersRepl.M(opsr.Delete))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "get_more"))
		stats.Record(ctx, mOpcountersRepl.M(opsr.GetMore))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "command"))
		stats.Record(ctx, mOpcountersRepl.M(opsr.Command))
	}

	return ctx
}

func recordDBRecordStats(ctx context.Context, sv *ServerStatus) context.Context {
	rs := sv.RecordStats
	if rs == nil {
		return ctx
	}

	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "accesses_not_in_memroy"))
		stats.Record(ctx, mDBRecordStats.M(rs.AccessesNotInMemory))
	}
	{
		ctx, _ := tag.New(ctx, tag.Upsert(keyType, "page_fault_exceptions_thrown"))
		stats.Record(ctx, mDBRecordStats.M(rs.PageFaultExceptionsThrown))
	}

	return ctx
}
