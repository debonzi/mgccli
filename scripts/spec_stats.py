import os
from typing import Dict, Any, List, NamedTuple
import argparse
import yaml

OAPISchema = Dict[str, Any]
OAPIStats = Dict[str, List[Any]]


class OAPI(NamedTuple):
    path: str
    name: str
    schema: OAPISchema


class OAPIOperation(NamedTuple):
    path: str
    method: str
    fields: OAPISchema

    def key(self) -> str:
        return self.method.upper() + " " + self.path


class OAPIResource(NamedTuple):
    name: str
    operations: List[OAPIOperation]


class ResponseContext(NamedTuple):
    path: str
    method: str
    code: str


# This is used to fix list indentations, as Pyyaml doesn't indent them :/
class YamlDumper(yaml.Dumper):
    def increase_indent(self, flow=False, indentless=False):
        return super(YamlDumper, self).increase_indent(flow, False)


OPERATION_KEYS = ["get", "put", "post", "delete", "options", "head", "patch", "trace"]


class Filterer:
    filters: List[str]
    filter_out: List[str]

    def should_include(self, key: str) -> bool:
        if self.filters and key not in self.filters:
            return False

        if self.filter_out and key in self.filter_out:
            return False

        return True


filterer = Filterer()


def load_yaml(path: str) -> OAPISchema:
    with open(path, "r") as fd:
        return yaml.load(fd, Loader=yaml.CLoader)


def load_oapis(d: str) -> [OAPI]:
    schemas = []
    for f in os.listdir(d):
        name, ext = os.path.splitext(f)
        if name == "index" or ext != ".yaml":
            print("ignored file:", f)
            continue

        path = os.path.join(d, f)
        schemas.append(OAPI(name=name, path=path, schema=load_yaml(path)))

    return schemas


def fill_responses_stats(op: OAPIOperation, responses: OAPISchema, dst: OAPIStats):
    obj = {op.key(): []}

    for code, r in responses.items():
        content = r.get("content", {})
        if not content:
            # Return for now. In the future check for the code?
            return

        for t, _ in content.items():
            if t != "application/json":
                obj[op.key()].append({code: t})

    if obj[op.key()] and filterer.should_include("non-json-responses"):
        dst.setdefault("non-json-responses", []).append(obj)


def fill_req_body_stats(op: OAPIOperation, r: OAPISchema, dst: OAPIStats):
    content = r.get("content", {})
    if content:
        for t, _ in content.items():
            if t != "application/json" and filterer.should_include("non-json-requests"):
                dst.setdefault("non-json-requests", []).append({op.key(): t})


def fill_operation_stats(op: OAPIOperation, dst: OAPIStats):
    responses = op.fields.get("responses", {})
    if responses:
        fill_responses_stats(op, responses, dst)

    req_body = op.fields.get("requestBody", {})
    if req_body:
        fill_req_body_stats(op, req_body, dst)

    if "operationId" not in op.fields and filterer.should_include(
        "missing_operation_id"
    ):
        dst.setdefault("missing_operation_id", []).append(op.key())

    return


def fill_resource_stats(r: OAPIResource, dst: OAPIStats):
    for op in r.operations:
        fill_operation_stats(op, dst)


def fill_resources(o: OAPI, dst: Dict[str, OAPIResource]) -> List[OAPIOperation]:
    tagless_ops = []
    for pn, p in o.schema.get("paths", {}).items():
        for path_field, sub_fields in p.items():
            if not isinstance(sub_fields, dict) or path_field not in OPERATION_KEYS:
                continue

            op = OAPIOperation(path=pn, method=path_field, fields=sub_fields)
            tags = sub_fields.get("tags")

            if tags:
                res_name = tags[0]
                dst.setdefault(
                    res_name, OAPIResource(name=res_name, operations=[])
                ).operations.append(op)
            else:
                tagless_ops.append(op)

    return tagless_ops


def get_oapi_stats(o: OAPI) -> OAPIStats:
    result = {}
    resources = {}
    tagless_ops = fill_resources(o, resources)

    for res in resources.values():
        fill_resource_stats(res, result)

    for op in tagless_ops:
        fill_operation_stats(op, result)
        if filterer.should_include("tagless_operations"):
            result.setdefault("tagless_operations", []).append(op.key())

    # TODO: Add stats for other fields

    return result


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Validate response and request bodies for all OAPI YAML"
        "files in directory"
    )
    parser.add_argument("dir", type=str, help="Directory of OpenAPI files")
    parser.add_argument(
        "--filter", type=str, action="append", default=[], help="Only show these stats"
    )
    parser.add_argument(
        "--filter-out",
        type=str,
        action="append",
        default=[],
        help="Don't show these stats",
    )

    args = parser.parse_args()

    filterer.filters = args.filter
    filterer.filter_out = args.filter_out

    oapis = load_oapis(args.dir)
    for o in oapis:
        stats = get_oapi_stats(o)
        if stats:
            print(
                yaml.dump(
                    {o.name: stats}, Dumper=YamlDumper, indent=2, allow_unicode=True
                )
            )
