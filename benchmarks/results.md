# Benchmark results

Measured on 2026-08-27 with Go 1.26.7 on Windows amd64, Intel Core i5-12400F (6 cores / 12 logical processors), and 31.8 GiB RAM. Each size was run three times with one timed operation per run:

```text
go test ./internal/importer ./internal/report -run '^$' \
  -bench 'Benchmark(ParseHAR|Report)(10k|50k|100k)$' \
  -benchtime=1x -benchmem -count=3
```

The table reports the median of three observed results, not a throughput guarantee.

| Operation | Entries | Median time | Median allocated bytes/op | Median allocations/op |
|---|---:|---:|---:|---:|
| Parse synthetic HAR | 10,000 | 52.618 ms | 35,987,960 | 279,875 |
| Parse synthetic HAR | 50,000 | 242.197 ms | 176,671,272 | 1,399,885 |
| Parse synthetic HAR | 100,000 | 475.139 ms | 357,567,224 | 2,799,880 |
| Generate offline report | 10,000 | 7.759 ms | 13,474,640 | 76 |
| Generate offline report | 50,000 | 49.063 ms | 58,951,592 | 73 |
| Generate offline report | 100,000 | 83.585 ms | 117,917,624 | 74 |

## Interpretation

The current HAR importer reads one JSON document and materializes normalized requests, so allocation growth is roughly linear and the 100k case allocates about 358 MB per operation. This is acceptable for a diagnostic workstation but does not meet a strict streaming target. The measured result supports the roadmap item to replace whole-document decoding with an iterative entry decoder.

Report generation serializes the request collection once into `data.js`. Operating-system file cache and garbage-collector state affect the one-operation samples, especially the 10k case. Repeat the command on the target platform before making capacity decisions.

