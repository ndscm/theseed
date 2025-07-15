import json
import logging
import typing

import seed.infra.python.seed_flag as seed_flag

flag_verbose = seed_flag.define_bool("verbose", False)
flag_debug = seed_flag.define_list(
    "debug",
    [
        "cloud",
        "devprod",
        "infra",
        "newtype",
    ],
)


class MptLogFormatter(logging.Formatter):

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


def load():
    debugging_loggers = flag_debug.get()
    if flag_verbose.get():
        if not debugging_loggers:
            logging.getLogger().setLevel(logging.DEBUG)
        logging.getLogger("__main__").setLevel(logging.DEBUG)
        for logger_name in debugging_loggers:
            logging.getLogger(logger_name).setLevel(logging.DEBUG)
