# Local API and diagnostics

[OpenAPI JSON](../internal/httpapi/schema/openapi.json) documents every application API
route, request body, response and error envelope. The running server serves the same
file at `GET /api/openapi.json`. It uses
[OpenAPI 3.1.1](https://spec.openapis.org/oas/v3.1.1.html).

Regenerate with `go run ./cmd/openapi > internal/httpapi/schema/openapi.json`.
CI compares generated output, validates the document, and checks successful and
failing responses from the Docker service. The typed schemas describe the wire
structure; the domain validator additionally enforces dates, references and business
rules described in [the data contract](data-format.md).

## Example with PowerShell

```powershell
$base = 'http://127.0.0.1:8080/api'
$dataset = Invoke-RestMethod -Method Post "$base/demo"
$body = @{ revision = $dataset.revision; name = 'My synthetic workspace' } | ConvertTo-Json
$dataset = Invoke-RestMethod -Method Patch "$base/datasets/$($dataset.id)" -ContentType 'application/json' -Body $body
Invoke-RestMethod "$base/datasets/$($dataset.id)/employees/person-01/analysis"
```

Every mutation of an existing dataset requires its current `revision`. After a 409,
reload the dataset and review the newer state before retrying; do not blindly replace
the revision. `DELETE /datasets/{id}` accepts `{"revision": 3}` and cascades through
the dataset's entity rows and recommendation evidence. It cannot be undone without
an earlier export. The UI offers a backup link and explicit deletion confirmation.

## Errors

```json
{
  "code": "revision_conflict",
  "request_id": "example-correlation-id"
}
```

| HTTP status | Meaning |
| --- | --- |
| 400 | Malformed JSON or multipart upload |
| 403 | Cross-origin write rejected |
| 404 | Requested dataset, profile or activity not found |
| 409 | Stale revision, including a delete that no longer matches |
| 422 | Import validation or participation/goal rule rejected |
| 500 | Unexpected server or persistence failure |

Validation errors can include `issues` with `path`, `code` and `detail`. The UI shows
the request ID and the `X-Request-ID` response header carries the same value. IDs are
generated on the server rather than trusted from client input.

Run `docker compose logs -f app` to inspect JSON logs. Each request records the route
pattern, method, status, duration in milliseconds, bytes and ID. Logs omit request
bodies and query strings. This is local diagnostics, not a full tracing/metrics stack.

## Contract check

```sh
python -m venv .tools/contract
# Windows: .tools/contract/Scripts/python.exe
.tools/contract/bin/python -m pip install -r scripts/requirements-contract.txt
.tools/contract/bin/python scripts/check_contract.py --base-url http://127.0.0.1:8080
```

The check creates a fictional dataset, validates live responses and error IDs,
and deletes its own dataset. These Python packages are testing dependencies only;
the application runtime remains a Go binary and static React files.
