"""Measure local HTTP latency on the bundled synthetic example, not model quality."""

import json
import math
import statistics
import time
import urllib.request


BASE = "http://127.0.0.1:8080/api"


def request(path, method="GET"):
    req = urllib.request.Request(BASE + path, method=method)
    with urllib.request.urlopen(req, timeout=15) as response:
        return json.load(response)


dataset = request("/demo", "POST")
root = "/datasets/" + dataset["id"]
result = {
    "dataset": "independent synthetic demo",
    "profiles": len(dataset["bundle"]["employees"]),
    "events": len(dataset["bundle"]["events"]),
    "skills": len(dataset["bundle"]["skills"]),
    "samples": 30,
    "warmups": 3,
    "concurrency": 1,
    "endpoints": {},
}
for name, endpoint in {
    "dataset": root,
    "recommendation_with_evidence_write": root + "/employees/person-01/analysis",
    "hr": root + "/hr",
}.items():
    for _ in range(3):
        request(endpoint)
    values = []
    for _ in range(30):
        start = time.perf_counter()
        request(endpoint)
        values.append((time.perf_counter() - start) * 1000)
    result["endpoints"][name] = {
        "median_ms": round(statistics.median(values), 2),
        "p95_ms": round(sorted(values)[math.ceil(len(values) * .95) - 1], 2),
        "max_ms": round(max(values), 2),
    }
print(json.dumps(result, indent=2))
