from typing import Callable

import seed.infra.python.seed_flag as seed_flag


class MptInitializer:
    _module_initializers: list[Callable[[], None]] = []
    _done: bool = False

    def register_module_init(self, module_init: Callable[[], None]):
        if self._done:
            raise RuntimeError("Cannot register new init hooks after initialization.")
        return self._module_initializers.append(module_init)

    def initialize(self):
        if self._done:
            raise RuntimeError("Cannot run init hooks after initialization.")
        for hook in self._module_initializers:
            hook()
        self._done = True


_initializer = MptInitializer()


def register_module_init(module_init: Callable[[], None]) -> None:
    _initializer.register_module_init(module_init)


def initialize() -> seed_flag.ConfigStore:
    configs = seed_flag.parse()
    _initializer.initialize()
    return configs
