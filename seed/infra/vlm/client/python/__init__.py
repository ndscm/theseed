import seed.infra.billing.python.llm_bill as llm_bill
import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init
import seed.infra.python.seed_typing as seed_typing
import seed.infra.vlm.client.python.vlm_client as vlm_client

VlmClient = vlm_client.VlmClient

flag_vlm_server = seed_flag.define_string(
    "vlm_server", "https://ark.cn-beijing.volces.com/api/v3"
)
flag_vlm_model = seed_flag.define_string("vlm_model", "doubao-1.5-vision-pro-250328")
flag_vlm_prompt_price = seed_flag.define_string("vlm_prompt_price", "0.000003")
flag_vlm_completion_price = seed_flag.define_string(
    "vlm_completion_price", "0.000009"
)
flag_vlm_api_key_file = seed_flag.define_string(
    "vlm_api_key_file", "${ND_USER_SECRET_HOME}/volcengine/ARK_API_KEY"
)

_client: VlmClient


def init():
    global _client
    vlm_prompt_price = float(flag_vlm_prompt_price.get())
    vlm_completion_price = float(flag_vlm_completion_price.get())
    _client = VlmClient(
        server=flag_vlm_server.get(),
        model=flag_vlm_model.get(),
        api_key_file_path=flag_vlm_api_key_file.get(),
        price=llm_bill.LlmPrice(
            prompt_price=vlm_prompt_price,
            completion_price=vlm_completion_price,
        ),
    )


seed_init.register_module_init(init)


@seed_typing.unbind_callable_type(VlmClient.request)
async def request(*args, **kwargs):
    return await _client.request(*args, **kwargs)


def bill():
    return _client.bill()
