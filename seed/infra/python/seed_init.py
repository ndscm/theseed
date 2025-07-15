import logging
import os
import sys
from typing import Callable

import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_log as seed_log

logger = logging.getLogger(__name__)


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
    def _excepthook(exc_type, exc_value, exc_traceback):
        sys.__excepthook__(exc_type, exc_value, exc_traceback)
        logger.critical(exc_value)

    sys.excepthook = _excepthook

    logging.basicConfig(force=True, level=logging.INFO)
    for handler in logging.root.handlers:
        handler.setFormatter(seed_log.MptLogFormatter())
    logger.info(f"MPT Init: {os.path.basename(sys.argv[0])}")
    configs = seed_flag.parse()
    seed_log.load()
    _initializer.initialize()
    return configs
