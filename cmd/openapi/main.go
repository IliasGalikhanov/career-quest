// Command openapi writes the portable API contract from the domain's Go types.
package main

import (
	"careerquest/internal/domain"
	"careerquest/internal/store"
	"encoding/json"
	"os"
	"reflect"
	"strings"
)

type object = map[string]any

var definitions = object{}

func ref(name string) object { return object{"$ref": "#/components/schemas/" + name} }
func schema(t reflect.Type) object {
	if t.Kind() == reflect.Pointer {
		return object{"anyOf": []any{schema(t.Elem()), object{"type": "null"}}}
	}
	switch t.Kind() {
	case reflect.String:
		return object{"type": "string"}
	case reflect.Bool:
		return object{"type": "boolean"}
	case reflect.Int, reflect.Int64:
		return object{"type": "integer"}
	case reflect.Float64:
		return object{"type": "number"}
	case reflect.Map:
		return object{"type": []string{"object", "null"}, "additionalProperties": schema(t.Elem())}
	case reflect.Slice:
		return object{"type": []string{"array", "null"}, "items": schema(t.Elem())}
	case reflect.Struct:
		if _, ok := definitions[t.Name()]; ok {
			return ref(t.Name())
		}
		definitions[t.Name()] = object{}
		properties := object{}
		required := []string{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := strings.Split(field.Tag.Get("json"), ",")
			if tag[0] == "" || tag[0] == "-" {
				continue
			}
			properties[tag[0]] = schema(field.Type)
			if len(tag) == 1 {
				required = append(required, tag[0])
			}
		}
		definitions[t.Name()] = object{"type": "object", "properties": properties, "required": required}
		return ref(t.Name())
	}
	panic("unsupported schema type: " + t.String())
}
func obj(properties object, required ...string) object {
	return object{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}
func array(item any) object { return object{"type": "array", "items": item} }
func main() {
	for _, v := range []any{domain.Bundle{}, domain.Analysis{}, domain.Issue{}, store.Dataset{}, store.Summary{}} {
		schema(reflect.TypeOf(v))
	}
	str := object{"type": "string"}
	integer := object{"type": "integer"}
	revision := object{"type": "integer", "minimum": 1}
	definitions["Problem"] = obj(object{"code": str, "request_id": str, "issues": array(ref("Issue"))}, "code", "request_id")
	definitions["Preview"] = obj(object{"name": str, "employees": integer, "events": integer, "skills": integer, "history": integer, "as_of_date": str, "uncertain_dates": integer}, "name", "employees", "events", "skills", "history", "as_of_date", "uncertain_dates")
	counts := object{"type": "object", "additionalProperties": integer}
	definitions["HR"] = obj(object{"people": integer, "date": str, "statuses": counts, "gaps": array(obj(object{"skill_id": str, "people": integer, "eligible": integer, "total_gap": integer}, "skill_id", "people", "eligible", "total_gap")), "without_step": array(obj(object{"employee_id": str, "name": str, "reasons": array(str)}, "employee_id", "name", "reasons")), "participation": array(obj(object{"event_id": str, "statuses": counts}, "event_id", "statuses"))}, "people", "date", "statuses", "gaps", "without_step", "participation")
	paths := object{}
	add := func(path, method, id, summary string, response any, body any, status string) {
		responseOf := func(description string, s any) object {
			return object{"description": description, "headers": object{"X-Request-ID": object{"schema": str, "description": "Server-generated correlation ID; also included in error bodies."}}, "content": object{"application/json": object{"schema": s}}}
		}
		op := object{"operationId": id, "summary": summary, "responses": object{status: responseOf("Success", response), "default": responseOf("400 malformed request; 403 cross-origin write; 404 missing record; 409 stale revision; 422 validation or domain rule; 500 internal failure.", ref("Problem"))}}
		params := []any{}
		for _, name := range []string{"id", "employee", "record"} {
			if strings.Contains(path, "{"+name+"}") {
				params = append(params, object{"name": name, "in": "path", "required": true, "schema": str})
			}
		}
		if id == "hr" {
			for _, name := range []string{"department", "role"} {
				params = append(params, object{"name": name, "in": "query", "schema": str})
			}
		}
		if len(params) > 0 {
			op["parameters"] = params
		}
		if body != nil {
			media := "application/json"
			if strings.HasPrefix(path, "/import/") {
				media = "multipart/form-data"
			}
			op["requestBody"] = object{"required": true, "content": object{media: object{"schema": body}}}
		}
		if paths[path] == nil {
			paths[path] = object{}
		}
		paths[path].(object)[method] = op
	}
	add("/health", "get", "health", "Check SQLite availability", obj(object{"status": str, "algorithm": str}, "status", "algorithm"), nil, "200")
	add("/openapi.json", "get", "contract", "Download this OpenAPI contract", object{"type": "object"}, nil, "200")
	add("/datasets", "get", "datasets", "List local datasets", array(ref("Summary")), nil, "200")
	add("/demo", "post", "demo", "Create an independent synthetic workspace", ref("Dataset"), nil, "201")
	add("/example", "get", "example", "Download the independent synthetic input bundle", ref("Bundle"), nil, "200")
	add("/datasets/{id}", "get", "dataset", "Read a complete workspace snapshot", ref("Dataset"), nil, "200")
	add("/datasets/{id}", "patch", "rename", "Rename with optimistic concurrency", ref("Dataset"), obj(object{"revision": revision, "name": object{"type": "string", "minLength": 1, "maxLength": 120}}, "revision", "name"), "200")
	add("/datasets/{id}", "delete", "delete", "Delete dataset and related records; requires current revision", obj(object{"deleted": object{"type": "boolean"}}, "deleted"), obj(object{"revision": revision}, "revision"), "200")
	add("/datasets/{id}/export", "get", "export", "Export domain state; recommendation evidence is separate", ref("Bundle"), nil, "200")
	add("/datasets/{id}/evidence", "get", "evidence", "Export up to 500 recorded recommendation runs", array(obj(object{"revision": integer, "employee_id": str, "algorithm": str, "created_at": str, "evidence": obj(object{"settings": ref("Settings"), "analysis": ref("Analysis")}, "settings", "analysis")}, "revision", "employee_id", "algorithm", "created_at", "evidence")), nil, "200")
	add("/datasets/{id}/employees/{employee}/analysis", "get", "analysis", "Compute recommendations and append bounded diagnostic evidence", ref("Analysis"), nil, "200")
	add("/datasets/{id}/employees/{employee}/goal", "put", "goal", "Update or clear the career goal", ref("Dataset"), obj(object{"revision": revision, "goal": object{"anyOf": []any{ref("Goal"), object{"type": "null"}}}}, "revision", "goal"), "200")
	add("/datasets/{id}/settings", "put", "settings", "Update explicit calculation settings", ref("Dataset"), obj(object{"revision": revision, "settings": ref("Settings")}, "revision", "settings"), "200")
	add("/datasets/{id}/participation", "post", "start", "Start one eligible activity", ref("Dataset"), obj(object{"revision": revision, "employee_id": str, "event_id": str}, "revision", "employee_id", "event_id"), "200")
	add("/datasets/{id}/participation/{record}", "put", "finish", "Complete or drop participation; frozen gains are applied only once", ref("Dataset"), obj(object{"revision": revision, "status": object{"type": "string", "enum": []string{"completed", "dropped"}}, "reflection": object{"type": "string", "maxLength": 2000}}, "revision", "status"), "200")
	add("/datasets/{id}/hr", "get", "hr", "Aggregate skill gaps and participation with SQL profile filters", ref("HR"), nil, "200")
	upload := obj(object{"files": array(object{"type": "string", "format": "binary"}), "synthetic": object{"type": "string", "enum": []string{"true"}}, "append_to": str, "revision": object{"type": "string"}}, "files", "synthetic")
	upload["description"] = "5 MiB total multipart limit. One bundle.json, or employees.json + skills.json + events.json + activity_history.csv. Append accepts employees.json and activity_history.csv only, with append_to and current revision."
	add("/import/preview", "post", "preview", "Validate files without writing", ref("Preview"), upload, "200")
	add("/import/commit", "post", "import", "Revalidate and atomically create or append a dataset", ref("Dataset"), upload, "201")
	doc := object{"openapi": "3.1.1", "info": object{"title": "Career Quest local API", "version": "0.2.0", "description": "Single-user loopback service for synthetic data. No authentication. Mutations require the latest revision. A checkbox declaration is not a personal-data detector."}, "servers": []any{object{"url": "http://127.0.0.1:8080/api"}}, "paths": paths, "components": object{"schemas": definitions}}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(doc); err != nil {
		panic(err)
	}
}
