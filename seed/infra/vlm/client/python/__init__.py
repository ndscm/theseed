import os

import seed.infra.python.seed_typing as seed_typing
import seed.infra.vlm.client.python.vlm_client as vlm_client

_client = vlm_client.VlmClient(
    server=os.environ.get("VLM_SERVER", "https://ark.cn-beijing.volces.com/api/v3"),
    model=os.environ.get("VLM_MODEL", "doubao-1.5-vision-pro-250328"),
    api_key_file_path=os.environ.get(
        "VLM_API_KEY_FILE", "${ND_USER_SECRET_HOME}/volcengine/ARK_API_KEY"
    ),
)


@seed_typing.unbind_callable_type(vlm_client.VlmClient.request)
async def request(*args, **kwargs):
    return await _client.request(*args, **kwargs)
