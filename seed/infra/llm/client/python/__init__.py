import seed.infra.llm.client.python.llm_client as llm_client
import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init
import seed.infra.python.seed_typing as seed_typing

flag_llm_server = seed_flag.define_string(
    "llm_server", "https://ark.cn-beijing.volces.com/api/v3"
)
flag_llm_model = seed_flag.define_string("llm_model", "deepseek-v3-250324")
flag_llm_api_key_file = seed_flag.define_string(
    "llm_api_key_file", "${ND_USER_SECRET_HOME}/volcengine/ARK_API_KEY"
)

_client: llm_client.LlmClient


def init():
    global _client
    _client = llm_client.LlmClient(
        server=flag_llm_server.get(),
        model=flag_llm_model.get(),
        api_key_file_path=flag_llm_api_key_file.get(),
    )


seed_init.register_module_init(init)


@seed_typing.unbind_callable_type(llm_client.LlmClient.request)
async def request(*args, **kwargs):
    return await _client.request(*args, **kwargs)
