"""Finite local load check. Creates and deletes only its own synthetic datasets."""
import argparse
import concurrent.futures
import copy
import json
import math
from pathlib import Path
import statistics
import time
import urllib.request
import uuid

parser = argparse.ArgumentParser()
parser.add_argument("--base-url", default="http://127.0.0.1:8080")
parser.add_argument("--samples", type=int, default=40)
parser.add_argument("--workers", type=int, default=4)
parser.add_argument("--output", default="docs/benchmarks/http-load.json")
args = parser.parse_args()
if not 1 <= args.samples <= 1000 or not 1 <= args.workers <= 16:
    parser.error("Use 1–1000 samples and 1–16 workers")
base = args.base_url.rstrip("/") + "/api"
source = json.loads((Path(__file__).resolve().parents[1] / "examples/demo.json").read_text(encoding="utf-8-sig"))

def request(path, method="GET", body=None, content_type="application/json"):
    if body is not None and not isinstance(body, bytes):
        body = json.dumps(body).encode()
    req = urllib.request.Request(base + path, data=body, method=method, headers={"Content-Type": content_type})
    with urllib.request.urlopen(req, timeout=60) as response:
        return json.load(response)

results = {"kind": "synthetic local technical load check", "samples_per_endpoint": args.samples, "concurrency": args.workers, "sizes": []}
for size in (12, 500, 2000):
    bundle = copy.deepcopy(source)
    bundle["employees"] = []
    bundle["activity_history"] = []
    bundle["meta"]["dataset"] = f"Temporary synthetic load check {size}"
    for i in range(size):
        person = copy.deepcopy(source["employees"][i % len(source["employees"])])
        person["employee_id"] = f"load-{i:05}"
        person["full_name"] = f"Synthetic profile {i}"
        person["manager_id"] = None
        bundle["employees"].append(person)
        for j in range(3):
            bundle["activity_history"].append({"record_id": f"load-{i}-{j}", "employee_id": person["employee_id"], "event_id": "story-circle", "date": f"2026-01-{j+1:02}", "status": "no_show", "completion_pct": 0, "assigned_by": "self"})
    boundary = "cq" + uuid.uuid4().hex
    raw = json.dumps(bundle, ensure_ascii=False).encode()
    body = (f'--{boundary}\r\nContent-Disposition: form-data; name="synthetic"\r\n\r\ntrue\r\n--{boundary}\r\nContent-Disposition: form-data; name="files"; filename="load.json"\r\nContent-Type: application/json\r\n\r\n'.encode() + raw + f'\r\n--{boundary}--\r\n'.encode())
    started = time.perf_counter()
    dataset = request("/import/commit", "POST", body, "multipart/form-data; boundary=" + boundary)
    item = {"profiles": size, "history_rows": size * 3, "input_bytes": len(raw), "import_ms": round((time.perf_counter() - started) * 1000, 2), "endpoints": {}}
    prefix = "/datasets/" + dataset["id"]
    try:
        for name, path in {"recommendation": prefix + "/employees/load-00000/analysis", "hr": prefix + "/hr", "dataset": prefix}.items():
            request(path)
            def measured(_):
                start = time.perf_counter()
                try:
                    request(path)
                    return (time.perf_counter() - start) * 1000, None
                except Exception as error:
                    return (time.perf_counter() - start) * 1000, type(error).__name__
            with concurrent.futures.ThreadPoolExecutor(max_workers=args.workers) as pool:
                samples = list(pool.map(measured, range(args.samples)))
            values = sorted(value for value, _ in samples)
            item["endpoints"][name] = {"median_ms": round(statistics.median(values), 2), "p95_ms": round(values[math.ceil(len(values)*.95)-1], 2), "max_ms": round(max(values), 2), "errors": sum(error is not None for _, error in samples)}
        results["sizes"].append(item)
        print(json.dumps(item))
    finally:
        request(prefix, "DELETE", {"revision": dataset["revision"]})
Path(args.output).write_text(json.dumps(results, indent=2) + "\n", encoding="utf-8")
if any(endpoint["errors"] for size in results["sizes"] for endpoint in size["endpoints"].values()):
    raise SystemExit("Load check contained failed requests; inspect the report")
