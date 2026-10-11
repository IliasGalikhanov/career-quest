"""Validate the OpenAPI document and optional live responses on synthetic data."""
import argparse
import json
from pathlib import Path
import urllib.error
import urllib.request
from jsonschema import Draft202012Validator
from openapi_spec_validator import validate

ROOT = Path(__file__).resolve().parents[1]
doc = json.loads((ROOT / "internal/httpapi/schema/openapi.json").read_text(encoding="utf-8-sig"))
validate(doc)
print("OpenAPI 3.1 document: valid")

def check(schema, value):
    Draft202012Validator({**schema, "components": doc["components"]}).validate(value)

check({"$ref": "#/components/schemas/Bundle"}, json.loads((ROOT / "examples/demo.json").read_text(encoding="utf-8-sig")))
parser = argparse.ArgumentParser()
parser.add_argument("--base-url")
args = parser.parse_args()
if args.base_url:
    base = args.base_url.rstrip("/") + "/api"

    def call(path, template, method="GET", body=None, expected=200):
        request = urllib.request.Request(base + path, data=None if body is None else json.dumps(body).encode(), method=method, headers={"Content-Type": "application/json"})
        try:
            response = urllib.request.urlopen(request, timeout=30)
        except urllib.error.HTTPError as error:
            response = error
        with response:
            data = json.load(response)
            assert response.status == expected, (path, response.status, data)
            assert response.headers.get("X-Request-ID"), "Missing diagnostic ID"
            responses = doc["paths"][template][method.lower()]["responses"]
            schema = responses.get(str(expected), responses["default"])["content"]["application/json"]["schema"]
            check(schema, data)
            if expected >= 400:
                assert data["request_id"] == response.headers["X-Request-ID"]
            return data

    call("/health", "/health")
    call("/datasets", "/datasets")
    dataset = call("/demo", "/demo", "POST", expected=201)
    prefix = "/datasets/" + dataset["id"]
    try:
        call(prefix, "/datasets/{id}")
        call(prefix + "/employees/person-01/analysis", "/datasets/{id}/employees/{employee}/analysis")
        call(prefix + "/hr?department=Platform", "/datasets/{id}/hr")
        call(prefix + "/export", "/datasets/{id}/export")
        call(prefix + "/evidence", "/datasets/{id}/evidence")
        dataset = call(prefix, "/datasets/{id}", "PATCH", {"revision": dataset["revision"], "name": "Contract check"})
        call(prefix, "/datasets/{id}", "PATCH", {"revision": 1, "name": "Stale edit"}, expected=409)
        dataset = call(prefix + "/participation", "/datasets/{id}/participation", "POST", {"revision": dataset["revision"], "employee_id": "person-01", "event_id": "architecture-lab"})
        record = dataset["bundle"]["activity_history"][-1]["record_id"]
        dataset = call(prefix + "/participation/" + record, "/datasets/{id}/participation/{record}", "PUT", {"revision": dataset["revision"], "status": "completed", "reflection": "Contract round trip on fictional data."})
    finally:
        latest = call(prefix, "/datasets/{id}")
        call(prefix, "/datasets/{id}", "DELETE", {"revision": latest["revision"]})
    call(prefix, "/datasets/{id}", expected=404)
    print("Live response schemas, errors and request IDs: passed")
