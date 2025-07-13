import base64
import json
import logging
import os

import openai
import openai.types.shared_params.response_format_json_schema as json_schema

import seed.infra.billing.python.llm_bill as llm_bill

logger = logging.getLogger(__name__)


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
        image_png: bytes,
        prompt: str,
        system_prompt: str = "",
        response_schema: json_schema.JSONSchema | None = None,
    ):
        logger.debug(f"prompt: {prompt}")
        image_png_base64 = base64.b64encode(image_png).decode("utf-8")
        image_data_url = f"data:image/png;base64,{image_png_base64}"
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
            bill_summary = self._bill.summary()
            logger.debug(
                f"bill: \x1b[1;36m[in] {bill_summary.total_prompt_tokens} [out] {bill_summary.total_completion_tokens} [cost] {bill_summary.total_cost}\x1b[0m"
            )
        if not response.choices or not response.choices[0].message:
            raise ValueError("No valid response from LLM")
        response_content = response.choices[0].message.content or ""
        response_content = response_content.strip()
        logger.debug(f"response: {response_content}")
        if response_content.startswith("```json") and response_content.endswith("```"):
            response_content = response_content[len("```json") : -len("```")].strip()
        response_json = json.loads(response_content)
        return response_json

    def bill(self):
        return self._bill.summary()
