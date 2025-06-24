import pydantic

import seed.infra.python.seed_typing as seed_typing
import seed.infra.llm.client.python.llm_client as llm_client
import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init

llm_server = seed_flag.define(
    "llm_server", str, "https://ark.cn-beijing.volces.com/api/v3"
)
llm_model = seed_flag.define("llm_model", str, "deepseek-v3-250324")
llm_api_key_file = seed_flag.define(
    "llm_api_key_file", str, "${ND_USER_SECRET_HOME}/volcengine/ARK_API_KEY"
)

_client: llm_client.LlmClient


def init():
    global _client
    _client = llm_client.LlmClient(
        server=llm_server.get(),
        model=llm_model.get(),
        api_key_file_path=llm_api_key_file.get(),
    )


seed_init.register_module_init(init)


@seed_typing.unbind_callable_type(llm_client.LlmClient.request)
async def request(*args, **kwargs):
    return await _client.request(*args, **kwargs)
