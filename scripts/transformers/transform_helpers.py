from spec_types import OAPISchema
from urllib import request
import json
import yaml


def fetch_and_parse(json_oapi_url: str) -> OAPISchema:
    with request.urlopen(json_oapi_url, timeout=5) as response:
        return json.loads(response.read())


def load_yaml(path: str) -> OAPISchema:
    with open(path, "r") as fd:
        return yaml.load(fd, Loader=yaml.CLoader)


def save_external(spec: OAPISchema, path: str):
    with open(path, "w") as fd:
        yaml.dump(spec, fd, sort_keys=False, indent=4, allow_unicode=True)


def add_spec_uid(spec: OAPISchema, uid: str):
    spec["$id"] = uid
