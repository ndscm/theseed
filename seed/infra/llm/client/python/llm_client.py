import json
import logging
import os
import typing

import openai
import openai.types.shared_params.response_format_json_schema as response_format_json_schema
import pydantic

import seed.infra.billing.python.llm_bill as llm_bill
import seed.infra.python.seed_log as seed_log

logger = logging.getLogger(__name__)


def _parse_json_response(response_content: str):
    if response_content.startswith("```json") and response_content.endswith("```"):
        response_content = response_content[len("```json") : -len("```")].strip()
    response_json = json.loads(response_content)
    if "error" in response_json:
        raise RuntimeError(response_json["error"])
    return response_json


PydanticModel = typing.TypeVar("PydanticModel", bound=pydantic.BaseModel)


class LlmClient:
    server: str
    model: str
    api_key: str
    _client: openai.AsyncOpenAI

    _bill: llm_bill.LlmBill

    def __init__(
        self,
        server: str = "",
        model: str = "",
        api_key: str = "",
        api_key_file_path: str = "",
        price: llm_bill.LlmPrice | None = None,
    ):
        self.server = server
        if not self.server:
            self.server = os.environ.get(
                "LLM_SERVER", "https://ark.cn-beijing.volces.com/api/v3"
            )
        self.model = model
        if not self.model:
            self.model = os.environ.get("LLM_MODEL", "deepseek-v3-250324")
        self.api_key = api_key
        if not self.api_key:
            self.api_key = os.environ.get("LLM_API_KEY", "")
        if not self.api_key:
            if not api_key_file_path:
                api_key_file_path = os.environ.get(
                    "LLM_API_KEY_FILE", "${ND_USER_SECRET_HOME}/volcengine/ARK_API_KEY"
                )
            api_key_file_path = os.path.expanduser(api_key_file_path)
            with open(api_key_file_path, "r", encoding="utf-8") as f:
                self.api_key = f.read().strip()
        self._client = openai.AsyncOpenAI(
            api_key=self.api_key,
            base_url=self.server,
        )
        self._bill = llm_bill.LlmBill(
            title=f"{self.model} ({self.server})",
            price=price or llm_bill.LlmPrice(),
        )

    async def request(
        self,
        prompt: str,
        *,
        system_prompt: str = "",
        response_schema: response_format_json_schema.JSONSchema | None = None,
        task: str = "",
    ) -> str:
        logger.debug(
            "%s system prompt:\n%s\nprompt:\n%s", task or "llm", system_prompt, prompt
        )
        extra_kwargs = {}
        if response_schema:
            response_format = response_format_json_schema.ResponseFormatJSONSchema(
                type="json_schema",
                json_schema=response_schema,
            )
            extra_kwargs["response_format"] = response_format
        response = await self._client.chat.completions.create(
            model=self.model,
            messages=[
                {
                    "role": "system",
                    "content": [
                        {"type": "text", "text": system_prompt},
                    ],
                },
                {
                    "role": "user",
                    "content": [
                        {"type": "text", "text": prompt},
                    ],
                },
            ],
            temperature=0,
            stream=False,
            timeout=30,
            **extra_kwargs,
        )
        if response.usage:
            self._bill.append_usage(response.usage)
            logger.debug(
                "bill: \x1b[1;36m%s\x1b[0m",
                seed_log.Lazy(
                    lambda: (
                        s := self._bill.summary(),
                        f"[in] {s.total_prompt_tokens} [out] {s.total_completion_tokens} [cost] {s.total_cost}",
                    )[-1]
                ),
            )
        if not response.choices or not response.choices[0].message:
            raise ValueError("No valid response from LLM")
        response_content = response.choices[0].message.content or ""
        response_content = response_content.strip()
        logger.debug("%s response:\n%s", task or "llm", response_content)
        return response_content

    async def request_expect_pydantic(
        self,
        prompt: str,
        *,
        system_prompt: str = "",
        response_pydantic_type: typing.Type[PydanticModel],
        task: str = "",
    ) -> tuple[PydanticModel, str]:
        response_content: str = await self.request(
            prompt, system_prompt=system_prompt, task=task
        )
        response_json = _parse_json_response(response_content)
        response_pydantic = response_pydantic_type(**response_json)
        return response_pydantic, response_content

    async def request_expect_list(
        self,
        prompt: str,
        *,
        system_prompt: str = "",
        task: str = "",
    ) -> tuple[list, str]:
        response_content: str = await self.request(
            prompt, system_prompt=system_prompt, task=task
        )
        response_json = _parse_json_response(response_content)
        if not response_json or not isinstance(response_json, list):
            raise ValueError(f"invalid list: {response_content}")
        return response_json, response_content

    async def request_expect_pydantic_list(
        self,
        prompt: str,
        *,
        system_prompt: str = "",
        response_pydantic_item_type: typing.Type[PydanticModel],
        task: str = "",
    ) -> tuple[list[PydanticModel], str]:
        response_content: str = await self.request(
            prompt, system_prompt=system_prompt, task=task
        )
        response_json = _parse_json_response(response_content)
        if not response_json or not isinstance(response_json, list):
            raise ValueError(f"invalid list: {response_content}")
        response_pydantic = [
            response_pydantic_item_type(**item) for item in response_json
        ]
        return response_pydantic, response_content

    def bill(self):
        return self._bill.summary()
