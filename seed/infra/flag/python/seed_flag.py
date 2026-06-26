import abc
import argparse
import json
import logging
import os
import typing

logger = logging.getLogger(__name__)


class FlagHolder(abc.ABC):
    """Base class for configuration items."""

    _parsed: bool

    def __init__(self):
        super().__init__()
        self._parsed = False

    @abc.abstractmethod
    def get(self) -> typing.Any:
        pass

    @abc.abstractmethod
    def set(self, value: typing.Any) -> None:
        pass

    def finalize(self) -> None:
        self._parsed = True

    def check(self) -> None:
        if not self._parsed:
            raise RuntimeError("config is used before being parsed")


_bool_map = {
    "true": True,
    "t": True,
    "1": True,
    "yes": True,
    "on": True,
    "false": False,
    "f": False,
    "0": False,
    "no": False,
    "off": False,
}


class BoolFlagHolder(FlagHolder):
    value: bool
    default: bool

    def __init__(self, default: bool = False):
        super().__init__()
        self.default = default
        self.value = default

    def get(self) -> bool:
        super().check()
        return self.value

    def set(self, value) -> None:
        if isinstance(value, bool):
            self.value = value
            return
        if isinstance(value, str):
            self.value = _bool_map.get(value.lower(), self.default)
            return
        raise ValueError(f"expected a boolean value")


class StringFlagHolder(FlagHolder):
    value: str

    def __init__(self, default: str = ""):
        super().__init__()
        self.value = default

    def get(self) -> str:
        super().check()
        return self.value

    def set(self, value) -> None:
        if isinstance(value, str):
            self.value = value
            return
        raise ValueError(f"expected a string value")


class StringListFlagHolder(FlagHolder):
    value: list[str]

    def __init__(self, default: list[str] | None = None):
        super().__init__()
        self.value = default or []

    def get(self) -> list[str]:
        super().check()
        return self.value

    def set(self, value) -> None:
        if isinstance(value, list):
            self.value = [v for v in value if isinstance(v, str)]
            return
        if isinstance(value, str):
            self.value = value.split(",")
            return
        raise ValueError(f"expected a list of strings")


class FlagStore:
    _parser: argparse.ArgumentParser
    _configs: dict[str, FlagHolder]

    def __init__(self):
        self._parser = argparse.ArgumentParser()
        self._configs = {}

    def define(
        self,
        name: str,
        holder: FlagHolder,
        *args,
        default: typing.Any = None,
        **kwargs,
    ):
        self._configs[name] = holder
        if default is not None:
            holder.set(default)
        self._parser.add_argument(*args, **kwargs)

    def parse(self):
        for k, v in os.environ.items():
            k = k.lower()
            if k in self._configs:
                self._configs[k].set(v)
        parsed_args = self._parser.parse_args()
        for k, v in vars(parsed_args).items():
            if k in self._configs and v is not None:
                self._configs[k].set(v)
        for holder in self._configs.values():
            holder.finalize()
        simplified = {k: v.get() for k, v in self._configs.items()}
        logger.info(f"flags: {json.dumps(simplified, indent=2, ensure_ascii=False)}")
        return self


_store = FlagStore()


def define_bool(name: str, default: bool = False, help: str = "") -> BoolFlagHolder:
    holder = BoolFlagHolder(
        default=default,
    )
    _store.define(
        name,
        holder,
        f"--{name}",
        action=argparse.BooleanOptionalAction,
        default=default,
        help=help,
    )
    return holder


def define_string(name: str, default: str = "", help: str = "") -> StringFlagHolder:
    holder = StringFlagHolder(
        default=default,
    )
    _store.define(
        name,
        holder,
        f"--{name}",
        default=default,
        help=help,
    )
    return holder


def define_list(
    name: str, default: list[str] | None = None, help: str = ""
) -> StringListFlagHolder:
    holder = StringListFlagHolder(
        default=default,
    )
    _store.define(
        name,
        holder,
        f"--{name}",
        action="append",
        default=default or [],
        help=help,
    )
    return holder


def define_positional(name: str, default: str = "", help: str = "") -> StringFlagHolder:
    holder = StringFlagHolder(
        default=default,
    )
    _store.define(
        name,
        holder,
        f"{name}",
        default=default,
        help=help,
        nargs="?",
    )
    return holder


def define_positional_list(
    name: str, default: list[str] | None = None, help: str = ""
) -> StringListFlagHolder:
    holder = StringListFlagHolder(
        default=default,
    )
    _store.define(
        name,
        holder,
        f"{name}",
        default=default or [],
        help=help,
        nargs="*",
    )
    return holder


def parse() -> FlagStore:
    return _store.parse()
