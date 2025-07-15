import seed.infra.billing.python.llm_bill as llm_bill
import seed.infra.llm.client.python.llm_client as llm_client
import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init
import seed.infra.python.seed_typing as seed_typing

LlmClient = llm_client.LlmClient

flag_llm_server = seed_flag.define_string(
    "llm_server", "https://ark.cn-beijing.volces.com/api/v3"
)
flag_llm_model = seed_flag.define_string("llm_model", "deepseek-v3-250324")
flag_llm_prompt_price = seed_flag.define_string("llm_prompt_price", "0.000002")
flag_llm_completion_price = seed_flag.define_string(
    "llm_completion_price", "0.000008"
)
flag_llm_api_key_file = seed_flag.define_string(
    "llm_api_key_file", "${ND_USER_SECRET_HOME}/volcengine/ARK_API_KEY"
)

_client: LlmClient


def init():
    global _client
    llm_prompt_price = float(flag_llm_prompt_price.get())
    llm_completion_price = float(flag_llm_completion_price.get())
    _client = LlmClient(
        server=flag_llm_server.get(),
        model=flag_llm_model.get(),
        api_key_file_path=flag_llm_api_key_file.get(),
        price=llm_bill.LlmPrice(
            prompt_price=llm_prompt_price,
            completion_price=llm_completion_price,
        ),
    )


seed_init.register_module_init(init)


@seed_typing.unbind_callable_type(LlmClient.request)
async def request(*args, **kwargs):
    return await _client.request(*args, **kwargs)


@seed_typing.unbind_callable_type(LlmClient.request_expect_pydantic)
async def request_expect_pydantic(*args, **kwargs):
    return await _client.request_expect_pydantic(*args, **kwargs)


@seed_typing.unbind_callable_type(LlmClient.request_expect_list)
async def request_expect_list(*args, **kwargs):
    return await _client.request_expect_list(*args, **kwargs)


@seed_typing.unbind_callable_type(LlmClient.request_expect_pydantic_list)
async def request_expect_pydantic_list(*args, **kwargs):
    return await _client.request_expect_pydantic_list(*args, **kwargs)


def bill():
    return _client.bill()
