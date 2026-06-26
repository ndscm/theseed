from typing import Any, Callable, Concatenate, ParamSpec, TypeVar

T = TypeVar("T")
P = ParamSpec("P")


def unbind_callable_type(
    _: Callable[Concatenate[Any, P], T],
) -> Callable[[Callable], Callable[P, T]]:
    """A decorator that allows a function to pass through its typing hints"""

    def decorator(inner: Callable) -> Callable[P, T]:
        def typed_inner(*args: P.args, **kwargs: P.kwargs) -> T:
            return inner(*args, **kwargs)

        return typed_inner

    return decorator


def binded_callable_type(
    _: Callable[Concatenate[Any, P], T],
) -> Callable[[Callable], Callable[Concatenate[Any, P], T]]:
    """A decorator that allows a function to pass through its typing hints"""

    def decorator(inner: Callable) -> Callable[Concatenate[Any, P], T]:
        def typed_inner(self, *args: P.args, **kwargs: P.kwargs) -> T:
            return inner(self, *args, **kwargs)

        return typed_inner

    return decorator
