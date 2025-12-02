import asyncio

import seed.infra.llm.client.python as llm_client
import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init

flag_prompt = seed_flag.define_string("prompt", "")
arg_prompt_text = seed_flag.define_positional("prompt_text", "")


async def main():
    seed_init.initialize()
    prompt = ""
    if flag_prompt.get():
        with open(flag_prompt.get(), "r", encoding="utf-8") as f:
            prompt = f.read()
    if arg_prompt_text.get():
        if prompt:
            raise ValueError("Only one of --prompt and prompt_text can be specified")
        prompt = arg_prompt_text.get()
    if not prompt:
        raise ValueError("Either --prompt or prompt_text must be specified")
    response = await llm_client.request(prompt=prompt)
    print(response)


if __name__ == "__main__":
    asyncio.run(main())
