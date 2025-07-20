import ast
import json
import argparse

class DSLParser(ast.NodeVisitor):
    def __init__(self):
        self.api_definitions = {}
        self.servers = {}
        self.sources = []

    def visit_Assign(self, node):
        # Look for assignments like `api = {...}` or `server = Server(...)`
        if isinstance(node.value, ast.Call):
            func_name = self._get_func_name(node.value.func)

            if func_name == "Server":
                self._parse_server(node)
            elif func_name == "Source":
                self._parse_source(node)
        elif isinstance(node.value, ast.Dict):
            # Handling `api = { "insert": Work(...) }`
            self._parse_api_dict(node.value)
        self.generic_visit(node)

    def _parse_source(self, node):
        call = node.value

        # Defaults
        name = api = None
        arrival_rate = timeout = retries = None

        # Handle positional args
        args = call.args
        if len(args) >= 1:
            name = self._parse_value(args[0])
        if len(args) >= 2:
            api = self._parse_value(args[1])
        if len(args) >= 3:
            arrival_rate = self._parse_value(args[2])
        if len(args) >= 4:
            timeout = self._parse_value(args[3])
        if len(args) >= 5:
            retries = self._parse_value(args[4])

        # Override with keyword args if present
        for kw in call.keywords:
            val = self._parse_value(kw.value)
            if kw.arg == "name":
                name = val
            elif kw.arg == "api":
                api = val
            elif kw.arg == "arrival_rate":
                arrival_rate = val
            elif kw.arg == "timeout":
                timeout = val
            elif kw.arg == "retries":
                retries = val

        source = {
            "name": name,
            "api": api,
            "arrival_rate": arrival_rate,
            "timeout": timeout,
            "retries": retries
        }

        self.sources.append(source)

    def _parse_value(self, node):
        if isinstance(node, ast.Constant):# Python 3.8+
            return node.value
        elif isinstance(node, ast.Name):
            return node.id  # e.g. variable name
        elif isinstance(node, ast.Attribute):
            return f"{self._parse_value(node.value)}.{node.attr}"  # e.g. Constants.CLOSED
        elif isinstance(node, ast.Num):  # Python <3.8
            return node.n
        elif isinstance(node, ast.Str):  # Python <3.8
            return node.s
        return None

    def _parse_dependent_call(self, call_node):
        try:
            args = call_node.args
            return {
                "source": args[1].s,
                "target": args[0].s,
                "api": args[2].s,
                "blocking": True if (self._parse_value(args[3]) == "Constants.CLOSED") else False,
                "timeout": self._parse_value(args[4]),
                "retry": self._parse_value(args[5])
            }
        except (IndexError, AttributeError):
            return None

    def _parse_api_dict(self, dict_node):
        self.all_api_defs = {}  # Reset on every redefinition

        for key, val in zip(dict_node.keys, dict_node.values):
            if isinstance(val, ast.Call) and self._get_func_name(val.func) == "Work":
                api_name = key.s
                processing_rate = val.args[0].n
                downstream_calls = []

                if len(val.args) > 1 and isinstance(val.args[1], ast.List):
                    for elem in val.args[1].elts:
                        if isinstance(elem, ast.Call) and self._get_func_name(elem.func) == "DependentCall":
                            call_info = self._parse_dependent_call(elem)
                            if call_info:
                                downstream_calls.append(call_info)

                self.all_api_defs[api_name] = {
                    "processing_rate": processing_rate,
                    "all_downstream_calls": downstream_calls
                }

    def _get_arg_by_index(self, args, index):
        if index < len(args):
            return self._parse_value(args[index])
        return None

    def _parse_server(self, node):
        call = node.value
        args = call.args
        server_name = args[0].s

        api_var = args[1].id if isinstance(args[1], ast.Name) else None
        apis = {}

        if api_var:
            for api_name, api_info in self.all_api_defs.items():
                filtered_calls = [
                    call for call in api_info["all_downstream_calls"]
                    if call["source"] == server_name
                ]
                apis[api_name] = {
                    "processing_rate": api_info["processing_rate"],
                    "downstream_services": filtered_calls
                }

        server_obj = {
            "name": server_name,
            "qsize": self._get_kwarg(call, "qsize") or self._get_arg_by_index(args, 2),
            "threadpool": self._get_kwarg(call, "thread_pool") or self._get_arg_by_index(args, 4),
            "apis": apis
        }

        self.servers[server_name] = server_obj

    def _get_func_name(self, func):
        if isinstance(func, ast.Name):
            return func.id
        elif isinstance(func, ast.Attribute):
            return func.attr
        return None

    def _get_kwarg(self, call_node, kwarg_name):
        for kw in call_node.keywords:
            if kw.arg == kwarg_name:
                return kw.value.n
        return None

    def get_json(self):
        return json.dumps({
            "servers": list(self.servers.values()),
            "sources": self.sources
        }, indent=4)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Parse DSL and output JSON config")
    parser.add_argument("input_file", help="Path to the DSL Python file")
    parser.add_argument("-o", "--output", default="../config.json", help="Output JSON file path (default: ../config.json)")
    args = parser.parse_args()

    with open(args.input_file, "r") as f:
        tree = ast.parse(f.read())

    dsl_parser = DSLParser()
    dsl_parser.visit(tree)

    with open(args.output, "w") as f:
        f.write(dsl_parser.get_json())


