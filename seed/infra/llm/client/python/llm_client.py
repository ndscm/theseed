import json
import os
from typing import Any

import openai
import openai.types.shared_params.response_format_json_schema as response_format_json_schema


class LlmClient:
    sever: str
    model: str
    api_key: str
    _client: openai.AsyncOpenAI

    total_prompt_tokens = 0
    total_completion_tokens = 0

    def __init__(
        self,
        server: str = "",
        model: str = "",
        api_key: str = "",
        api_key_file_path: str = "",
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

    async def request(
        self,
        prompt: str,
        system_prompt: str = "",
        response_schema: response_format_json_schema.JSONSchema | None = None,
    ) -> tuple[Any, str]:
        print(f"[llm] prompt: {prompt}")
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
            prompt_tokens = response.usage.prompt_tokens
            completion_tokens = response.usage.completion_tokens
            self.total_prompt_tokens += prompt_tokens
            self.total_completion_tokens += completion_tokens
            print(
                f"[llm] usage: \x1b[1;36m[in] {prompt_tokens} [out] {completion_tokens}\x1b[0m"
            )
            print(
                f"[llm] total: \x1b[1;36m[in] {self.total_prompt_tokens} [out] {self.total_completion_tokens}\x1b[0m"
            )
        if not response.choices or not response.choices[0].message:
            raise ValueError("No valid response from LLM")
        response_content = response.choices[0].message.content or ""
        response_content = response_content.strip()
        print(f"[llm] response: {response_content}")
        if response_content.startswith("```json") and response_content.endswith("```"):
            response_content = response_content[len("```json") : -len("```")].strip()
        response_json = json.loads(response_content)
        return response_json, response_content
