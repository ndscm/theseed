import os

import seed.infra.python.seed_typing as seed_typing
import seed.infra.llm.client.python.llm_client as llm_client

_client = llm_client.LlmClient(
    server=os.environ.get("LLM_SERVER", "https://ark.cn-beijing.volces.com/api/v3"),
    model=os.environ.get("LLM_MODEL", "deepseek-v3-250324"),
    api_key_file_path=os.environ.get(
        "LLM_API_KEY_FILE", "${ND_USER_SECRET_HOME}/volcengine/ARK_API_KEY"
    ),
)


@seed_typing.unbind_callable_type(llm_client.LlmClient.request)
async def request(*args, **kwargs):
    return await _client.request(*args, **kwargs)
