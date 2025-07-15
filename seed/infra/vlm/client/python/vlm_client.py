import base64
import json
import logging
import mimetypes
import os
import typing

import openai
import openai.types.shared_params.response_format_json_schema as json_schema
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


class VlmClient:
    sever: str
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
                "VLM_SERVER", "https://ark.cn-beijing.volces.com/api/v3"
            )
        self.model = model
        if not self.model:
            self.model = os.environ.get("VLM_MODEL", "doubao-1.5-vision-pro-250328")
        self.api_key = api_key
        if not self.api_key:
            self.api_key = os.environ.get("VLM_API_KEY", "")
        if not self.api_key:
            if not api_key_file_path:
                api_key_file_path = os.environ.get(
                    "VLM_API_KEY_FILE", "${ND_USER_SECRET_HOME}/volcengine/ARK_API_KEY"
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
        image_format: str = "",
        image_bytes: bytes = b"",
        image_path: str = "",
        response_schema: json_schema.JSONSchema | None = None,
        task: str = "",
    ) -> str:
        logger.debug(
            "%s system prompt:\n%s\nprompt:\n%s", task or "vlm", system_prompt, prompt
        )

        image_mime_type, _ = mimetypes.guess_type(f"image.{image_format}")
        if image_path and image_bytes:
            raise ValueError("Cannot provide both image_path and image_bytes")
        if not image_path and not image_bytes:
            raise ValueError("Must provide either image_path or image_bytes")
        if image_path:
            if not os.path.exists(image_path):
                raise FileNotFoundError(f"image file not found: {image_path}")
            with open(image_path, "rb") as f:
                image_bytes = f.read()
            path_mime_type, _ = mimetypes.guess_type(image_path)
            if image_mime_type and image_mime_type != path_mime_type:
                raise ValueError(f"image format {image_format} mismatch: {image_path}")
            image_mime_type = path_mime_type
        if image_mime_type not in [
            "image/bmp",
            "image/gif",
            "image/jpeg",
            "image/png",
            "image/webp",
            "image/x-icon",
        ]:
            raise ValueError(f"unsupported image format: {image_mime_type}")

        image_base64 = base64.b64encode(image_bytes).decode("utf-8")
        image_data_url = f"data:{image_mime_type};base64,{image_base64}"
        extra_kwargs = {}
        if response_schema:
            response_format = json_schema.ResponseFormatJSONSchema(
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
                        {
                            "type": "image_url",
                            "image_url": {"url": image_data_url},
                        },
                        {
                            "type": "text",
                            "text": prompt,
                        },
                    ],
                },
            ],
            temperature=0,
            stream=False,
            timeout=10,
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
        logger.debug("%s response:\n%s", task or "vlm", response_content)
        return response_content

    async def request_expect_pydantic(
        self,
        prompt: str,
        *,
        system_prompt: str = "",
        image_format: str = "",
        image_bytes: bytes = b"",
        image_path: str = "",
        response_pydantic_type: typing.Type[PydanticModel],
        task: str = "",
    ) -> tuple[PydanticModel, str]:
        response_content: str = await self.request(
            prompt,
            system_prompt=system_prompt,
            image_format=image_format,
            image_bytes=image_bytes,
            image_path=image_path,
            task=task,
        )
        response_json = _parse_json_response(response_content)
        response_pydantic = response_pydantic_type(**response_json)
        return response_pydantic, response_content

    async def request_expect_list(
        self,
        prompt: str,
        *,
        system_prompt: str = "",
        image_format: str = "",
        image_bytes: bytes = b"",
        image_path: str = "",
        task: str = "",
    ) -> tuple[list, str]:
        response_content: str = await self.request(
            prompt,
            system_prompt=system_prompt,
            image_format=image_format,
            image_bytes=image_bytes,
            image_path=image_path,
            task=task,
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
        image_format: str = "",
        image_bytes: bytes = b"",
        image_path: str = "",
        response_pydantic_item_type: typing.Type[PydanticModel],
        task: str = "",
    ) -> tuple[list[PydanticModel], str]:
        response_content: str = await self.request(
            prompt,
            system_prompt=system_prompt,
            image_format=image_format,
            image_bytes=image_bytes,
            image_path=image_path,
            task=task,
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
