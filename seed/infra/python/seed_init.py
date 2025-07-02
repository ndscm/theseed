import logging
import os
import sys
from typing import Callable

import seed.infra.python.seed_flag as seed_flag

logger = logging.getLogger(__name__)

flag_verbose = seed_flag.define_bool("verbose", False)


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


class MptLoggingFormatter(logging.Formatter):

    COLORS = {
        "DEBUG": "\x1b[1;35m",
        "INFO": "\x1b[1;34m",
        "WARNING": "\x1b[1;33m",
        "ERROR": "\x1b[1;31m",
        "CRITICAL": "\x1b[1;37m\x1b[41m",
    }

    RESET = "\x1b[0m"

    def __init__(self):
        super().__init__(
            style="{",
            fmt="{asctime} {levelname:.1} {name}[{lineno}]\x1b[0m {message}",
            datefmt="%Y-%m-%dT%H:%M:%S",
        )

    def format(self, record):
        message = super().format(record)
        color = self.COLORS.get(record.levelname, self.RESET)
        return f"{color}{message}"


def initialize() -> seed_flag.ConfigStore:
    def _excepthook(exc_type, exc_value, exc_traceback):
        sys.__excepthook__(exc_type, exc_value, exc_traceback)
        logger.critical(exc_value)

    sys.excepthook = _excepthook

    logging.basicConfig(force=True, level=logging.INFO)
    for handler in logging.root.handlers:
        handler.setFormatter(MptLoggingFormatter())
    logger.info(f"MPT Init: {os.path.basename(sys.argv[0])}")
    configs = seed_flag.parse()
    if flag_verbose.get():
        logging.basicConfig(force=True, level=logging.DEBUG)
        for handler in logging.root.handlers:
            handler.setFormatter(MptLoggingFormatter())
    _initializer.initialize()
    return configs
