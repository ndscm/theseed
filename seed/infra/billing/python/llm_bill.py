import logging

import openai.types.completion_usage as completion_usage
import pydantic

logger = logging.getLogger(__name__)


class LlmPrice(pydantic.BaseModel):
    currency: str = "CNY"
    prompt_price: float = 1.0
    completion_price: float = 1.0


class LlmBillSummary(pydantic.BaseModel):
    currency: str = "CNY"
    total_prompt_tokens: int = 0
    total_prompt_cost: float = 0.0
    total_completion_tokens: int = 0
    total_completion_cost: float = 0.0
    total_cost: float = 0.0


class LlmBill:
    title: str
    _price: LlmPrice
    _usages: list[completion_usage.CompletionUsage]

    def __init__(
        self,
        title: str,
        price: LlmPrice,
    ):
        self.title = title
        self._price = price
        self._usages = []

    def append_usage(self, usage: completion_usage.CompletionUsage):
        self._usages.append(usage)
        logger.info(
            f"usage: \x1b[1;36m[in] {usage.prompt_tokens} [out] {usage.completion_tokens}\x1b[0m"
        )

    def summary(self):
        result = LlmBillSummary(
            currency=self._price.currency,
            total_prompt_tokens=0,
            total_completion_tokens=0,
            total_prompt_cost=0.0,
            total_completion_cost=0.0,
            total_cost=0.0,
        )
        for usage in self._usages:
            result.total_prompt_tokens += usage.prompt_tokens
            result.total_completion_tokens += usage.completion_tokens
        result.total_prompt_cost = result.total_prompt_tokens * self._price.prompt_price
        result.total_completion_cost = (
            result.total_completion_tokens * self._price.completion_price
        )
        result.total_cost = result.total_prompt_cost + result.total_completion_cost
        return result
