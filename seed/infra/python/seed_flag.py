import logging
import types
from typing import Any, Generic, Type, TypeVar

import pydantic_settings

logger = logging.getLogger(__name__)


class ConfigStore:
    _types: dict[str, Any] = {}
    _entries: dict[str, Any] = {}
    _settings: pydantic_settings.BaseSettings | None = None

    def define(self, name: str, config_type: Any, config_value: Any):
        if self._settings is not None:
            raise RuntimeError("Cannot define new config after parsing settings.")
        if name in self._entries:
            raise ValueError(f"Config item '{name}' is already defined.")
        self._types[name] = config_type
        self._entries[name] = config_value

    def parse(self):
        DynamicSettings = types.new_class(
            "DynamicSettings",
            (pydantic_settings.BaseSettings,),
            exec_body=lambda ns: ns.update(
                {
                    "__annotations__": self._types,
                    "__module__": __name__,
                    "__qualname__": "DynamicSettings",
                    "model_config": pydantic_settings.SettingsConfigDict(
                        cli_avoid_json=True,
                        cli_implicit_flags=True,
                        cli_parse_args=True,
                    ),
                    **self._entries,
                },
            ),
        )
        self._settings = DynamicSettings()
        if self._settings:
            logger.info(f"MPT Configs: {self._settings.model_dump_json(indent=2)}")
        return self

    def get(self, name: str) -> Any:
        if self._settings is None:
            raise RuntimeError("Settings have not been parsed yet.")
        return getattr(self._settings, name, None)


_store = ConfigStore()


_T = TypeVar("_T")


class ConfigItemHolder(Generic[_T]):
    _store: ConfigStore
    _name: str

    def __init__(
        self, store: ConfigStore, name: str, config_type: Type, config_value: _T
    ):
        self._store = store
        self._name = name
        self._store.define(name, config_type, config_value)

    def get(self) -> _T:
        return self._store.get(self._name)


def define(name: str, config_type: Type, config_field: _T) -> ConfigItemHolder[_T]:
    return ConfigItemHolder(_store, name, config_type, config_field)


def parse() -> ConfigStore:
    return _store.parse()
