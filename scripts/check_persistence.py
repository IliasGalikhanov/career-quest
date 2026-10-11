"""Black-box check: a completed activity survives container recreation.

Uses only the Python standard library. Does not remove the Docker data volume.
"""

import json
import subprocess
import urllib.request


BASE = "http://127.0.0.1:8080/api"


def request(path, method="GET", payload=None):
    data = None if payload is None else json.dumps(payload).encode()
    req = urllib.request.Request(
        BASE + path, data=data, method=method,
        headers={"Content-Type": "application/json"},
    )
    with urllib.request.urlopen(req, timeout=15) as response:
        return json.load(response)


dataset = request("/demo", "POST")
root = "/datasets/" + dataset["id"]
dataset = request(root + "/participation", "POST", {
    "revision": dataset["revision"],
    "employee_id": "person-01", "event_id": "architecture-lab",
})
record = dataset["bundle"]["activity_history"][-1]["record_id"]
dataset = request(root + "/participation/" + record, "PUT", {
    "revision": dataset["revision"], "status": "completed",
})
before = request(root + "/employees/person-01/analysis")
assert before["levels"]["architecture"] == 3
subprocess.run(["docker", "compose", "down"], check=True)
subprocess.run(["docker", "compose", "up", "-d", "--wait"], check=True)
after = request(root + "/employees/person-01/analysis")
restored = request(root)
assert restored["revision"] == dataset["revision"]
assert restored["bundle"] == dataset["bundle"]
assert after["levels"] == before["levels"]
assert after["coverage"] == before["coverage"]
print("PASS: dataset, revision, history, skill gains and coverage survived container recreation")
